package ffa

import (
	"github.com/df-mc/atomic"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/sets"
)

// providers maps between a *player.Player and a *Provider.
var providers sync.Map

// LookupProvider looks up the *Provider of the *player.Player passed.
func LookupProvider(p *player.Player) (*Provider, bool) {
	h, ok := providers.Load(p)
	if ok {
		return h.(*Provider), ok
	}
	return nil, false
}

// ffas ...
var ffas []*Provider

// Providers returns all the ffa providers.
func Providers() []*Provider {
	return ffas
}

// Provider is a simple FFA provider for any FFA game.
type Provider struct {
	w    *world.World
	game game.Game

	playerMu sync.Mutex
	players  sets.Set[*player.Player]

	open atomic.Bool
	pvp  atomic.Bool
}

// NewProvider ...
func NewProvider(game game.Game, w *world.World) *Provider {
	p := &Provider{
		players: make(sets.Set[*player.Player]),
		open:    *atomic.NewBool(true),
		pvp:     *atomic.NewBool(true),
		game:    game,
		w:       w,
	}
	ffas = append(ffas, p)
	return p
}

// Game ...
func (s *Provider) Game() game.Game {
	return s.game
}

// Players ...
func (s *Provider) Players() []*player.Player {
	s.playerMu.Lock()
	defer s.playerMu.Unlock()
	return s.players.Values()
}

// PlayerCount ...
func (s *Provider) PlayerCount() int {
	s.playerMu.Lock()
	defer s.playerMu.Unlock()
	return len(s.players)
}

// AddPlayer ...
func (s *Provider) AddPlayer(p *player.Player) {
	providers.Store(p, s)

	s.playerMu.Lock()
	s.players.Add(p)
	s.playerMu.Unlock()

	lobby.Lobby().RemovePlayer(p, false)

	s.w.AddEntity(p)
	p.SetAttackImmunity(time.Second * 3)
	p.Teleport(s.w.Spawn().Vec3Middle())

	var fall float64
	switch s.Game() {
	case game.NoDebuff():
		fall = 1.5
	case game.Sumo():
		fall = 0.5
	}
	time.AfterFunc(time.Millisecond*300, func() {
		p.SetVelocity(mgl64.Vec3{0, -fall, 0})
	})

	kit.Apply(s.game.Kit(true), p)
	if u, ok := user.Lookup(p); ok {
		u.SetBoard(s)
	}
}

// RemovePlayer ...
func (s *Provider) RemovePlayer(p *player.Player, force bool) {
	providers.Delete(p)

	s.playerMu.Lock()
	s.players.Delete(p)
	s.playerMu.Unlock()

	if !force {
		lobby.Lobby().AddPlayer(p)
	}
}

// SendScoreboard ...
func (s *Provider) SendScoreboard(p *player.Player) {
	if u, ok := user.Lookup(p); ok {
		stats := u.Stats()
		board.Send(p, "scoreboard.ffa", stats.Kills, stats.KillStreak, stats.Deaths)
	}
}

// Open checks if the arena is open.
func (s *Provider) Open() bool {
	return s.open.Load()
}

// PvP checks if PvP is enabled for the provider.
func (s *Provider) PvP() bool {
	return s.pvp.Load()
}

// ToggleStatus will toggle the open/closed status of the arena.
func (s *Provider) ToggleStatus() (old bool) {
	return s.open.Toggle()
}

// TogglePvP toggles PvP for the provider and returns the old value.
func (s *Provider) TogglePvP() bool {
	return s.pvp.Toggle()
}
