package kit

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/vails/lang"
)

// Lobby represents the kit given when players join the lobby.
type Lobby struct{}

// Items ...
func (Lobby) Items(p *player.Player) [36]item.Stack {
	return [36]item.Stack{
		0: item.NewStack(item.Sword{Tier: item.ToolTierDiamond}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.unranked"),
		).WithValue("lobby", 0),
		1: item.NewStack(item.Sword{Tier: item.ToolTierIron}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.ranked"),
		).WithValue("lobby", 1),
		2: item.NewStack(item.Sword{Tier: item.ToolTierStone}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.ffa"),
		).WithValue("lobby", 2),

		6: item.NewStack(item.NetherStar{}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.spectate"),
		).WithValue("lobby", 3),
		7: item.NewStack(item.NetherStar{}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.party"),
		).WithValue("lobby", 4),
		8: item.NewStack(item.EnchantedBook{}, 1).WithCustomName(
			"§r"+lang.Translatef(p.Locale(), "item.lobby.settings"),
		).WithValue("lobby", 5),
	}
}

// Armour ...
func (Lobby) Armour(*player.Player) [4]item.Stack {
	return [4]item.Stack{}
}

// Effects ...
func (Lobby) Effects(*player.Player) []effect.Effect {
	return []effect.Effect{}
}
