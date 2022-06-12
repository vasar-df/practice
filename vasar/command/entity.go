package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
)

// Entity is a command that spawns in a fake player entity.
type Entity struct {
	Name string `cmd:"name"`
}

// Run ...
func (e Entity) Run(s cmd.Source, _ *cmd.Output) {
	p := s.(*player.Player)
	p.World().AddEntity(player.New(e.Name, p.Skin(), p.Position()))
}

// Allow ...
func (Entity) Allow(s cmd.Source) bool {
	return allow(s, false)
}
