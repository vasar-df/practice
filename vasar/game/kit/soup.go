package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	it "github.com/vasar-network/practice/vasar/item"
	"time"
)

// Soup represents the kit given when players join Soup.
type Soup struct {
	// FFA returns true if the kit is for FFA, varying some options such as effects and potions.
	FFA bool
}

// Items ...
func (s Soup) Items(*player.Player) [36]item.Stack {
	items := [36]item.Stack{
		item.NewStack(item.Sword{Tier: item.ToolTierIron}, 1).WithEnchantments(item.NewEnchantment(enchantment.Sharpness{}, 1), item.NewEnchantment(enchantment.Unbreaking{}, 3)),
	}
	for i := 1; i < 36; i++ {
		items[i] = item.NewStack(it.Stew{}, 1)
	}
	if !s.FFA {
		items[1] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
		items[26] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
		items[35] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
	}
	return items
}

// Armour ...
func (Soup) Armour(*player.Player) [4]item.Stack {
	durability := item.NewEnchantment(enchantment.Unbreaking{}, 3)
	protection := item.NewEnchantment(enchantment.Protection{}, 2)
	return [4]item.Stack{
		item.NewStack(item.Helmet{Tier: item.ArmourTierIron}, 1).WithEnchantments(durability, item.NewEnchantment(enchantment.Protection{}, 2)),
		item.NewStack(item.Chestplate{Tier: item.ArmourTierIron}, 1).WithEnchantments(durability, protection),
		item.NewStack(item.Leggings{Tier: item.ArmourTierIron}, 1).WithEnchantments(durability, protection),
		item.NewStack(item.Boots{Tier: item.ArmourTierIron}, 1).WithEnchantments(durability, protection),
	}
}

// Effects ...
func (s Soup) Effects(*player.Player) []effect.Effect {
	if s.FFA {
		return []effect.Effect{effect.New(effect.Speed{}, 1, time.Hour*24).WithoutParticles()}
	}
	return []effect.Effect{}
}
