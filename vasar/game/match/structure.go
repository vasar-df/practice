package match

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

// emptyStructure is an empty structure implementation, allowing areas to be cleared.
type emptyStructure [3]int

// Dimensions ...
func (d emptyStructure) Dimensions() [3]int {
	return d
}

// At ...
func (d emptyStructure) At(int, int, int, func(x int, y int, z int) world.Block) (world.Block, world.Liquid) {
	return block.Air{}, nil
}
