package module

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	it "github.com/vasar-network/practice/vasar/item"
	"github.com/vasar-network/practice/vasar/user"
)

// Rods is a module providing functionality for fishing rods.
type Rods struct {
	player.NopHandler

	u *user.User
}

// NewRods ...
func NewRods(u *user.User) *Rods {
	return &Rods{u: u}
}

// HandleItemUse ...
func (r *Rods) HandleItemUse(_ *event.Context) {
	p := r.u.Player()
	held, _ := p.HeldItems()
	if i, ok := held.Item().(it.FishingRod); ok && !p.HasCooldown(i) {
		r.u.ToggleRod()
	}
}
