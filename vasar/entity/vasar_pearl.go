package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
	_ "unsafe"
)

// PearlHandler ...
type PearlHandler interface {
	// HandleTeleport ...
	HandleTeleport(*event.Context, *player.Player, mgl64.Vec3)
}

// NopPearlHandler ...
type NopPearlHandler struct{}

func (NopPearlHandler) HandleTeleport(*event.Context, *player.Player, mgl64.Vec3) {}

// VasarPearl is a copy of an ender pearl with some edits.
type VasarPearl struct {
	transform
	yaw, pitch float64

	age   int
	close bool
	h     PearlHandler

	owner world.Entity

	c *entity.ProjectileComputer
}

// NewEnderPearl ...
func NewEnderPearl(pos, vel mgl64.Vec3, yaw, pitch float64, owner world.Entity) *VasarPearl {
	e := &VasarPearl{
		yaw:   yaw,
		pitch: pitch,
		c: &entity.ProjectileComputer{MovementComputer: &entity.MovementComputer{
			Gravity:           0.065,
			Drag:              0.0025,
			DragBeforeGravity: true,
		}},
		owner: owner,
		h:     NopPearlHandler{},
	}
	e.transform = newTransform(e, pos)
	e.vel = vel
	return e
}

// Name ...
func (e *VasarPearl) Name() string {
	return "Ender Pearl"
}

// EncodeEntity ...
func (e *VasarPearl) EncodeEntity() string {
	return "minecraft:ender_pearl"
}

// Scale ...
func (e *VasarPearl) Scale() float64 {
	return 0.575
}

// BBox ...
func (e *VasarPearl) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Rotation ...
func (e *VasarPearl) Rotation() (float64, float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.yaw, e.pitch
}

// Tick ...
func (e *VasarPearl) Tick(w *world.World, current int64) {
	if e.close {
		_ = e.Close()
		return
	}
	e.mu.Lock()
	m, result := e.c.TickMovement(e, e.pos, e.vel, e.yaw, e.pitch, e.ignores)
	e.yaw, e.pitch = m.Rotation()
	e.pos, e.vel = m.Position(), m.Velocity()
	h := e.h
	e.mu.Unlock()

	e.age++
	m.Send()

	if m.Position()[1] < float64(w.Range()[0]) && current%10 == 0 {
		e.close = true
		return
	}

	if result != nil {
		var isEntity bool
		if r, ok := result.(trace.EntityResult); ok {
			if l, ok := r.Entity().(entity.Living); ok {
				isEntity = ok
				if _, vulnerable := l.Hurt(0.0, damage.SourceProjectile{Projectile: e, Owner: e.Owner()}); vulnerable {
					l.KnockBack(m.Position(), 0.435, 0.355)
				}
			}
		}

		if owner := e.Owner(); owner != nil {
			if p, ok := owner.(*player.Player); ok {
				pos := p.Position()
				w.PlaySound(pos, sound.Teleport{})

				ctx := event.C()
				h.HandleTeleport(ctx, p, m.Position())
				if !ctx.Cancelled() {
					session_ViewEntityTeleport(player_session(p), owner, m.Position())
					p.Move(m.Position().Sub(pos), 0, 0)
				}

				w.AddParticle(m.Position(), particle.EndermanTeleportParticle{})
				w.PlaySound(m.Position(), sound.Teleport{})

				p.Hurt(0, damage.SourceFall{})

				if isEntity {
					p.SetAttackImmunity(245 * time.Millisecond)
				}
			}
		}

		e.close = true
	}
}

// ignores returns whether the ender pearl should ignore collision with the entity passed.
func (e *VasarPearl) ignores(otherEntity world.Entity) bool {
	_, ok := otherEntity.(entity.Living)
	g, ok2 := otherEntity.(interface{ GameMode() world.GameMode })
	return !ok || otherEntity == e || (e.age < 5 && otherEntity == e.owner) || (ok2 && !g.GameMode().AllowsInteraction())
}

// Handle ...
func (e *VasarPearl) Handle(h PearlHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.h = h
}

// Owner ...
func (e *VasarPearl) Owner() world.Entity {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.owner
}

// Own ...
func (e *VasarPearl) Own(owner world.Entity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.owner = owner
}

//go:linkname player_session github.com/df-mc/dragonfly/server/player.(*Player).session
//noinspection ALL
func player_session(*player.Player) *session.Session

//go:linkname session_ViewEntityTeleport github.com/df-mc/dragonfly/server/session.(*Session).ViewEntityTeleport
//noinspection ALL
func session_ViewEntityTeleport(*session.Session, world.Entity, mgl64.Vec3)
