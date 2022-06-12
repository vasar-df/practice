package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/vails/lang"
)

// Spectate ...
type Spectate struct{}

// Run ...
func (s Spectate) Run(source cmd.Source, output *cmd.Output) {
	p := source.(*player.Player)
	if _, ok := lobby.LookupProvider(p); !ok {
		output.Error(lang.Translatef(p.Locale(), "user.feature.disabled"))
		return
	}
	p.SendForm(form.NewSpectate())
}

// Allow ...
func (Spectate) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
