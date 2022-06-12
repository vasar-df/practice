package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
)

// Sumo represents the kit given when players join Sumo.
type Sumo struct {
	// FFA returns true if the kit is for FFA, varying some options such as effects and potions.
	FFA bool
}

// Items ...
func (Sumo) Items(*player.Player) [36]item.Stack {
	return [36]item.Stack{}
}

// Armour ...
func (Sumo) Armour(*player.Player) [4]item.Stack {
	return [4]item.Stack{}
}

// Effects ...
func (s Sumo) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
