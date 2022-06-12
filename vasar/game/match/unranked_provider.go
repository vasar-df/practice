package match

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sirupsen/logrus"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/sets"
	"github.com/vasar-network/vap"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

// UnrankedProvider handles the management of matches and the players in them.
type UnrankedProvider struct {
	*provider

	requests map[*player.Player]request

	duels bool
}

// request ...
type request struct {
	t, r *player.Player
	g    game.Game
	v    *vap.Vap

	c chan struct{}
}

// newRequest ...
func newRequest(p *UnrankedProvider, to, r *player.Player, g game.Game, v *vap.Vap) request {
	req := request{t: to, r: r, g: g, v: v, c: make(chan struct{})}
	go func() {
		t := time.NewTimer(time.Second * 30)
		select {
		case <-t.C:
			rU, ok := user.Lookup(req.r)
			tU, okTwo := user.Lookup(req.t)
			if !ok || !okTwo {
				// Somehow left midway through, so just ignore this.
				return
			}

			tU.Message("duel.expired", rU.DisplayName())
			rU.Message("duel.expire", tU.DisplayName())
			p.RemoveRequestTo(to)
		case <-req.c:
			t.Stop()
		}
	}()
	return req
}

// NewUnrankedProvider ...
func NewUnrankedProvider(log *logrus.Logger, duels bool) *UnrankedProvider {
	unranked = &UnrankedProvider{requests: make(map[*player.Player]request), duels: duels}
	unranked.provider = newProvider(log, unranked, duels)
	return unranked
}

// EnterQueue ...
func (m *UnrankedProvider) EnterQueue(g game.Game, p *player.Player) {
	u, ok := user.Lookup(p)
	if !ok {
		// User somehow logged out while entering the queue.
		return
	}

	kit.Apply(kit.Queue{}, p)
	u.Message("queue.message.enter", "Unranked", g.Name())

	providers.Store(p, m)

	m.mu.Lock()
	m.queue[g] = append(m.queue[g], p)
	ch := make(chan struct{})
	m.queueChan[g][p] = ch
	m.mu.Unlock()

	go func() {
		desiredPing := u.DesiredPingRange()
		desiredGroups := make([]string, 0, len(u.DesiredDeviceGroups()))
		for _, group := range u.DesiredDeviceGroups() {
			desiredGroups = append(desiredGroups, group.String())
		}

		u.RenewPingRange()
		u.UpdatePingRange(u.PingRange().Extend(desiredPing.Max(), desiredPing.Max()))
		u.RenewQueuedSince()

		u.SetBoard(m)

		c := make(chan struct{}, 1)
		c <- struct{}{}

		var duration time.Duration
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-c:
				remaining := int(duration.Seconds())
				if remaining > 0 && remaining%5 == 0 {
					u.ResetRecentOpponent()
				}
				if remaining%10 == 0 {
					if remaining > 0 {
						u.UpdatePingRange(u.PingRange().Extend(2, 2))
					}
					p.Message(queueMessage(p.Locale(), g, "Unranked", map[string]any{
						"Ping":     u.PingRange(),
						"Queueing": strings.Join(desiredGroups, ", "),
					}))
					m.tryMatch(g, u)
				}

				u.Board().SendScoreboard(p)
			case <-t.C:
				duration += time.Second
				c <- struct{}{}
			case <-ch:
				return
			}
		}
	}()
}

// RequestDuel ...
func (m *UnrankedProvider) RequestDuel(g game.Game, v *vap.Vap, from, to *player.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[to] = newRequest(m, to, from, g, v)
	fU, ok := user.Lookup(from)
	tU, okTwo := user.Lookup(to)
	if !ok || !okTwo {
		// Somehow left midway through, so just ignore this.
		return
	}
	fU.Message("duel.request", g.Name(), tU.DisplayName())
	tU.Message("duel.requested", fU.DisplayName(), fU.Latency(), g.Name())
}

// RemoveRequestTo ...
func (m *UnrankedProvider) RemoveRequestTo(to *player.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.requests, to)
}

// RemoveRequestsFrom ...
func (m *UnrankedProvider) RemoveRequestsFrom(from *player.Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for to, r := range m.requests {
		if from == r.r {
			close(r.c)
			delete(m.requests, to)
		}
	}
}

// AcceptDuel ...
func (m *UnrankedProvider) AcceptDuel(p *player.Player) bool {
	m.mu.Lock()
	r, ok := m.requests[p]
	if !ok {
		m.mu.Unlock()
		return false
	}
	delete(m.requests, p)
	m.mu.Unlock()

	close(r.c)

	m.RemoveRequestTo(r.r)
	m.RemoveRequestsFrom(r.r)
	m.RemoveRequestsFrom(r.t)

	fU, ok := user.Lookup(r.r)
	tU, okTwo := user.Lookup(r.t)
	if !ok || !okTwo {
		// Somehow left midway through, so just ignore this.
		return false
	}
	fU.Message("duel.accepted", tU.DisplayName())
	tU.Message("duel.accept", fU.DisplayName())

	t, ok := user.Lookup(p)
	f, okTwo := user.Lookup(r.r)
	if ok && okTwo {
		m.StartMatch(r.g, r.v, sets.New[*user.User](t, f), false)
		return true
	}
	return false
}

// DeclineDuel ...
func (m *UnrankedProvider) DeclineDuel(p *player.Player) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.requests[p]
	if !ok {
		return false
	}
	delete(m.requests, p)

	fU, ok := user.Lookup(r.r)
	tU, okTwo := user.Lookup(r.t)
	if !ok || !okTwo {
		// Somehow left midway through, so just ignore this.
		return false
	}
	fU.Message("duel.declined", tU.DisplayName())
	tU.Message("duel.decline", fU.DisplayName())
	return true
}

// SendScoreboard ...
func (m *UnrankedProvider) SendScoreboard(p *player.Player) {
	if u, ok := user.Lookup(p); ok {
		users := user.Count()
		g, _ := m.Game(p)
		board.Send(p,
			"scoreboard.queue.unranked",
			users,
			users-lobby.Lobby().PlayerCount(),
			g.Name(),
			u.PingRange(),
			parseDuration(time.Since(u.QueuedSince())),
		)
	}
}

// tryMatch tries to match a player with another in queue.
func (m *UnrankedProvider) tryMatch(g game.Game, u *user.User) {
	m.mu.Lock()
	if !slices.Contains(m.queue[g], u.Player()) {
		m.mu.Unlock()
		return
	}
	for _, o := range m.queue[g] {
		if oU, ok := user.Lookup(o); ok {
			if !canMatch(u, oU) {
				// Check the next member in queue.
				continue
			}

			u.SetRecentOpponent(oU)
			oU.SetRecentOpponent(u)

			for _, u := range []*user.User{u, oU} {
				p := u.Player()
				ind := slices.Index(m.queue[g], p)
				m.queue[g] = slices.Delete(m.queue[g], ind, ind+1)

				c := m.queueChan[g][p]
				close(c)
				delete(m.queueChan[g], p)
			}

			m.mu.Unlock()
			m.StartMatch(g, RandomArena(g), sets.New(u, oU), true)
			return
		}
	}
	m.mu.Unlock()
}
