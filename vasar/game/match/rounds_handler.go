package match

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/vails/lang"
	"time"
)

// RoundsHandler represents a handler to implement rounded functionality for a match.
type RoundsHandler struct {
	NopHandler

	m *Match

	currentRound int
	totalRounds  int

	statistics map[*player.Player]*struct {
		wins   int
		losses int
	}
}

// NewRoundsHandler creates a new instance of a rounds handler.
func NewRoundsHandler(m *Match, rounds int) *RoundsHandler {
	if !m.duel {
		panic("rounds handler: match must be a duel")
	}
	return &RoundsHandler{
		m:           m,
		totalRounds: rounds,
		statistics: make(map[*player.Player]*struct {
			wins   int
			losses int
		}),
	}
}

// HandlePrepare ...
func (r *RoundsHandler) HandlePrepare(duration *time.Duration) {
	*duration = time.Minute * 15
	for u := range r.m.players {
		u.Player().SetImmobile()
	}
}

// HandleStart ...
func (r *RoundsHandler) HandleStart(initial bool) {
	for u := range r.m.players {
		if initial {
			u.Message("match.message.rounds")
		} else {
			u.Message("round.message.start")
		}
	}
}

// HandleUserAdd ...
func (r *RoundsHandler) HandleUserAdd(p *player.Player) {
	r.statistics[p] = &struct {
		wins   int
		losses int
	}{}
}

// HandleUserRemove ...
func (r *RoundsHandler) HandleUserRemove(ctx *event.Context, p *player.Player) {
	if ctx.Cancelled() {
		// Was cancelled, not much we can do.
		return
	}

	o := r.m.opponent(p)
	stats, opponentStats := r.statistics[p], r.statistics[o]

	stats.losses++
	opponentStats.wins++
	if opponentStats.wins < r.totalRounds {
		ctx.Cancel()

		r.m.startCount = time.Second * 4
		r.m.s = countDownState
		r.currentRound++

		t := title.New(lang.Translatef(p.Locale(), "round.title.lost")).WithFadeInDuration(0)
		t = t.WithSubtitle(lang.Translatef(p.Locale(), "round.subtitle.info", stats.wins, stats.losses))
		t = t.WithDuration(time.Second * 2).WithFadeOutDuration(time.Second)
		p.SendTitle(t)

		t = title.New(lang.Translatef(o.Locale(), "round.title.won")).WithFadeInDuration(0)
		t = t.WithSubtitle(lang.Translatef(o.Locale(), "round.subtitle.info", opponentStats.wins, opponentStats.losses))
		t = t.WithDuration(time.Second * 2).WithFadeOutDuration(time.Second)
		o.SendTitle(t)

		for pl := range r.m.players {
			kit.Apply(r.m.g.Kit(false), pl.Player())
		}

		p.SetAttackImmunity(r.m.startCount)
		o.SetAttackImmunity(r.m.startCount)
		p.SetImmobile()
		o.SetImmobile()

		r.m.ClearPlacements()
		r.m.teleport()
	}
}
