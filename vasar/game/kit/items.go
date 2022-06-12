package kit

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/vasar-network/practice/vasar/game/healing"
	"github.com/vasar-network/vails"

	_ "unsafe"
)

// bed is a dummy item used for the queue kit.
type bed struct{}

// EncodeItem ...
func (bed) EncodeItem() (name string, meta int16) {
	return "minecraft:bed", 14
}

// Apply ...
func Apply(kit vails.Kit, p *player.Player) {
	p.Inventory().Clear()
	p.Armour().Clear()

	p.SetHeldItems(item.Stack{}, item.Stack{})
	if s := player_session(p); s != session.Nop {
		_ = s.SetHeldSlot(0)
	}

	p.StopSneaking()
	p.StopSwimming()
	p.StopSprinting()
	p.StopFlying()
	p.ResetFallDistance()
	p.SetGameMode(world.GameModeSurvival)

	p.Heal(20, healing.SourceKit{})
	p.SetFood(20)
	for _, eff := range p.Effects() {
		p.RemoveEffect(eff.Type())
	}

	inv := p.Inventory()
	armour := kit.Armour(p)
	for slot, it := range kit.Items(p) {
		_ = inv.SetItem(slot, it)
	}
	for _, eff := range kit.Effects(p) {
		p.AddEffect(eff)
	}
	p.Armour().Set(armour[0], armour[1], armour[2], armour[3])
}

// init registers the dummy items/enchantments.
func init() {
	world.RegisterItem(bed{})
}

//go:linkname player_session github.com/df-mc/dragonfly/server/player.(*Player).session
//noinspection ALL
func player_session(*player.Player) *session.Session
