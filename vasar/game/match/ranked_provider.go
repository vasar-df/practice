package match

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sirupsen/logrus"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/sets"
	"golang.org/x/exp/slices"
	"math"
	"strings"
	"time"
)

// RankedProvider handles the management of matches and the players in them.
type RankedProvider struct {
	*provider
}

// NewRankedProvider ...
func NewRankedProvider(log *logrus.Logger) *RankedProvider {
	ranked = &RankedProvider{}
	ranked.provider = newProvider(log, ranked, true)
	return ranked
}

// EnterQueue ...
func (m *RankedProvider) EnterQueue(g game.Game, p *player.Player) {
	u, ok := user.Lookup(p)
	if !ok {
		// User somehow logged out while entering queue.
		return
	}

	kit.Apply(kit.Queue{}, p)
	p.Message(lang.Translatef(p.Locale(), "queue.message.enter", "Ranked", g.Name()))

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

		u.RenewEloRange(g)
		u.UpdateEloRange(u.EloRange().Extend(100, 100))
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
						u.ResetRecentOpponent()
						u.UpdatePingRange(u.PingRange().Extend(2, 2))
						u.UpdateEloRange(u.EloRange().Extend(20, 20))
					}
					p.Message(queueMessage(p.Locale(), g, "Ranked", map[string]any{
						"Elo":      u.EloRange(),
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

// SendScoreboard ...
func (m *RankedProvider) SendScoreboard(p *player.Player) {
	if u, ok := user.Lookup(p); ok {
		users := user.Count()
		g, _ := m.Game(p)
		board.Send(p,
			"scoreboard.queue.ranked",
			users,
			users-lobby.Lobby().PlayerCount(),
			g.Name(),
			u.EloRange(), u.PingRange(),
			parseDuration(time.Since(u.QueuedSince())),
		)
	}
}

// tryMatch tries to match a player with another in queue.
func (m *RankedProvider) tryMatch(g game.Game, u *user.User) {
	m.mu.Lock()
	if !slices.Contains(m.queue[g], u.Player()) {
		m.mu.Unlock()
		return
	}
	for _, o := range m.queue[g] {
		if oU, ok := user.Lookup(o); ok {
			if !canMatchRanked(g, u, oU) {
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

// eloEarnings returns the elo earned using two provided elos.
func eloEarnings(eloOne, eloTwo int32) int32 {
	increase := int32(10)
	if eloOne < 1000 {
		increase = 23
	} else if eloOne >= 1000 && eloOne < 1100 {
		increase = 17
	} else if eloOne >= 1100 && eloOne < 1200 {
		increase = 16
	} else if eloOne >= 1200 && eloOne < 1300 {
		increase = 14
	} else if eloOne >= 1300 && eloOne < 1400 {
		increase = 13
	} else if eloOne >= 1400 && eloOne < 1500 {
		increase = 12
	} else if eloOne >= 1500 && eloOne < 1600 {
		increase = 11
	} else if eloOne >= 1600 && eloOne < 1700 {
		increase = 10
	} else if eloOne >= 1700 && eloOne < 1800 {
		increase = 9
	} else if eloOne >= 1800 && eloOne < 1900 {
		increase = 8
	} else if eloOne >= 1900 && eloOne < 2000 {
		increase = 7
	} else if eloOne >= 2000 {
		increase = 6
	}

	difference := math.Abs(float64(eloOne) - float64(eloTwo))
	if eloOne < eloTwo {
		if difference >= 50 && difference < 100 {
			increase += 2
		} else if difference >= 100 && difference < 150 {
			increase += 4
		} else if difference >= 150 && difference < 200 {
			increase += 6
		} else if difference >= 200 && difference < 250 {
			increase += 8
		} else if difference >= 250 && difference < 300 {
			increase += 10
		} else if difference >= 300 {
			increase += 12
		}
	} else if eloOne > eloTwo {
		if difference >= 50 {
			increase -= 4
		} else if difference >= 100 && difference < 150 {
			increase -= 6
		} else if difference >= 150 && difference < 200 {
			increase -= 8
		} else if difference >= 200 && difference < 250 {
			increase -= 10
		} else if difference >= 250 && difference < 300 {
			increase -= 12
		} else if difference >= 300 {
			increase -= 14
		}
	}
	if increase <= 0 {
		return 1
	} else if increase > 30 {
		return 30
	}
	return increase
}

// eloLosings returns the elo loss using two provided elos.
func eloLosings(eloOne, eloTwo int32) int32 {
	decrease := int32(10)
	if eloOne < 1000 {
		decrease = 7
	} else if eloOne >= 1000 && eloOne < 1200 {
		decrease = 17
	} else if eloOne >= 1200 && eloOne < 1400 {
		decrease = 18
	} else if eloOne >= 1400 && eloOne < 1600 {
		decrease = 19
	} else if eloOne >= 1600 && eloOne < 1800 {
		decrease = 20
	} else if eloOne >= 1800 && eloOne < 2000 {
		decrease = 21
	} else if eloOne >= 2000 && eloOne < 2200 {
		decrease = 22
	} else if eloOne >= 2200 {
		decrease = 25
	}

	difference := math.Abs(float64(eloOne) - float64(eloTwo))
	if eloOne < eloTwo {
		if difference >= 50 && difference < 100 {
			decrease -= 2
		} else if difference >= 100 && difference < 150 {
			decrease -= 4
		} else if difference >= 150 && difference < 200 {
			decrease -= 6
		} else if difference >= 200 && difference < 250 {
			decrease -= 8
		} else if difference >= 250 {
			decrease -= 10
		}
	} else if eloOne > eloTwo {
		if difference > 50 && difference < 100 {
			decrease += 2
		} else if difference >= 100 && difference < 150 {
			decrease += 4
		} else if difference >= 150 && difference < 200 {
			decrease += 6
		} else if difference >= 200 && difference < 250 {
			decrease += 8
		} else if difference >= 250 {
			decrease += 10
		}
	}
	if decrease <= 0 {
		return 1
	} else if decrease > 30 {
		return 30
	}
	return decrease
}

// max returns the maximum of two integers.
func max(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

// canMatchRanked ...
func canMatchRanked(g game.Game, u, o *user.User) bool {
	if !canMatch(u, o) {
		// If we can't match regularly, we can't match in a ranked game-mode.
		return false
	}
	return u.EloRange().Compare(int(o.Stats().GameElo[g.String()])) &&
		o.EloRange().Compare(int(u.Stats().GameElo[g.String()]))
}
