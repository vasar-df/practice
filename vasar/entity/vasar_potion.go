package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
	"time"
)

// PotionHandler ...
type PotionHandler interface {
	// HandleSplash handles when a player is affected by a potion splash.
	HandleSplash(*event.Context, *player.Player, *player.Player)
	// HandleParticle handles when the potion particle is displayed.
	HandleParticle(*event.Context, *player.Player)
	// HandleParticleColour handles the potion particle colour shown on splash.
	HandleParticleColour(*event.Context, *player.Player, *color.RGBA)
}

// NopPotionHandler ...
type NopPotionHandler struct{}

func (NopPotionHandler) HandleSplash(*event.Context, *player.Player, *player.Player)      {}
func (NopPotionHandler) HandleParticle(*event.Context, *player.Player)                    {}
func (NopPotionHandler) HandleParticleColour(*event.Context, *player.Player, *color.RGBA) {}

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	transform
	yaw, pitch float64

	age   int
	close bool
	d     bool
	h     PotionHandler

	owner world.Entity

	t potion.Potion
	c *entity.ProjectileComputer
}

const (
	maxDebuffHit    = 1.0993 // Original Value: 1.0393
	maxDebuffMiss   = 0.9593 // Original Value: 0.9093
	maxNoDebuffHit  = 1.0925 // Original Value: 1.0325
	maxNoDebuffMiss = 0.9525 // Original Value: 0.9025
)

// NewSplashPotion ...
func NewSplashPotion(pos, vel mgl64.Vec3, yaw, pitch float64, t potion.Potion, debuff bool, owner world.Entity) *SplashPotion {
	s := &SplashPotion{
		yaw:   yaw,
		pitch: pitch,
		owner: owner,

		t: t,
		c: &entity.ProjectileComputer{MovementComputer: &entity.MovementComputer{
			Gravity:           0.080,
			Drag:              0.0025,
			DragBeforeGravity: true,
		}},
		h: NopPotionHandler{},

		d: debuff,
	}
	s.transform = newTransform(s, pos)
	s.vel = vel
	return s
}

// Name ...
func (s *SplashPotion) Name() string {
	return "Splash Potion"
}

// EncodeEntity ...
func (s *SplashPotion) EncodeEntity() string {
	return "minecraft:splash_potion"
}

// Scale ...
func (s *SplashPotion) Scale() float64 {
	return 0.575
}

// BBox ...
func (s *SplashPotion) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Rotation ...
func (s *SplashPotion) Rotation() (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.yaw, s.pitch
}

// Type returns the type of potion the splash potion will grant effects for when thrown.
func (s *SplashPotion) Type() potion.Potion {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.t
}

// debuff returns true if the potion is a debuff potion.
func (s *SplashPotion) debuff() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.d
}

// Tick ...
func (s *SplashPotion) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.ignores)
	s.yaw, s.pitch = m.Rotation()
	s.pos, s.vel = m.Position(), m.Velocity()
	s.mu.Unlock()

	s.age++
	m.Send()

	if m.Position()[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		aabb := s.BBox().Translate(m.Position())
		colour := color.RGBA{R: 0x38, G: 0x5d, B: 0xc6, A: 0xff}
		if effects := s.t.Effects(); len(effects) > 0 {
			colour, _ = effect.ResultingColour(effects)

			debuff := s.debuff()
			expansion := s.expansion()
			ignore := func(e world.Entity) bool {
				_, living := e.(entity.Living)
				return !living || e == s
			}

			affected := make(map[entity.Living]float64)
			if entityResult, ok := result.(*trace.EntityResult); ok {
				if splashed, ok := entityResult.Entity().(entity.Living); ok {
					if debuff {
						affected[splashed] = maxDebuffHit
					} else {
						affected[splashed] = maxNoDebuffHit
					}
				}
			}

			for _, e := range w.EntitiesWithin(aabb.GrowVec3(expansion.Mul(2)), ignore) {
				pos := e.Position()
				if e.BBox().Translate(pos).IntersectsWith(aabb.GrowVec3(expansion)) {
					splashed := e.(entity.Living)
					if debuff {
						affected[splashed] = maxDebuffMiss
					} else {
						affected[splashed] = maxNoDebuffMiss
					}
				}
			}

			for splashed, potency := range affected {
				ctx := event.C()
				if owner, ok := s.owner.(*player.Player); ok {
					if splashedPlayer, ok := splashed.(*player.Player); ok {
						s.h.HandleSplash(ctx, owner, splashedPlayer)
						if ctx.Cancelled() {
							continue
						}
					}
				}
				for _, eff := range effects {
					if p, ok := eff.Type().(effect.PotentType); ok {
						splashed.AddEffect(effect.NewInstant(p.WithPotency(potency), eff.Level()))
						continue
					}

					dur := time.Duration(float64(eff.Duration()) * 0.75 * potency)
					if dur < time.Second {
						continue
					}
					splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
				}
			}
		} else if s.t == potion.Water() {
			if blockResult, ok := result.(*trace.BlockResult); ok {
				pos := blockResult.BlockPosition().Side(blockResult.Face())
				if _, ok := w.Block(pos).(block.Fire); ok {
					w.SetBlock(pos, nil, nil)
				}

				for _, f := range cube.HorizontalFaces() {
					h := pos.Side(f)
					if _, ok := w.Block(h).(block.Fire); ok {
						w.SetBlock(h, nil, nil)
					}
				}
			}
		}

		pos := m.Position()
		w.PlaySound(pos, sound.GlassBreak{})

		s.close = true

		ctx := event.C()
		if p, ok := s.owner.(*player.Player); ok {
			s.h.HandleParticleColour(ctx, p, &colour)
		}
		if ctx.Cancelled() {
			// Cancelled, don't show the particle colour to anyone.
			return
		}
		for _, e := range w.EntitiesWithin(cube.Box(-15, -15, -15, 15, 15, 15).Translate(pos), nil) {
			if p, ok := e.(*player.Player); ok {
				ctx := event.C()
				s.h.HandleParticle(ctx, p)
				if !ctx.Cancelled() {
					p.ShowParticle(pos, particle.Splash{Colour: colour})
				}
			}
		}
	}
}

// expansion returns the expansion that should be used for the bounding box.
func (s *SplashPotion) expansion() mgl64.Vec3 {
	if s.debuff() {
		return mgl64.Vec3{2.5, 3.5, 2.5}
	}
	return mgl64.Vec3{2, 3, 2}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(e world.Entity) bool {
	_, ok := e.(entity.Living)
	g, ok2 := e.(interface{ GameMode() world.GameMode })
	return !ok || e == s || (s.age < 5 && e == s.owner) || (ok2 && !g.GameMode().AllowsInteraction())
}

// Handle ...
func (s *SplashPotion) Handle(h PotionHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.h = h
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// Own ...
func (s *SplashPotion) Own(owner world.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}
