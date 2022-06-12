package module

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/role"
	"golang.org/x/exp/slices"
	"time"
)

// Combat is a handler that is used for combat related things, such as the combat tag.
type Combat struct {
	player.NopHandler

	u *user.User
}

// bannedCommands is a list of commands disallowed in combat.
var bannedCommands = []string{"spawn", "rekit"}

// NewCombat ...
func NewCombat(u *user.User) *Combat {
	return &Combat{u: u}
}

// HandleCommandExecution ...
func (c *Combat) HandleCommandExecution(ctx *event.Context, cmd cmd.Command, _ []string) {
	if c.u.Tagged() && !c.u.Roles().Contains(role.Operator{}) && slices.Contains(bannedCommands, cmd.Name()) {
		c.u.Message("user.feature.disabled")
		ctx.Cancel()
	}
}

// HandleHurt ...
func (c *Combat) HandleHurt(ctx *event.Context, _ *float64, _ *time.Duration, s damage.Source) {
	if ctx.Cancelled() {
		// Was cancelled at some point, so just ignore this.
		return
	}

	var attacker *player.Player
	if a, ok := s.(damage.SourceEntityAttack); ok {
		if p, ok := a.Attacker.(*player.Player); ok {
			attacker = p
		}
	} else if t, ok := s.(damage.SourceProjectile); ok {
		if p, ok := t.Owner.(*player.Player); ok {
			attacker = p
		}
	}
	if attacker == nil {
		// No attacker, so we don't need to do anything.
		return
	}

	if a, ok := user.Lookup(attacker); ok {
		_, notify := ffa.LookupProvider(a.Player())
		a.Tag(c.u.Player(), false, notify)

		_, notify = ffa.LookupProvider(c.u.Player())
		c.u.Tag(attacker, false, notify)
	}
}

// HandleDeath ...
func (c *Combat) HandleDeath(s damage.Source) {
	if c.u.Tagged() {
		c.u.RemoveTag()
	}
	src, ok := s.(damage.SourceEntityAttack)
	if !ok {
		// Not an entity attack, so we don't care.
		return
	}
	if p, ok := src.Attacker.(*player.Player); ok {
		if u, ok := user.Lookup(p); ok {
			_, notify := ffa.LookupProvider(p)
			u.Tag(nil, true, notify)
		}
	}
}
