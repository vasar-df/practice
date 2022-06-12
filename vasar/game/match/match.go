package match

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/hako/durafmt"
	hook "github.com/justtaldevelops/webhook"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/sets"
	"github.com/vasar-network/vails/webhook"
	"github.com/vasar-network/vap"
	"golang.org/x/exp/maps"
	"math"
	"strings"
	"sync"
	"time"
)

// Match represents a private arena where players battle each other. Matches are generated under the provider's virtual
// world, to prevent potential slowdowns in a match-per-world scenario due to the amount of worlds ticking.
type Match struct {
	g game.Game
	p Provider

	v *vap.Vap

	w *world.World

	gridPos GridPos
	pos     cube.Pos

	mu sync.Mutex

	s state

	players sets.Set[*user.User]
	dead    sets.Set[*player.Player]
	alive   sets.Set[*player.Player]

	spectators sets.Set[*player.Player]

	statistics map[*player.Player]*Stats

	duel        bool
	updateStats bool

	started  bool
	handlers []Handler

	startCount, endCount  time.Duration
	maxDuration, duration time.Duration

	center mgl64.Vec3

	spawns map[*player.Player]mgl64.Vec3
	blocks sets.Set[cube.Pos]

	c    chan struct{}
	once sync.Once
}

// state represents one of the possible states a match can be in.
type state int

const (
	countDownState state = iota
	fightingState
	endingState
)

// NewMatch initiates and returns a new match, starting it immediately.
func NewMatch(g game.Game, p Provider, pos GridPos, v *vap.Vap, w *world.World, players sets.Set[*user.User], duel, updateStats bool) (*Match, error) {
	if len(players) < 2 {
		return nil, fmt.Errorf("match: match needs at least two players")
	}
	m := &Match{
		g: g,
		v: v,

		w: w,
		p: p,

		gridPos: pos,
		pos:     cube.Pos{pos[0]*1000 + 1000, 0, pos[1]*1000 + 1000},

		players: players,
		alive:   make(sets.Set[*player.Player]),
		dead:    make(sets.Set[*player.Player]),

		spectators: make(sets.Set[*player.Player]),

		statistics: make(map[*player.Player]*Stats),

		spawns: make(map[*player.Player]mgl64.Vec3),
		blocks: make(sets.Set[cube.Pos]),

		s: countDownState,

		duel:        duel,
		updateStats: updateStats,

		startCount: time.Second * 6,
		endCount:   time.Second * 3,

		maxDuration: time.Minute * 30,

		c: make(chan struct{}),
	}
	for u := range players {
		m.alive.Add(u.Player())
	}
	switch g {
	case game.Boxing():
		m.handlers = append(m.handlers, NewHitsHandler(m, 100))
	case game.StickFight(), game.Sumo():
		m.handlers = append(m.handlers, NewRoundsHandler(m, 3))
	}
	m.prepare()
	return m, nil
}

// Lookup searches for a match using a *user.User. If no match was found, the second return value will be false.
func Lookup(p *player.Player) (*Match, bool) {
	prov, ok := LookupProvider(p)
	if !ok {
		return nil, false
	}
	return prov.Match(p)
}

// Spectating ...
func Spectating(p *player.Player) (*Match, bool) {
	prov, ok := LookupProvider(p)
	if !ok {
		return nil, false
	}
	for _, m := range prov.RunningMatches() {
		if m.Spectating(p) {
			return m, true
		}
	}
	return nil, false
}

// Game ...
func (m *Match) Game() game.Game {
	return m.g
}

// Center returns the center spawn of the match.
func (m *Match) Center() mgl64.Vec3 {
	return m.center
}

// LogHit records the provided attacker's hit to the statistics log. If the attacker is not alive, the hit will be ignored.
func (m *Match) LogHit(attacker, attacked *player.Player) {
	m.mu.Lock()
	process := m.alive.Contains(attacker)
	m.mu.Unlock()

	if process {
		stats := m.statistics[attacker]
		for _, h := range m.handlers {
			if !h.HandleUserStartHit(attacker, attacked, stats.Hits+1) {
				// Don't process the hit.
				return
			}
		}

		stats.Hits++
		for _, h := range m.handlers {
			h.HandleUserHit(attacker, attacked)
		}
	}
}

// LoggedHits returns the amount of hits the provided player has logged.
func (m *Match) LoggedHits(p *player.Player) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statistics[p].Hits
}

// LogDamage records the provided attacker's damage dealt to the statistics log. If the attacker is not alive, the damage
// will be ignored.
func (m *Match) LogDamage(attacker, _ *player.Player, h float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.alive.Contains(attacker) {
		m.statistics[attacker].Damage += h
	}
}

// LoggedDamage returns the amount of damage the provided player has dealt.
func (m *Match) LoggedDamage(p *player.Player) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statistics[p].Damage
}

// LogPlacement records the provided block placement to the statistics log.
func (m *Match) LogPlacement(pos cube.Pos) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.s == fightingState {
		m.blocks.Add(pos)
		return true
	}
	return false
}

// LoggedPlacement returns the block placed at the provided position. If no block was placed at the position, the second
// return value will be false.
func (m *Match) LoggedPlacement(pos cube.Pos) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.blocks.Contains(pos)
}

// LogDestruction records the provided block destruction to the statistics log.
func (m *Match) LogDestruction(pos cube.Pos) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks.Delete(pos)
}

// ClearPlacements clears all placed blocks.
func (m *Match) ClearPlacements() {
	for pos := range m.blocks {
		m.w.SetBlock(pos, nil, nil)
	}
	m.blocks.Clear()
}

// TotalAlive ...
func (m *Match) TotalAlive() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.alive)
}

// Players returns all the players in the match.
func (m *Match) Players() []*user.User {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Keys(m.players)
}

// RemovePlayer attempts to remove the given user from the match. If the match has multiple rounds, the removal will not
// be successful, unless force is true.
func (m *Match) RemovePlayer(p *player.Player, force, requested bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if force {
		// Can't do much else but end the match if possible.
		if requested {
			m.disconnect(p, false)
		} else {
			m.alive.Delete(p)
			m.dead.Delete(p)
			providers.Delete(p)
		}
		if len(m.alive) == 1 {
			m.end()
		}
		return m.s != endingState
	}
	if m.s == endingState {
		// The match is already ending, so we can't remove players.
		return false
	}
	if !m.alive.Contains(p) {
		// Already dead.
		return true
	}

	ctx := event.C()
	for _, h := range m.handlers {
		h.HandleUserRemove(ctx, p)
	}
	if ctx.Cancelled() {
		// Cancelled, don't remove the player.
		return true
	}

	t := title.New(lang.Translatef(p.Locale(), "match.title.defeat"))
	t = t.WithSubtitle(lang.Translatef(p.Locale(), "match.subtitle.defeat")).WithFadeInDuration(0)
	t = t.WithDuration(time.Second * 3).WithFadeOutDuration(0)
	p.SendTitle(t)

	m.alive.Delete(p)
	m.dead.Add(p)
	if len(m.alive) == 1 {
		m.end()
	}
	return false
}

// AddSpectator adds a spectator to the match.
func (m *Match) AddSpectator(p *player.Player, silent bool) {
	u, ok := user.Lookup(p)
	if !ok {
		// Somehow left midway through, so just ignore.
		return
	}
	lobby.Lobby().RemovePlayer(p, false)
	p.Inventory().Clear()
	p.Armour().Clear()

	providers.Store(p, m.p)
	p.SetGameMode(world.GameModeSpectator)

	m.w.AddEntity(p)
	p.Teleport(m.center)
	u.Message("match.message.spectate.join")
	m.mu.Lock()
	if !silent {
		m.broadcast(true, "match.message.spectate.joined", u.DisplayName())
	}
	m.spectators.Add(p)
	m.mu.Unlock()
}

// Spectators ...
func (m *Match) Spectators() []*player.Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Keys(m.spectators)
}

// Spectating ...
func (m *Match) Spectating(p *player.Player) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.spectators.Contains(p)
}

// RemoveSpectator ...
func (m *Match) RemoveSpectator(p *player.Player, force bool) {
	providers.Delete(p)

	u, ok := user.Lookup(p)
	if !ok {
		// Somehow left midway through, so just ignore.
		return
	}
	if !force {
		lobby.Lobby().AddPlayer(p)
	}

	u.Message("match.message.spectate.leave")
	m.mu.Lock()
	m.spectators.Delete(p)
	m.broadcast(true, "match.message.spectate.left", u.DisplayName())
	m.mu.Unlock()
}

// Close closes the match and it's ticker.
func (m *Match) Close() {
	m.once.Do(func() {
		close(m.c)

		m.mu.Lock()
		for p := range m.alive {
			m.disconnect(p, true)
		}
		for p := range m.dead {
			m.disconnect(p, true)
		}
		for p := range m.spectators {
			p.Message(lang.Translatef(p.Locale(), "match.message.spectate.end"))
			lobby.Lobby().AddPlayer(p)
			m.spectators.Delete(p)
		}
		m.blocks.Clear()
		m.mu.Unlock()

		d := m.v.Dimensions()
		m.w.BuildStructure(m.pos, emptyStructure(d))

		pos, offset := m.pos.Vec3(), m.pos.Add(d).Vec3()

		for _, e := range m.w.EntitiesWithin(cube.Box(pos.X(), pos.Y(), pos.Z(), offset.X(), offset.Y(), offset.Z()), nil) {
			_ = e.Close()
		}
	})
}

// SendScoreboard updates the scoreboard for the provided player to reflect the match's current state.
func (m *Match) SendScoreboard(p *player.Player) {
	ctx := event.C()
	for _, h := range m.handlers {
		h.HandleScoreboardUpdate(ctx, p)
	}
	if ctx.Cancelled() {
		// Cancelled, don't update the scoreboard.
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.duel {
		board.Send(p,
			"scoreboard.duels",
			parseDuration(m.duration),
			p.Latency().Milliseconds()*2,
			m.opponent(p).Latency().Milliseconds()*2,
		)
		return
	}
	board.Send(p,
		"scoreboard.matches",
		parseDuration(m.duration),
		len(m.alive),
		len(m.players),
	)
}

// disconnect removes a user from the match.
func (m *Match) disconnect(p *player.Player, showStatistics bool) {
	u, ok := user.Lookup(p)
	if ok {
		u.SetNameTagFromRole()
	}

	m.alive.Delete(p)
	providers.Delete(p)

	lobby.Lobby().AddPlayer(p)

	if showStatistics && ok {
		finalStats := make(map[string]any)
		for k, v := range m.statistics {
			finalStats[k.Name()] = *v
		}
		u.SetPostMatchStats(finalStats)
		_ = p.Inventory().SetItem(3, item.NewStack(item.Clock{}, 1).WithCustomName(lang.Translatef(p.Locale(),
			"item.statistics.name",
		)).WithValue("stats", true))
	}
	p.SetMobile()
}

// prepare sets up the match to be played, creating multiple features such as the arena, ticking, etc.
func (m *Match) prepare() {
	m.w.BuildStructure(m.pos, m.v)
	for _, h := range m.handlers {
		h.HandlePrepare(&m.maxDuration)
	}
	for u := range m.players {
		p := u.Player()
		u.DisableProjectiles()

		lobby.Lobby().RemovePlayer(p, false)
		providers.Store(p, m.p)

		kit.Apply(m.g.Kit(false), p)

		p.SetAttackImmunity(time.Second * 6)
		p.SetNameTag(text.Colourf("<red>%v</red>", u.DisplayName()))

		for _, h := range m.handlers {
			h.HandleUserAdd(p)
		}

		m.statistics[p] = &Stats{}
		m.w.AddEntity(p)
	}

	name, _, positions := m.v.Arena()
	offset := m.pos.Vec3()

	if m.duel {
		for u := range m.players {
			p := u.Player()
			otherP := m.opponent(p)
			otherU, ok := user.Lookup(otherP)
			if !ok {
				// Somehow left midway through, just ignore.
				continue
			}

			u.Message("match.message.found")
			u.Message("match.message.opponent", otherU.DisplayName())
			u.Message("match.message.ping", otherU.Latency())

			if m.ranked() {
				elo, otherElo := u.Stats().GameElo[m.g.String()], otherU.Stats().GameElo[m.g.String()]
				eloDiff := lang.Translatef(p.Locale(), "match.message.elo.diff.neutral")
				if elo < otherElo {
					eloDiff = lang.Translatef(p.Locale(),
						"match.message.elo.diff.negative",
						otherElo-elo,
					)
				} else if elo > otherElo {
					eloDiff = lang.Translatef(p.Locale(),
						"match.message.elo.diff.positive",
						elo-otherElo,
					)
				}

				u.Message("match.message.elo", otherElo, eloDiff)
			}

			u.Message("match.message.map", name)
		}
	}

	m.center = offset.Add(positions[1])
	if m.duel {
		players := m.alive.Values()
		m.spawns[players[0]] = offset.Add(positions[0])
		m.spawns[players[1]] = offset.Add(positions[2])
	} else {
		for p := range m.alive {
			m.spawns[p] = m.center
		}
	}
	m.teleport()

	go m.startTicking()
}

// teleport sends all alive players to their spawns.
func (m *Match) teleport() {
	for p := range m.alive {
		pos := m.spawns[p]
		p.Teleport(pos)
		if m.duel {
			oppPos := m.spawns[m.opponent(p)]

			diff := oppPos.Sub(pos)
			diff[1] = 0
			l := diff.Len()
			diff[1] = oppPos[1] - pos[1]

			pitch := mgl64.RadToDeg(-math.Atan2(diff[1], l))
			yaw := mgl64.RadToDeg(math.Atan2(diff[2], diff[0])) - 90
			if yaw < 0 {
				yaw += 360
			}

			// TODO: Remove this hack. (Fix Dragonfly issue #463)
			time.AfterFunc(time.Millisecond*50, func() {
				currentYaw, currentPitch := p.Rotation()
				if u, ok := user.Lookup(p); ok {
					u.Rotate(yaw-currentYaw, pitch-currentPitch)
				}
			})
		}
	}
}

// start commences the start of the match, setting the state to fighting and notifying all participants.
func (m *Match) start() {
	init := !m.started
	players := maps.Keys(m.players)
	m.mu.Unlock()
	for _, u := range players {
		u.SendCustomSound("note.harp", 1, 2, false)
		u.EnableProjectiles()

		if init {
			u.Message("match.message.start")
			u.SetBoard(m)
		}

		p := u.Player()
		p.SendTitle(title.New(" "))
		p.SetMobile()
	}
	m.mu.Lock()

	m.s = fightingState
	m.started = true
	for _, h := range m.handlers {
		h.HandleStart(init)
	}
}

// end commences the end of the match, setting the state to ending.
func (m *Match) end() {
	m.s = endingState
	for u := range m.players {
		p := u.Player()
		m.statistics[p].Items = p.Inventory().Items()
	}

	for p := range m.alive {
		t := title.New(lang.Translatef(p.Locale(), "match.title.victory")).WithFadeInDuration(0)
		t = t.WithSubtitle(lang.Translatef(p.Locale(), "match.subtitle.victory")).WithFadeOutDuration(0)
		t = t.WithDuration(time.Second * 3)
		p.SendTitle(t)

		p.SetAttackImmunity(time.Second * 6)
		p.SetMobile()

		p.SetHeldItems(item.Stack{}, item.Stack{})
		p.Inventory().Clear()
		p.Armour().Clear()
		for _, eff := range p.Effects() {
			p.RemoveEffect(eff.Type())
		}

		if m.duel {
			otherP := m.opponent(p)

			u, ok := user.Lookup(p)
			otherU, okTwo := user.Lookup(otherP)
			if ok && okTwo {
				if m.ranked() {
					user.Broadcast("match.message.ranked.win", u.DisplayName(), otherU.DisplayName())
				}
				m.broadcast(false, "match.message.end")
				m.broadcast(true, "match.message.details", u.DisplayName(), otherU.DisplayName())
				if len(m.spectators) > 0 {
					spectators := make([]string, 0, len(m.spectators))
					for s := range m.spectators {
						if otherS, ok := user.Lookup(s); ok {
							spectators = append(spectators, otherS.DisplayName())
						}
					}
					m.broadcast(true, "match.message.spectators", len(m.spectators), strings.Join(spectators, ", "))
				}

				stats, otherStats := u.Stats(), otherU.Stats()
				if m.ranked() {
					elo, otherElo := stats.GameElo[m.g.String()], otherStats.GameElo[m.g.String()]
					earnedElo, lostElo := eloEarnings(elo, otherElo), eloLosings(otherElo, elo)

					stats.RankedWins++
					stats.Elo = max(stats.Elo+earnedElo, 0)
					stats.GameElo[m.g.String()] = max(elo+earnedElo, 0)
					u.SetStats(stats)

					otherStats.RankedLosses++
					otherStats.Elo = max(otherStats.Elo-lostElo, 0)
					otherStats.GameElo[m.g.String()] = max(otherElo-lostElo, 0)
					otherU.SetStats(otherStats)

					m.broadcast(true, "match.message.elo.changes",
						p.Name(),
						elo,
						earnedElo,
						otherP.Name(),
						otherElo,
						lostElo,
					)
					webhook.Send(webhook.Ranked, hook.Webhook{
						Embeds: []hook.Embed{
							{
								Title: fmt.Sprintf("Ranked %v (%v vs %v)", m.g.Name(), p.Name(), otherP.Name()),
								Description: strings.Join([]string{
									fmt.Sprintf("Winner: **`%v (Disguise: %v - Elo Gained: %v)`**", p.Name(), u.DisplayName(), earnedElo),
									fmt.Sprintf("Loser: **`%v (Disguise: %v - Elo Lost: %v)`**", otherP.Name(), otherU.DisplayName(), lostElo),
									fmt.Sprintf("Duration: **`%v`**", durafmt.ParseShort(m.duration)),
								}, "\n"),
								Color: 0xFFFFFF,
							},
						},
					})
				} else if m.updateStats {
					stats.UnrankedWins++
					u.SetStats(stats)

					otherStats.UnrankedLosses++
					otherU.SetStats(otherStats)
				}
			}
		}
	}
	// TODO: Sort out proper party end logic?
}

// broadcast broadcasts a message to the match players.
func (m *Match) broadcast(spectators bool, key string, a ...any) {
	for u := range m.players {
		u.Message(key, a...)
	}
	if spectators {
		for p := range m.spectators {
			p.Message(lang.Translatef(p.Locale(), key, a...))
		}
	}
}

// startTicking performs a startTicking of the match, updating the state and performing actions depending on the state.
func (m *Match) startTicking() {
	c := make(chan struct{}, 1)
	c <- struct{}{}
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-m.c:
			return
		case <-c:
			m.mu.Lock()
			switch m.s {
			case countDownState:
				wasStarted := m.started
				if m.startCount -= time.Second; m.startCount > 0 {
					remaining := int(m.startCount.Seconds())
					m.broadcast(true, "match.message.count", remaining)
					for u := range m.players {
						p := u.Player()
						if !wasStarted {
							p.SendTitle(title.New(lang.Translatef(p.Locale(), "match.title.count", remaining)))
						}
						u.SendCustomSound("note.harp", 1, 1, false)
					}
				} else {
					m.start()
				}

				if !wasStarted {
					// Don't display the scoreboard if the match didn't start yet.
					m.mu.Unlock()
					continue
				}
				fallthrough
			case fightingState:
				m.duration += time.Second

				p := maps.Keys(m.players)
				m.mu.Unlock()
				for _, u := range p {
					u.Board().SendScoreboard(u.Player())
				}
				m.mu.Lock()

				if m.duration == m.maxDuration {
					m.broadcast(true, "match.message.late")

					m.mu.Unlock()
					m.p.StopMatch(m.g, m)
					continue
				}
			case endingState:
				if m.endCount -= time.Second; m.endCount == 0 {
					m.mu.Unlock()
					m.p.StopMatch(m.g, m)
					continue
				}
			}
			m.mu.Unlock()
		case <-t.C:
			c <- struct{}{}
		}
	}
}

// ranked returns true if the match is under a ranked provider.
func (m *Match) ranked() bool {
	_, ok := m.p.(*RankedProvider)
	return ok
}

// opponent returns a random opponent of the given user.
func (m *Match) opponent(p *player.Player) *player.Player {
	for o := range m.players {
		if op := o.Player(); op != p {
			return op
		}
	}
	panic("should never happen")
}

// parseDuration parses a time.Duration and returns it formatted minutes:seconds.
func parseDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}
