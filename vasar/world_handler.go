package vasar

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/vasar-network/practice/vasar/entity"
)

// WorldHandler ...
type WorldHandler struct {
	world.NopHandler
}

// HandleEntitySpawn ...
func (WorldHandler) HandleEntitySpawn(e world.Entity) {
	if v, ok := e.(*entity.VasarPearl); ok {
		v.Handle(PearlHandler{})
	} else if p, ok := e.(*entity.SplashPotion); ok {
		p.Handle(PotionHandler{})
	}
}

// HandleSound ...
func (WorldHandler) HandleSound(ctx *event.Context, s world.Sound, _ mgl64.Vec3) {
	if _, ok := s.(sound.Attack); ok {
		ctx.Cancel()
	}
}

// HandleLiquidHarden ...
func (WorldHandler) HandleLiquidHarden(ctx *event.Context, _ cube.Pos, _, _, _ world.Block) {
	ctx.Cancel()
}
