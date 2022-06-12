package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	it "github.com/vasar-network/practice/vasar/item"
)

// Gapple represents the kit given when players join Gapple.
type Gapple struct{}

// Items ...
func (Gapple) Items(*player.Player) [36]item.Stack {
	items := [36]item.Stack{
		item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithEnchantments(item.NewEnchantment(enchantment.Unbreaking{}, 10)),
		item.NewStack(item.GoldenApple{}, 16),
	}

	items[2] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
	items[26] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
	items[35] = item.NewStack(it.Potion{Type: potion.Swiftness()}, 1)
	return items
}

// Armour ...
func (Gapple) Armour(*player.Player) [4]item.Stack {
	durability := item.NewEnchantment(enchantment.Unbreaking{}, 10)
	return [4]item.Stack{
		item.NewStack(item.Helmet{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Leggings{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Boots{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
	}
}

// Effects ...
func (Gapple) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
