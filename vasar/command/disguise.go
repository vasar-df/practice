package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// Disguise is a command that allows a player to disguise themselves as another player.
type Disguise struct{}

// Run ...
func (Disguise) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	o.Error(lang.Translatef(l, "action.unimplemented"))
}

// Allow ...
func (Disguise) Allow(s cmd.Source) bool {
	return allow(s, false, role.Plus{}, role.Trial{}, role.Famous{}, role.Nitro{})
}
