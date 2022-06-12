package module

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/vasar-network/practice/vasar/user"
	"time"
)

// Moderation is a module that handles moderation actions.
type Moderation struct {
	player.NopHandler

	u *user.User
}

// NewModeration creates a new moderation module.
func NewModeration(u *user.User) *Moderation {
	return &Moderation{u: u}
}

// HandleJoin ...
func (m *Moderation) HandleJoin() {
	for _, u := range user.All() {
		if u.Vanished() {
			m.u.Player().HideEntity(u.Player())
		}
	}
}

// HandleCommandExecution ...
func (m *Moderation) HandleCommandExecution(ctx *event.Context, _ cmd.Command, _ []string) {
	if m.u.Frozen() {
		m.u.Message("command.usage.frozen")
		ctx.Cancel()
	}
}

// HandleHurt ...
func (m *Moderation) HandleHurt(ctx *event.Context, _ *float64, _ *time.Duration, _ damage.Source) {
	if m.u.Frozen() {
		ctx.Cancel()
	}
}

// HandleItemUse ...
func (m *Moderation) HandleItemUse(ctx *event.Context) {
	if m.u.Frozen() {
		ctx.Cancel()
	}
}

// HandleAttackEntity ...
func (m *Moderation) HandleAttackEntity(ctx *event.Context, _ world.Entity, _, _ *float64, _ *bool) {
	if m.u.Frozen() {
		ctx.Cancel()
	}
}
