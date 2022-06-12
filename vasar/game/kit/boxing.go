package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/player"
	"time"
)

// Boxing represents the kit given when players join Boxing.
type Boxing struct{}

// Items ...
func (Boxing) Items(*player.Player) [36]item.Stack {
	return [36]item.Stack{
		item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithEnchantments(item.NewEnchantment(enchantment.Unbreaking{}, 20)),
	}
}

// Armour ...
func (Boxing) Armour(*player.Player) [4]item.Stack {
	return [4]item.Stack{}
}

// Effects ...
func (Boxing) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{
		effect.New(effect.Speed{}, 1, time.Hour*24).WithoutParticles(),
	}
}
