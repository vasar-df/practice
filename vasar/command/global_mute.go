package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/practice/vasar"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// GlobalMute is a command that globally mutes the chat.
type GlobalMute struct {
	v *vasar.Vasar
}

// NewGlobalMute ...
func NewGlobalMute(v *vasar.Vasar) GlobalMute {
	return GlobalMute{v: v}
}

// Run ...
func (g GlobalMute) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if g.v.ToggleGlobalMute() {
		o.Printf(lang.Translatef(l, "command.globalmute.disabled"))
	} else {
		o.Printf(lang.Translatef(l, "command.globalmute.enabled"))
	}
}

// Allow ...
func (GlobalMute) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}
