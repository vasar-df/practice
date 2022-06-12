package item

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/creative"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// FishingRod ...
type FishingRod struct{}

// init registers FishingRod as an item and adds it to the creative inventory.
func init() {
	world.RegisterItem(FishingRod{})
	creative.RegisterItem(item.NewStack(FishingRod{}, 1))
}

func (FishingRod) DurabilityInfo() item.DurabilityInfo {
	return item.DurabilityInfo{
		MaxDurability: 355,
		BrokenItem:    func() item.Stack { return item.Stack{} },
	}
}

// MaxCount ...
func (FishingRod) MaxCount() int {
	return 1
}

// Rod ...
func (FishingRod) Rod() bool {
	return true
}

// Cooldown ...
func (FishingRod) Cooldown() time.Duration {
	return time.Second
}

// Use ...
func (FishingRod) Use(_ *world.World, _ item.User, ctx *item.UseContext) bool {
	ctx.DamageItem(1)
	return true
}

// EncodeItem ...
func (FishingRod) EncodeItem() (name string, meta int16) {
	return "minecraft:fishing_rod", 0
}
