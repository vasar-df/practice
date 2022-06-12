package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// Freeze is a command used to freeze a player.
type Freeze struct {
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (f Freeze) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(f.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	target, ok := f.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if s == target {
		o.Error(lang.Translatef(l, "command.usage.self"))
		return
	}
	t, ok := user.Lookup(target)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if t.Frozen() {
		user.Alert(s, "staff.alert.unfreeze", target.Name())
		o.Print(lang.Translatef(l, "command.freeze.unfreeze", target.Name()))
		t.Player().Message(lang.Translatef(t.Player().Locale(), "command.freeze.unfrozen"))
	} else {
		user.Alert(s, "staff.alert.freeze", target.Name())
		o.Print(lang.Translatef(l, "command.freeze.freeze", target.Name()))
		t.Player().Message(lang.Translatef(t.Player().Locale(), "command.freeze.frozen"))
	}
	t.ToggleFreeze()
}

// Allow ...
func (f Freeze) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}
