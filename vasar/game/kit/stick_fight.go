package kit

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
)

// StickFight represents the kit given when players join Stick Fight.
type StickFight struct{}

// Items ...
func (StickFight) Items(*player.Player) [36]item.Stack {
	return [36]item.Stack{
		item.NewStack(item.Stick{}, 1),
		item.NewStack(item.Shears{}, 1),
		item.NewStack(block.Wool{}, 6),
	}
}

// Armour ...
func (StickFight) Armour(*player.Player) [4]item.Stack {
	return [4]item.Stack{
		item.NewStack(item.Helmet{Tier: item.ArmourTierLeather}, 1),
		item.NewStack(item.Chestplate{Tier: item.ArmourTierLeather}, 1),
		item.NewStack(item.Leggings{Tier: item.ArmourTierLeather}, 1),
		item.NewStack(item.Boots{Tier: item.ArmourTierLeather}, 1),
	}
}

// Effects ...
func (s StickFight) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
