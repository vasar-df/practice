package item

import (
	_ "embed"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	ent "github.com/vasar-network/practice/vasar/entity"
	"time"
)

// VasarPearl is an edited item for ender pearls.
type VasarPearl struct{}

// Use ...
func (VasarPearl) Use(w *world.World, user item.User, ctx *item.UseContext) bool {
	yaw, pitch := user.Rotation()
	e := ent.NewEnderPearl(entity.EyePosition(user), entity.DirectionVector(user).Mul(2.3), yaw, pitch, user)
	w.AddEntity(e)

	w.PlaySound(user.Position(), sound.ItemThrow{})
	ctx.SubtractFromCount(1)
	return true
}

// Cooldown ...
func (VasarPearl) Cooldown() time.Duration {
	return time.Second
}

// MaxCount ...
func (VasarPearl) MaxCount() int {
	return 16
}

// EncodeItem ...
func (VasarPearl) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_pearl", 0
}
