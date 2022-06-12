package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
)

// Stats is a command that displays the stats of a player.
type Stats struct {
	Targets cmd.Optional[[]cmd.Target] `cmd:"target"`
}

// StatsOffline is a command that displays the stats of an offline player.
type StatsOffline struct {
	Target cmd.Optional[string] `cmd:"target"`
}

// Run ...
func (st Stats) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	targets := st.Targets.LoadOr(nil)
	if len(targets) <= 0 {
		f, _ := form.NewCasualStats(p, p.XUID())
		p.SendForm(f)
		return
	}
	if len(targets) > 1 {
		o.Error(lang.Translatef(p.Locale(), "command.targets.exceed"))
		return
	}
	t, ok := user.Lookup(targets[0].(*player.Player))
	if !ok {
		o.Error(lang.Translatef(p.Locale(), "command.target.unknown"))
		return
	}
	if !t.Settings().Privacy.PublicStatistics {
		o.Error(lang.Translatef(p.Locale(), "command.stats.private"))
		return
	}
	f, _ := form.NewCasualStats(p, t.Player().XUID())
	p.SendForm(f)
}

// Run ...
func (st StatsOffline) Run(s cmd.Source, o *cmd.Output) {
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// The user somehow left in the middle of this, so just stop in our tracks.
		return
	}
	target, hasTarget := st.Target.Load()
	if !hasTarget {
		o.Error(lang.Translatef(u.Player().Locale(), "command.target.unknown"))
		return
	}
	t, err := data.LoadOfflineUser(target)
	if err != nil {
		o.Error(lang.Translatef(u.Player().Locale(), "command.target.unknown"))
		return
	}
	f, err := form.NewCasualStats(u.Player(), t.XUID())
	if err != nil {
		o.Error(err)
		return
	}
	u.Player().SendForm(f)
}

// Allow ...
func (Stats) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

// Allow ...
func (StatsOffline) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
