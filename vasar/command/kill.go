package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
)

// Kill is a command that allows the player to kill themselves or another player if they have the permission.
type Kill struct {
	Targets cmd.Optional[[]cmd.Target] `cmd:"target"`
}

// suicide is a custom damage source for suicidal people.
type suicide struct{}

// ReducedByArmour ...
func (s suicide) ReducedByArmour() bool {
	return false
}

// ReducedByResistance ...
func (s suicide) ReducedByResistance() bool {
	return false
}

// Run ...
func (k Kill) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	l := locale(p)
	if targets := k.Targets.LoadOr(nil); len(targets) > 0 {
		u, ok := user.Lookup(p)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		if !u.Roles().Contains(role.Operator{}) {
			o.Error(lang.Translatef(l, "command.kill.disabled"))
			return
		}
		for _, tI := range targets {
			t, ok := tI.(*player.Player)
			if !ok {
				o.Error(lang.Translatef(l, "command.target.unknown"))
				return
			}
			t.Hurt(t.MaxHealth(), suicide{})
		}
	} else {
		p.Hurt(p.MaxHealth(), suicide{})
	}
}

// Allow ...
func (Kill) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
