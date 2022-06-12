package match

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sirupsen/logrus"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/sets"
	"github.com/vasar-network/vap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// providers maps between a *player.Player and a Provider.
var providers sync.Map

// LookupProvider looks up the *Provider of the *player.Player passed.
func LookupProvider(p *player.Player) (Provider, bool) {
	h, ok := providers.Load(p)
	if ok {
		return h.(Provider), ok
	}
	return nil, false
}

var (
	ranked   *RankedProvider
	unranked *UnrankedProvider
)

// Ranked ...
func Ranked() *RankedProvider {
	return ranked
}

// Unranked ...
func Unranked() *UnrankedProvider {
	return unranked
}

// Provider ...
type Provider interface {
	QueuedUsers(g game.Game) int
	PlayingUsers(g game.Game) int
	EnterQueue(g game.Game, p *player.Player)
	Queued(g game.Game, p *player.Player) bool
	ExitQueue(p *player.Player)

	StartMatch(g game.Game, v *vap.Vap, players sets.Set[*user.User], updateStats bool)
	Match(p *player.Player) (*Match, bool)
	RunningMatches() []*Match
	StopMatch(g game.Game, match *Match)

	Game(pl *player.Player) (game.Game, bool)
}

// provider ...
type provider struct {
	p Provider

	log *logrus.Logger

	w *world.World

	mu sync.Mutex
	g  *Grid

	duels bool

	queue     map[game.Game][]*player.Player
	queueChan map[game.Game]map[*player.Player]chan struct{}
	matches   map[game.Game]sets.Set[*Match]
}

// newProvider ...
func newProvider(log *logrus.Logger, p Provider, duels bool) *provider {
	d := &provider{
		p: p,

		log: log,

		duels: duels,

		w: world.Config{Log: log, Dim: world.Overworld}.New(),
		g: NewGrid(),

		queue:     make(map[game.Game][]*player.Player),
		queueChan: make(map[game.Game]map[*player.Player]chan struct{}),
		matches:   make(map[game.Game]sets.Set[*Match]),
	}
	d.w.SetDefaultGameMode(world.GameModeSurvival)
	d.w.SetDifficulty(world.DifficultyNormal)
	d.w.StopWeatherCycle()
	d.w.SetDefaultGameMode(world.GameModeAdventure)
	d.w.SetTime(6000)
	d.w.StopTime()
	d.w.SetTickRange(0)
	for _, g := range game.Games() {
		d.matches[g] = make(sets.Set[*Match])
		d.queueChan[g] = make(map[*player.Player]chan struct{})
	}
	return d
}

// World ...
func (m *provider) World() *world.World {
	return m.w
}

// QueuedUsers ...
func (m *provider) QueuedUsers(g game.Game) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queue[g])
}

// PlayingUsers ...
func (m *provider) PlayingUsers(g game.Game) (playing int) {
	m.mu.Lock()
	matches := maps.Clone(m.matches[g])
	m.mu.Unlock()
	for match := range matches {
		playing += match.TotalAlive()
	}
	return
}

// Queued ...
func (m *provider) Queued(g game.Game, p *player.Player) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Contains(m.queue[g], p)
}

// ExitQueue ...
func (m *provider) ExitQueue(p *player.Player) {
	g, ok := m.Game(p)
	if !ok {
		return
	}

	m.mu.Lock()
	ind := slices.Index(m.queue[g], p)
	if ind == -1 {
		m.mu.Unlock()
		return
	}
	providers.Delete(p)
	m.queue[g] = slices.Delete(m.queue[g], ind, ind+1)
	c := m.queueChan[g][p]
	close(c)
	delete(m.queueChan[g], p)
	m.mu.Unlock()

	p.Message(lang.Translatef(p.Locale(), "queue.message.leave"))
	kit.Apply(kit.Lobby{}, p)
	if u, ok := user.Lookup(p); ok {
		u.SetBoard(lobby.Lobby())
	}
}

// StartMatch ...
func (m *provider) StartMatch(g game.Game, v *vap.Vap, players sets.Set[*user.User], updateStats bool) {
	m.mu.Lock()

	pos := m.g.Next()
	m.g.Close(pos)

	m.mu.Unlock()

	match, err := NewMatch(g, m.p, pos, v, m.w, players, m.duels, updateStats)
	if err != nil {
		m.log.Errorf("failed to create match: %v", err)
		return
	}
	m.mu.Lock()
	m.matches[g].Add(match)
	m.mu.Unlock()
}

// RunningMatches ...
func (m *provider) RunningMatches() (matches []*Match) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ma := range m.matches {
		for m := range ma {
			matches = append(matches, m)
		}
	}
	return
}

// Match ...
func (m *provider) Match(p *player.Player) (*Match, bool) {
	u, ok := user.Lookup(p)
	if !ok {
		// User is not logged in.
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, matches := range m.matches {
		for match := range matches {
			if match.players.Contains(u) {
				return match, true
			}
		}
	}
	return nil, false
}

// StopMatch ...
func (m *provider) StopMatch(g game.Game, match *Match) {
	match.Close()

	m.mu.Lock()
	m.matches[g].Delete(match)
	m.g.Open(match.gridPos)
	m.mu.Unlock()
}

// Game ...
func (m *provider) Game(p *player.Player) (game.Game, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for g, users := range m.queue {
		if slices.Contains(users, p) {
			return g, true
		}
	}
	if u, ok := user.Lookup(p); ok {
		for _, matches := range m.matches {
			for match := range matches {
				if match.players.Contains(u) {
					return match.Game(), true
				}
			}
		}
	}
	return game.Game{}, false
}

// arenas maps between a game and the arenas for the game.
var arenas = make(map[game.Game][]*vap.Vap)

// init registers all arenas and Gob types.
func init() {
	var mu sync.Mutex
	a := make(map[game.Game]map[string]*vap.Vap)
	for _, g := range game.Games() {
		a[g] = make(map[string]*vap.Vap)
	}

	var w sync.WaitGroup
	err := filepath.Walk("assets/arenas/duels", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			w.Add(1)
			go func() {
				v, err := vap.ReadFile(path)
				if err != nil {
					panic(err)
				}
				name, games, _ := v.Arena()

				mu.Lock()
				defer mu.Unlock()
				for _, g := range games {
					a[game.ByID(g)][name] = v
				}
				w.Done()
			}()
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	w.Wait()

	for g, m := range a {
		k := maps.Keys(m)
		slices.Sort(k)
		for _, n := range k {
			arenas[g] = append(arenas[g], m[n])
		}
	}
}

// Arenas ...
func Arenas(g game.Game) []*vap.Vap {
	return arenas[g]
}

// RandomArena ...
func RandomArena(g game.Game) *vap.Vap {
	vaps := arenas[g]
	return vaps[rand.Intn(len(vaps))]
}

// queueMessage returns the appropriate queue message for the options provided.
func queueMessage(locale language.Tag, game game.Game, variant string, opts map[string]any) string {
	b := &strings.Builder{}
	b.WriteString("\n")
	b.WriteString(lang.Translatef(locale, "queue.message.title", variant, game.Name()))
	b.WriteString("\n")

	keys := maps.Keys(opts)
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteString(lang.Translatef(locale, "queue.message.option", key, opts[key]))
		b.WriteString("\n")
	}

	b.WriteString(lang.Translatef(locale, "queue.message.searching"))
	b.WriteString("\n\n\n")
	return b.String()
}

// canMatch ...
func canMatch(u *user.User, o *user.User) bool {
	if u == o {
		// You can't match with yourself.
		return false
	}
	if recentOpponent, ok := u.RecentOpponent(); ok && recentOpponent == o {
		// Player one already matched with player two recently.
		return false
	}
	if recentOpponent, ok := o.RecentOpponent(); ok && recentOpponent == u {
		// Player two already matched with player one recently.
		return false
	}
	return u.DesiredPingRange().Compare(int(o.Latency())) &&
		o.DesiredPingRange().Compare(int(u.Latency())) &&
		slices.Contains(u.DesiredDeviceGroups(), o.DeviceGroup()) &&
		slices.Contains(o.DesiredDeviceGroups(), u.DeviceGroup())
}
