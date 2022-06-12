package lobby

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
)

// inventoryHandler handles inventory related stuff.
type inventoryHandler inventory.NopHandler

// HandleTake ...
func (h inventoryHandler) HandleTake(ctx *event.Context, _ int, _ item.Stack) {
	ctx.Cancel()
}

// HandlePlace ...
func (h inventoryHandler) HandlePlace(ctx *event.Context, _ int, _ item.Stack) {
	ctx.Cancel()
}

// HandleDrop ...
func (h inventoryHandler) HandleDrop(ctx *event.Context, _ int, _ item.Stack) {
	ctx.Cancel()
}
