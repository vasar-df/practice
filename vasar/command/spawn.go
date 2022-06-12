package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/vails/lang"
)

// Spawn is a command that teleports the player to the spawn.
type Spawn struct{}

// Run ...
func (Spawn) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	p := s.(*player.Player)
	if prov, ok := ffa.LookupProvider(p); ok {
		prov.RemovePlayer(p, false)
		return
	} else if m, ok := match.Lookup(p); ok {
		if !m.RemovePlayer(p, true, true) {
			o.Print(lang.Translatef(l, "user.feature.disabled"))
		}
		return
	} else if m, ok := match.Spectating(p); ok {
		m.RemoveSpectator(p, false)
		return
	}
	lobby.Lobby().AddPlayer(p)
}

// Allow ...
func (Spawn) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
