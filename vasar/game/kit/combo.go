package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	it "github.com/vasar-network/practice/vasar/item"
)

// Combo represents the kit given when players join Combo.
type Combo struct{}

// Items ...
func (Combo) Items(*player.Player) [36]item.Stack {
	return [36]item.Stack{
		0: item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithEnchantments(item.NewEnchantment(enchantment.Unbreaking{}, 5)),
		1: item.NewStack(item.EnchantedApple{}, 32),

		7: item.NewStack(it.Potion{Type: potion.Swiftness()}, 1),
		8: item.NewStack(it.Potion{Type: potion.Swiftness()}, 1),
	}
}

// Armour ...
func (Combo) Armour(*player.Player) [4]item.Stack {
	durability := item.NewEnchantment(enchantment.Unbreaking{}, 5)
	return [4]item.Stack{
		item.NewStack(item.Helmet{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Leggings{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Boots{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
	}
}

// Effects ...
func (Combo) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
