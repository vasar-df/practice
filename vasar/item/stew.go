package item

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/vasar-network/practice/vasar/game/healing"
)

// Stew is a food item.
type Stew struct{}

// MaxCount ...
func (Stew) MaxCount() int {
	return 1
}

// Use ...
func (s Stew) Use(_ *world.World, user item.User, ctx *item.UseContext) bool {
	living, ok := user.(entity.Living)
	if ok && living.Health() < living.MaxHealth() {
		living.Heal(7, healing.SourceStew{})
		ctx.SubtractFromCount(1)
	}
	return ok
}

// EncodeItem ...
func (Stew) EncodeItem() (name string, meta int16) {
	return "minecraft:mushroom_stew", 0
}
