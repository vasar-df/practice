package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// CPS is a command to start watching the CPS of a user.
type CPS struct {
	Targets cmd.Optional[[]cmd.Target] `cmd:"target"`
}

// Run ...
func (c CPS) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// The user somehow left in the middle of this, so just stop in our tracks.
		return
	}
	targets := c.Targets.LoadOr(nil)
	if len(targets) == 0 {
		user.Alert(s, "staff.alert.cps.off")
		o.Print(lang.Translatef(l, "command.cps.stop"))
		u.StopWatchingClicks()
		return
	}
	if len(targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	target, ok := targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	t, ok := user.Lookup(target)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if role.Staff(t.Roles().Highest()) {
		o.Error(lang.Translatef(l, "command.cps.staff"))
		return
	}
	if watcher, ok := u.WatchingClicks(); ok && watcher == t {
		o.Error(lang.Translatef(l, "command.cps.already"))
		return
	}
	user.Alert(s, "staff.alert.cps.on", t.Player().Name())
	o.Print(lang.Translatef(l, "command.cps.start", target.Name()))
	u.StartWatchingClicks(t)
}

// Allow ...
func (c CPS) Allow(s cmd.Source) bool {
	return allow(s, false, role.Trial{})
}
