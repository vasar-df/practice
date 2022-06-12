package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/vails/lang"
)

// FFA is a command that allows an operator to enable or disable an FFA arena.
type FFA struct {
	Arena arena `cmd:"arena"`
}

// Run ...
func (f FFA) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	for _, prov := range ffa.Providers() {
		if prov.Game().String() == string(f.Arena) {
			if !prov.ToggleStatus() {
				o.Print(lang.Translatef(l, "command.arena.opened", prov.Game().Name()))
				return
			}
			for _, p := range prov.Players() {
				prov.RemovePlayer(p, false)
			}
			o.Print(lang.Translatef(l, "command.arena.closed", prov.Game().Name()))
			break
		}
	}
}

// Allow ...
func (FFA) Allow(s cmd.Source) bool {
	return allow(s, true)
}

type (
	arena string
)

// Type ...
func (arena) Type() string {
	return "arena"
}

// Options ...
func (arena) Options(cmd.Source) (arenas []string) {
	for _, g := range game.FFA() {
		arenas = append(arenas, g.String())
	}
	return
}
