package module

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/healing"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/cape"
	"math"
)

// Settings is a module which handles player settings.
type Settings struct {
	player.NopHandler

	u *user.User
}

// NewSettings ...
func NewSettings(u *user.User) *Settings {
	return &Settings{u: u}
}

// HandleJoin ...
func (s *Settings) HandleJoin() {
	c, _ := cape.ByName(s.u.Settings().Advanced.Cape)
	sk := s.u.Player().Skin()
	sk.Cape = c.Cape()
	s.u.Player().SetSkin(sk)
}

// HandleSkinChange ...
func (s *Settings) HandleSkinChange(_ *event.Context, sk *skin.Skin) {
	c, _ := cape.ByName(s.u.Settings().Advanced.Cape)
	(*sk).Cape = c.Cape()
}

// HandleMove ...
func (s *Settings) HandleMove(_ *event.Context, pos mgl64.Vec3, newYaw, _ float64) {
	p := s.u.Player()
	if !s.u.Settings().Gameplay.ToggleSprint || p.Sprinting() {
		return
	}
	delta := pos.Sub(p.Position())
	if mgl64.FloatEqual(delta[0], 0) && mgl64.FloatEqual(delta[2], 0) {
		return
	}
	diff := (mgl64.RadToDeg(math.Atan2(delta[2], delta[0])) - 90) - newYaw
	if diff < 0 {
		diff += 360
	}
	if diff <= 65 && diff >= -65 {
		p.StartSprinting()
	}
}

// HandleDeath ...
func (s *Settings) HandleDeath(src damage.Source) {
	if source, ok := src.(damage.SourceEntityAttack); ok {
		if p, ok := source.Attacker.(*player.Player); ok {
			u, ok := user.Lookup(p)
			if !ok {
				return
			}
			if prov, ok := ffa.LookupProvider(p); ok {
				u.Player().Heal(20, healing.SourceKill{})
				if u.Settings().Gameplay.AutoReapplyKit {
					kit.Apply(prov.Game().Kit(true), p)
				}
			}
		}
	}
}

// HandleAttackEntity ...
func (s *Settings) HandleAttackEntity(ctx *event.Context, e world.Entity, _, _ *float64, _ *bool) {
	if ctx.Cancelled() {
		return
	}
	if e, ok := e.(*player.Player); ok && !e.AttackImmune() {
		s.u.MultiplyParticles(e, s.u.Settings().Advanced.ParticleMultiplier)
	}
}
