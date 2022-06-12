package module

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	ent "github.com/vasar-network/practice/vasar/entity"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	it "github.com/vasar-network/practice/vasar/item"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"time"
)

// Game is a module that monitors games and makes sure that players are properly removed from them on death or quit.
type Game struct {
	player.NopHandler

	u *user.User
	p *player.Player

	c chan struct{}
}

// NewGame ...
func NewGame(u *user.User) *Game {
	return &Game{u: u, p: u.Player(), c: make(chan struct{})}
}

// HandleHurt ...
func (g *Game) HandleHurt(ctx *event.Context, dmg *float64, immunity *time.Duration, s damage.Source) {
	if ctx.Cancelled() {
		// Was cancelled at some point.
		return
	}

	var attacker *player.Player
	if src, ok := s.(damage.SourceEntityAttack); ok {
		if p, ok := src.Attacker.(*player.Player); ok {
			attacker = p
		}

		if prov, ok := ffa.LookupProvider(g.p); ok && prov.Game() == game.Sumo() {
			// Damage dealt by other players in Sumo is invalidated, only in free for all.
			*dmg = 0.0
		} else if m, ok := match.Lookup(g.p); ok && (m.Game() == game.Sumo() || m.Game() == game.StickFight() || m.Game() == game.Boxing()) {
			// Damage is also invalidated in the games above, in matches only.
			*dmg = 0.0
		}
		if m, ok := match.Lookup(g.p); ok && m.Game() == game.Combo() {
			*immunity = time.Millisecond * 100
			return
		}
		*immunity = time.Millisecond * 470
	} else if src, ok := s.(damage.SourceProjectile); ok {
		if p, ok := src.Owner.(*player.Player); ok {
			attacker = p
		}
		*immunity = time.Millisecond * 200
	}

	if attacker != nil {
		if m, ok := match.Lookup(g.p); ok {
			m.LogDamage(attacker, g.p, g.p.FinalDamageFrom(*dmg, s))
			return
		}

		if u, ok := user.Lookup(attacker); ok {
			if u.Settings().Gameplay.PreventInterference && u.Tagged() && u.Attacker() != g.p {
				if u.Settings().Gameplay.PreventClutter {
					attacker.HideEntity(g.p)
					time.AfterFunc(time.Second*10, func() {
						if g.p.World() == attacker.World() {
							attacker.ShowEntity(g.p)
						}
					})
				}
				attacker.Message(lang.Translatef(attacker.Locale(), "user.interference.disabled"))
				ctx.Cancel()
				return
			}
			if g.u.Settings().Gameplay.PreventInterference && g.u.Tagged() && g.u.Attacker() != attacker {
				if g.u.Settings().Gameplay.PreventClutter {
					g.p.HideEntity(attacker)
					time.AfterFunc(time.Second*10, func() {
						if g.p.World() == attacker.World() {
							g.p.ShowEntity(attacker)
						}
					})
				}
				attacker.Message(lang.Translatef(attacker.Locale(), "target.interference.disabled"))
				ctx.Cancel()
				return
			}
		}
	}
}

// HandleMove ...
func (g *Game) HandleMove(ctx *event.Context, pos mgl64.Vec3, _ float64, _ float64) {
	_, ok := g.u.AirDuration()
	if !ok && !g.p.OnGround() {
		g.u.RenewAirDuration()
	} else if ok && g.p.OnGround() {
		g.u.ResetAirDuration()
	}

	if _, ok := lobby.LookupProvider(g.p); ok {
		if g.p.GameMode().AllowsTakingDamage() && pos.Y() <= 0 {
			lobby.Lobby().AddPlayer(g.p)
			ctx.Cancel()
		} else if g.u.Roles().Contains(role.Plus{}) {
			if dur, ok := g.u.AirDuration(); ok && dur >= time.Millisecond*900 && dur <= time.Second*3 {
				g.u.SendCustomParticle(8, 0, pos, true) // Add a flame particle.
			}
		}
	} else if f, ok := ffa.LookupProvider(g.p); ok && g.p.GameMode().AllowsTakingDamage() {
		var minY float64
		switch f.Game() {
		case game.NoDebuff():
			minY = 0
		case game.Sumo():
			minY = 150
		}
		if pos.Y() <= minY {
			g.HandleDeath(damage.SourceVoid{})
			ctx.Cancel()
		}
	} else if m, ok := match.Lookup(g.p); ok {
		if center := m.Center(); m.Spectating(g.p) && center.Sub(pos).Len() >= 50 {
			g.p.Teleport(center)
			ctx.Cancel()
		} else if m.Center().Sub(pos).Y() >= 3 {
			g.HandleDeath(damage.SourceFall{})
			ctx.Cancel()
		}
	}
}

// HandleDeath ...
func (g *Game) HandleDeath(damage.Source) {
	if m, ok := match.Lookup(g.p); ok && m.RemovePlayer(g.p, false, false) {
		// Removal was cancelled, so ignore this death.
		return
	}

	g.p.StartFlying()
	g.p.SetInvisible()

	g.p.SetGameMode(world.GameModeSpectator)
	if g.u.PearlCoolDown() {
		g.u.TogglePearlCoolDown()
	}

	g.Death(g.p.World())
	if _, ok := ffa.LookupProvider(g.p); ok {
		g.u.Board().SendScoreboard(g.p)
	}

	g.p.Inventory().Clear()
	g.p.Armour().Clear()
	g.p.SetHeldItems(item.Stack{}, item.Stack{})
	g.p.ResetFallDistance()

	for _, eff := range g.p.Effects() {
		g.p.RemoveEffect(eff.Type())
	}

	instant := g.u.Settings().Gameplay.InstantRespawn
	if instant {
		g.p.SetVisible()
		if prov, ok := ffa.LookupProvider(g.p); ok {
			prov.RemovePlayer(g.p, false)
		}
	}
	time.AfterFunc(time.Millisecond*1400, func() {
		if !instant {
			g.p.SetVisible()
			if prov, ok := ffa.LookupProvider(g.p); ok {
				prov.RemovePlayer(g.p, false)
			}
		}
	})
}

// Death processes the death of the user.
func (g *Game) Death(w *world.World) {
	pots := g.u.Potions()
	pos := g.p.Position()
	lightning := ent.NewLightning(pos)
	viewers := make([]*player.Player, 0, 16)
	for _, e := range g.p.World().EntitiesWithin(cube.Box(-50, -25, -50, 50, 25, 50).Translate(pos), nil) {
		if p, ok := e.(*player.Player); ok {
			u, ok := user.Lookup(p)
			if ok {
				if set := u.Settings(); set.Visual.Lightning && (u != g.u || !set.Gameplay.InstantRespawn) {
					u.SendSound(pos, sound.Thunder{})
					u.SendSound(pos, sound.Explosion{})
					p.ShowEntity(lightning)
					viewers = append(viewers, p)
				}
			}
		}
	}
	time.AfterFunc(time.Millisecond*250, func() {
		for _, v := range viewers {
			v.HideEntity(lightning)
		}
	})

	c := player.New(g.p.Name(), g.p.Skin(), pos)
	c.SetAttackImmunity(time.Millisecond * 1400)
	c.SetNameTag(g.p.NameTag())
	c.SetScale(g.p.Scale())
	w.AddEntity(c)

	for _, viewer := range w.Viewers(c.Position()) {
		viewer.ViewEntityAction(c, entity.DeathAction{})
	}

	if d := g.u.Attacker(); d != nil {
		src := d.Position()
		c.KnockBack(src, 0.5, 0.2)
		g.p.KnockBack(src, 0.6, 0.2)

		if da, ok := user.Lookup(d); ok {
			stats := g.u.Stats()
			stats.Deaths++
			stats.KillStreak = 0
			g.u.SetStats(stats)

			stats = da.Stats()
			stats.Kills++
			stats.KillStreak++
			if stats.KillStreak > stats.BestKillStreak {
				stats.BestKillStreak = stats.KillStreak
			}
			da.SetStats(stats)

			if prov, ok := ffa.LookupProvider(g.p); ok {
				da.Board().SendScoreboard(d)

				if prov.Game() == game.NoDebuff() {
					user.Broadcast("death.message.ffa.nodebuff", da.DisplayName(), da.Potions(), g.u.DisplayName(), pots)
				} else {
					user.Broadcast("death.message.ffa.general", da.DisplayName(), g.u.DisplayName())
				}
			}
		}
	}

	time.AfterFunc(time.Millisecond*1400, func() {
		//g.u.SendCustomParticle(18, 0, c.Position(), true)
		_ = c.Close()
	})
}

// HandleItemUse ...
func (g *Game) HandleItemUse(ctx *event.Context) {
	held, _ := g.p.HeldItems()
	if _, ok := held.Item().(item.Usable); ok && g.u.ProjectilesDisabled() {
		g.u.Message("projectiles.disabled")
		ctx.Cancel()
		return
	}
	if _, ok := held.Item().(it.VasarPearl); ok && (g.u.Player().GameMode() != world.GameModeCreative || !g.u.Roles().Contains(role.Operator{})) {
		if g.u.PearlCoolDown() {
			ctx.Cancel()
			return
		}

		g.u.TogglePearlCoolDown()
		g.p.SendTitle(title.New().WithActionText(lang.Translatef(g.p.Locale(), "pearl.cooldown.started")))

		const duration = time.Second * 15
		expected := time.Now().Add(duration)
		go func() {
			ticker := time.NewTicker(time.Millisecond * 50)
			defer ticker.Stop()
			for {
				select {
				case now := <-ticker.C:
					remaining := expected.Sub(now)
					if !g.u.PearlCoolDown() {
						g.u.ResetExperienceProgress()
						return
					}
					if now.After(expected) {
						g.u.TogglePearlCoolDown()
						g.p.SendTitle(title.New().WithActionText(lang.Translatef(g.p.Locale(), "pearl.cooldown.ended")))
						return
					}

					level := int(remaining.Seconds())
					progress := float64(remaining.Milliseconds()) / float64(duration.Milliseconds())
					g.u.SendExperienceProgress(level, progress)
				case <-g.c:
					// User was closed, so return.
					return
				}
			}
		}()
	}
}

// HandleAttackEntity ...
func (g *Game) HandleAttackEntity(ctx *event.Context, e world.Entity, force, height *float64, _ *bool) {
	if ctx.Cancelled() {
		return
	}
	ga := game.NoDebuff() // Default to NoDebuff.
	if prov, ok := ffa.LookupProvider(g.p); ok {
		if !prov.PvP() {
			g.u.Message("pvp.disabled.arena")
			ctx.Cancel()
			return
		}
	} else if m, ok := match.Lookup(g.p); ok {
		ga = m.Game()
	}

	switch ga {
	case game.StickFight():
		*force, *height = 0.630, 0.383
	case game.Sumo():
		*force, *height = 0.394, 0.399
	case game.BuildUHC():
		*force, *height = 0.394, 0.394
	case game.Boxing():
		*force, *height = 0.383, 0.387
	case game.Soup():
		*force, *height = 0.39, 0.39
	default:
		*force, *height = 0.394, 0.394 // Original Value(s): 0.390, 0.394
	}

	if o, ok := e.(interface{ OnGround() bool }); ok && !o.OnGround() {
		if dist := e.Position().Y() - g.p.Position().Y(); dist >= 3 {
			*height -= dist / 28.795
		}
	}

	if m, ok := match.Lookup(g.p); ok {
		if p, ok := e.(*player.Player); ok && !p.AttackImmune() {
			m.LogHit(g.p, p)
		}
	}
}

// HandleItemDrop ...
func (g *Game) HandleItemDrop(ctx *event.Context, e *entity.Item) {
	if _, ok := e.Item().Item().(item.GlassBottle); ok {
		// TODO: Remove this hack!
		time.AfterFunc(time.Millisecond*50, func() {
			_ = e.Close()
		})
		return
	}
	ctx.Cancel()
}

// HandleQuit ...
func (g *Game) HandleQuit() {
	match.Unranked().RemoveRequestTo(g.p)
	match.Unranked().RemoveRequestsFrom(g.p)
	if m, ok := match.Lookup(g.p); ok {
		m.RemovePlayer(g.p, true, false)
	}
	if m, ok := match.Spectating(g.p); ok {
		m.RemoveSpectator(g.p, true)
	}
	if prov, ok := match.LookupProvider(g.p); ok {
		prov.ExitQueue(g.p)
	}
	if prov, ok := ffa.LookupProvider(g.p); ok {
		prov.RemovePlayer(g.p, true)
	}
	if prov, ok := lobby.LookupProvider(g.p); ok {
		prov.RemovePlayer(g.p, true)
	}
	close(g.c)
}

// HandleItemDamage ...
func (g *Game) HandleItemDamage(ctx *event.Context, _ item.Stack, _ int) {
	if _, ok := ffa.LookupProvider(g.p); ok {
		ctx.Cancel()
	}
}
