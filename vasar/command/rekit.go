package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/vails/lang"
)

// Rekit is a command that gives the user the kit of the FFA game they are currently in.
type Rekit struct{}

// Run ...
func (Rekit) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	if prov, ok := ffa.LookupProvider(p); ok {
		kit.Apply(prov.Game().Kit(true), p)
		return
	}
	o.Error(lang.Translatef(p.Locale(), "user.feature.disabled"))
}

// Allow ...
func (Rekit) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
