package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
)

// Queue represents the kit given when players join queue.
type Queue struct{}

// Items ...
func (Queue) Items(*player.Player) [36]item.Stack {
	return [36]item.Stack{item.NewStack(bed{}, 1).WithCustomName("§r§cLeave Queue").WithValue("queue", 0)}
}

// Armour ...
func (Queue) Armour(*player.Player) [4]item.Stack {
	return [4]item.Stack{}
}

// Effects ...
func (Queue) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
