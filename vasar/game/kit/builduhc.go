package kit

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	it "github.com/vasar-network/practice/vasar/item"
	"time"
)

// BuildUHC represents the kit given when players join BuildUHC.
type BuildUHC struct{}

// Items ...
func (BuildUHC) Items(*player.Player) [36]item.Stack {
	durability := item.NewEnchantment(enchantment.Unbreaking{}, 10)
	efficiency := item.NewEnchantment(enchantment.Efficiency{}, 1)
	return [36]item.Stack{
		item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(it.FishingRod{}, 1),
		item.NewStack(item.Bow{}, 1),
		item.NewStack(item.GoldenApple{}, 6),
		item.NewStack(item.GoldenApple{}, 3).WithCustomName("§r"+text.Colourf("<yellow>Golden Head</yellow>")).WithValue("head", true),
		item.NewStack(item.Pickaxe{Tier: item.ToolTierDiamond}, 1).WithEnchantments(durability, efficiency),
		item.NewStack(item.Axe{Tier: item.ToolTierDiamond}, 1).WithEnchantments(durability, efficiency),
		item.NewStack(block.Planks{}, 64),
		item.NewStack(block.Cobblestone{}, 64),
		item.NewStack(item.Arrow{}, 64),
		item.NewStack(item.Bucket{Content: block.Water{}}, 1),
		item.NewStack(item.Bucket{Content: block.Water{}}, 1),
		item.NewStack(item.Bucket{Content: block.Lava{}}, 1),
		item.NewStack(item.Bucket{Content: block.Lava{}}, 1),
	}
}

// Armour ...
func (BuildUHC) Armour(*player.Player) [4]item.Stack {
	durability := item.NewEnchantment(enchantment.Unbreaking{}, 10)
	return [4]item.Stack{
		item.NewStack(item.Helmet{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Chestplate{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Leggings{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
		item.NewStack(item.Boots{Tier: item.ArmourTierDiamond}, 1).WithEnchantments(durability),
	}
}

// Effects ...
func (BuildUHC) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{effect.New(effect.Speed{}, 1, time.Hour*24).WithoutParticles()}
}
