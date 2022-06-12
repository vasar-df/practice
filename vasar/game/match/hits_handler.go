package match

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/user"
	"time"
)

// HitsHandler represents a handler to implement hit-based winning functionality for a match. This is used in certain
// game-modes, such as boxing.
type HitsHandler struct {
	NopHandler

	m *Match

	hitsToWin int
}

// NewHitsHandler creates a new hits handler.
func NewHitsHandler(m *Match, hitsToWin int) *HitsHandler {
	if !m.duel {
		panic("hits handler: match must be a duel")
	}
	return &HitsHandler{m: m, hitsToWin: hitsToWin}
}

// HandlePrepare ...
func (h *HitsHandler) HandlePrepare(duration *time.Duration) {
	*duration = time.Minute * 15
}

// HandleStart ...
func (h *HitsHandler) HandleStart(initial bool) {
	if !initial {
		// Only send tips if this is the first start.
		return
	}
	for u := range h.m.players {
		u.Message("match.message.hits")
	}
}

// HandleScoreboardUpdate ...
func (h *HitsHandler) HandleScoreboardUpdate(ctx *event.Context, p *player.Player) {
	h.m.mu.Lock()
	defer h.m.mu.Unlock()

	o := h.m.opponent(p)

	hitsA, hitsB := h.m.statistics[p].Hits, h.m.statistics[o].Hits
	diff := text.Colourf("<gold>(0)</gold>")
	if hitsA < hitsB {
		diff = text.Colourf("<red>(-%d)</red>", hitsB-hitsA)
	} else if hitsA > hitsB {
		diff = text.Colourf("<green>(+%d)</green>", hitsA-hitsB)
	}

	ctx.Cancel()
	board.Send(p,
		"scoreboard.duels.boxing",
		parseDuration(h.m.duration),
		diff,
		hitsA,
		hitsB,
		p.Latency().Milliseconds()*2,
		o.Latency().Milliseconds()*2,
	)
}

// surpassedHits represents the surpassed hits damage source, used primarily for boxing deaths.
type surpassedHits struct{}

// ReducedByArmour ...
func (surpassedHits) ReducedByArmour() bool {
	return false
}

// ReducedByResistance ...
func (surpassedHits) ReducedByResistance() bool {
	return false
}

// HandleUserStartHit ...
func (h *HitsHandler) HandleUserStartHit(_, victim *player.Player, hits int) bool {
	if hits == h.hitsToWin {
		victim.Hurt(victim.MaxHealth(), surpassedHits{})
		return false
	}
	return true
}

// HandleUserHit ...
func (h *HitsHandler) HandleUserHit(attacker, victim *player.Player) {
	if a, ok := user.Lookup(attacker); ok {
		a.Board().SendScoreboard(attacker)
	}
	if v, ok := user.Lookup(victim); ok {
		v.Board().SendScoreboard(victim)
	}
}
