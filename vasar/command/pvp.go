package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/practice/vasar"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/user"
)

// PvP is a command that lets a user enable or disable pvp in any FFA arena.
type PvP struct {
	v    *vasar.Vasar
	Game gameType `cmd:"game"`
}

// NewPvP ...
func NewPvP(v *vasar.Vasar) PvP {
	return PvP{v: v}
}

// Run ...
func (p PvP) Run(s cmd.Source, _ *cmd.Output) {
	if string(p.Game) == "global" {
		if !p.v.TogglePvP() {
			user.Broadcast("command.pvp.enable.global", s.Name())
			return
		}
		user.Broadcast("command.pvp.disable.global", s.Name())
		return
	}
	g := game.ByString(string(p.Game))
	for _, p := range ffa.Providers() {
		if p.Game() == g {
			if !p.TogglePvP() {
				user.Broadcast("command.pvp.enable.mode", s.Name(), g.Name())
				return
			}
			user.Broadcast("command.pvp.disable.mode", s.Name(), g.Name())
			return
		}
	}
}

// gameType ...
type gameType string

// Type ...
func (gameType) Type() string { return "game" }

// Options ...
func (gameType) Options(cmd.Source) []string {
	options := []string{"global"}
	for _, g := range game.FFA() {
		options = append(options, g.String())
	}
	return options
}

// Allow ...
func (PvP) Allow(s cmd.Source) bool {
	return allow(s, true)
}
