package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// FishingHook ...
type FishingHook struct {
	transform

	age int

	owner *player.Player

	close bool

	c *entity.ProjectileComputer
}

// NewFishingHook ...
func NewFishingHook(pos, vel mgl64.Vec3, owner *player.Player) *FishingHook {
	f := &FishingHook{
		owner: owner,
		c: &entity.ProjectileComputer{MovementComputer: &entity.MovementComputer{
			Gravity:           0.1,
			Drag:              0.02,
			DragBeforeGravity: true,
		}},
	}
	f.transform = newTransform(f, pos)
	f.vel = vel.Normalize().Add(mgl64.Vec3{
		rand.Float64(),
		rand.Float64(),
		rand.Float64(),
	}.Mul(0.007499999832361937)).Mul(1.3)
	f.vel[0] += vel[0]
	f.vel[2] += vel[2]
	return f
}

// Tick ...
func (f *FishingHook) Tick(w *world.World, current int64) {
	if f.close {
		_ = f.Close()
		return
	}

	held, _ := f.owner.HeldItems()
	if r, ok := held.Item().(interface {
		Rod() bool
	}); !ok || !r.Rod() {
		f.close = true
		return
	}

	f.mu.Lock()
	vel := f.vel
	m, result := f.c.TickMovement(f, f.pos, f.vel, 0, 0, f.ignores)
	f.pos, f.vel = m.Position(), m.Velocity()

	f.age++
	f.mu.Unlock()

	m.Send()

	if m.Position()[1] < float64(w.Range()[0]) && current%10 == 0 {
		f.close = true
		return
	}

	if result != nil {
		if res, ok := result.(trace.EntityResult); ok {
			if l, ok := res.Entity().(entity.Living); ok && !l.AttackImmune() {
				if _, vulnerable := l.Hurt(0.0, damage.SourceProjectile{Projectile: f, Owner: f.Owner()}); vulnerable {
					if entity.DirectionVector(f.owner).Dot(entity.DirectionVector(l)) > 0 {
						// Pull back the target.
						l.KnockBack(l.Position().Add(vel), 0.230, 0.372)
					} else {
						// Push back the target.
						l.KnockBack(l.Position().Sub(vel), 0.374, 0.372)
					}
				}
			}
		}
		f.close = true
	}
}

// Name ...
func (f *FishingHook) Name() string {
	return "Fishing Hook"
}

// EncodeEntity ...
func (f *FishingHook) EncodeEntity() string {
	return "minecraft:fishing_hook"
}

// BBox ...
func (f *FishingHook) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Owner ...
func (f *FishingHook) Owner() world.Entity {
	return f.owner
}

// ignores returns whether the arrow should ignore collision with the entity passed.
func (f *FishingHook) ignores(e world.Entity) bool {
	_, ok := e.(entity.Living)
	g, ok2 := e.(interface{ GameMode() world.GameMode })
	return !ok || e == f || (f.age < 5 && e == f.owner) || (ok2 && !g.GameMode().AllowsInteraction())
}
