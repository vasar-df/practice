package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// Who is a command that displays information about a specific player.
type Who struct {
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (w Who) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(w.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	if target, ok := w.Targets[0].(*player.Player); ok {
		op := true
		if p, ok := s.(*player.Player); ok {
			if u, ok := user.Lookup(p); ok {
				op = u.Roles().Contains(role.Operator{})
			}
		}

		t, ok := user.Lookup(target)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		if op {
			o.Print(lang.Translatef(l, "command.who.op",
				t.Player().Name(),
				t.Device(),
				t.DeviceGroup(),
				t.Latency(),
				t.DeviceID(),
				t.SelfSignedID(),
			))
		} else {
			o.Print(lang.Translatef(l, "command.who.staff",
				t.Player().Name(),
				t.Device(),
				t.DeviceGroup(),
				t.Latency(),
			))
		}
	}
}

// Allow ...
func (Who) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}
