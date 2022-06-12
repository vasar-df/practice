package item

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	ent "github.com/vasar-network/practice/vasar/entity"
)

// VasarPotion is an edited item for splash potions.
type VasarPotion struct {
	// Type is the type of splash potion.
	Type potion.Potion
}

// Use ...
func (v VasarPotion) Use(w *world.World, user item.User, ctx *item.UseContext) bool {
	force := 0.45
	debuff := shouldDebuff(v.Type)
	if debuff {
		force = 0.65
	}

	if p, ok := user.(*player.Player); ok && p.Sprinting() {
		force += 0.05
	}

	yaw, pitch := user.Rotation()
	e := ent.NewSplashPotion(entity.EyePosition(user), entity.DirectionVector(user).Mul(force), yaw, pitch, v.Type, debuff, user)
	w.AddEntity(e)

	w.PlaySound(user.Position(), sound.ItemThrow{})
	ctx.SubtractFromCount(1)
	return true
}

// MaxCount ...
func (v VasarPotion) MaxCount() int {
	return 1
}

// EncodeItem ...
func (v VasarPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:splash_potion", int16(v.Type.Uint8())
}

// shouldDebuff returns true if the potion is a debuff potion.
func shouldDebuff(p potion.Potion) bool {
	switch p {
	case potion.Slowness(), potion.LongSlowness(), potion.StrongSlowness(), potion.Harming(), potion.StrongHarming(),
		potion.Poison(), potion.LongPoison(), potion.StrongPoison(), potion.Weakness(), potion.LongWeakness(),
		potion.Wither(), potion.TurtleMaster(), potion.LongTurtleMaster(), potion.StrongTurtleMaster():
		return true
	}
	return false
}
