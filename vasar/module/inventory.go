package module

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/role"
	"time"
)

// Inventory is a module that adds all items required to access basic functions outside the lobby.
type Inventory struct {
	player.NopHandler

	u *user.User
}

// NewInventory ...
func NewInventory(u *user.User) *Inventory {
	return &Inventory{u: u}
}

// HandleItemUseOnBlock ...
func (i *Inventory) HandleItemUseOnBlock(ctx *event.Context, _ cube.Pos, _ cube.Face, _ mgl64.Vec3) {
	p := i.u.Player()
	h, _ := p.HeldItems()
	if _, ok := h.Value("queue"); ok {
		if prov, ok := match.LookupProvider(p); ok {
			prov.ExitQueue(p)
			ctx.Cancel()
		}
	}
}

// HandleItemUse ...
func (i *Inventory) HandleItemUse(*event.Context) {
	p := i.u.Player()
	h, _ := p.HeldItems()
	if request, ok := h.Value("lobby"); ok {
		switch request {
		case 0:
			p.SendForm(form.NewUnrankedDuels())
		case 1:
			if !i.u.Roles().Contains(role.Plus{}, role.Operator{}) {
				if stats := i.u.Stats(); stats.UnrankedWins < 10 {
					i.u.Message("ranked.locked", 10-stats.UnrankedWins)
					return
				}
			}
			p.SendForm(form.NewRankedDuels())
		case 2:
			p.SendForm(form.NewFFA())
		case 3:
			p.SendForm(form.NewSpectate())
		case 5:
			p.SendForm(form.NewSettings(i.u))
		default:
			p.Message(text.Colourf("<red>This action isn't implemented yet.</red>"))
		}
	} else if _, ok = h.Value("queue"); ok {
		if prov, ok := match.LookupProvider(p); ok {
			prov.ExitQueue(p)
		}
	} else if _, ok := h.Value("stats"); ok {
		p.SendForm(form.NewPostMatchStats(i.u))
	}
}

// HandleItemConsume ...
func (i *Inventory) HandleItemConsume(_ *event.Context, h item.Stack) {
	if _, ok := h.Value("head"); ok {
		i.u.Player().AddEffect(effect.New(effect.Regeneration{}, 3, time.Second*9))
	}
}
