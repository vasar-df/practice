package module

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	"time"
)

// Protection is a module that ensures that players cannot break blocks or attack players in the lobby.
type Protection struct {
	player.NopHandler

	p *player.Player
}

// NewProtection ...
func NewProtection(p *player.Player) *Protection {
	return &Protection{p: p}
}

// HandleAttackEntity ...
func (p *Protection) HandleAttackEntity(ctx *event.Context, _ world.Entity, _ *float64, _ *float64, _ *bool) {
	if _, ok := lobby.LookupProvider(p.p); ok {
		ctx.Cancel()
	}
}

// HandleHurt ...
func (p *Protection) HandleHurt(ctx *event.Context, _ *float64, _ *time.Duration, s damage.Source) {
	if _, ok := lobby.LookupProvider(p.p); ok || (s == damage.SourceFall{}) {
		ctx.Cancel()
	}
}

// HandleFoodLoss ...
func (*Protection) HandleFoodLoss(ctx *event.Context, _ int, _ int) {
	ctx.Cancel()
}

// HandleBlockPlace ...
func (p *Protection) HandleBlockPlace(ctx *event.Context, pos cube.Pos, _ world.Block) {
	if m, ok := match.Lookup(p.p); ok && cube.PosFromVec3(m.Center()).Y()+5 > pos.Y() && m.LogPlacement(pos) {
		// Placement log was successful, don't cancel.
		return
	}
	ctx.Cancel()
}

// HandleBlockBreak ...
func (p *Protection) HandleBlockBreak(ctx *event.Context, pos cube.Pos, _ *[]item.Stack) {
	if m, ok := match.Lookup(p.p); ok && m.LoggedPlacement(pos) {
		// Placement was logged, so we can allow this break.
		return
	}
	ctx.Cancel()
}
