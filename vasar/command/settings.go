package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/user"
)

// Settings is a command that sends the settings form to the user.
type Settings struct{}

// Run ...
func (r Settings) Run(s cmd.Source, _ *cmd.Output) {
	if u, ok := user.Lookup(s.(*player.Player)); ok {
		u.Player().SendForm(form.NewSettings(u))
	}
}

// Allow ...
func (Settings) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
