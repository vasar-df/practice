package match

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Stats is a struct containing the statistics of a player.
type Stats struct {
	// Hits is the number of hits the player dealt before the match ended.
	Hits int
	// Damage is the total damage the player dealt to opponents before the match ended.
	Damage float64
	// Items contains the items the player had before the match ended.
	Items []item.Stack
}
