package vasar

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"image/color"
)

// PotionHandler ...
type PotionHandler struct{}

// HandleSplash ...
func (PotionHandler) HandleSplash(ctx *event.Context, owner, splashed *player.Player) {
	if splashed == owner {
		// Don't waste time.
		return
	}
	if s, ok := user.Lookup(splashed); ok && s.Tagged() && s.Attacker() != owner {
		ctx.Cancel()
	}
}

// HandleParticle ...
func (PotionHandler) HandleParticle(ctx *event.Context, owner *player.Player) {
	if s, ok := user.Lookup(owner); ok && !s.Settings().Visual.Splashes {
		ctx.Cancel()
	}
}

// HandleParticleColour ...
func (PotionHandler) HandleParticleColour(ctx *event.Context, owner *player.Player, c *color.RGBA) {
	if u, ok := user.Lookup(owner); ok {
		switch u.Settings().Advanced.PotionSplashColor {
		case "invisible":
			ctx.Cancel()
		case "red":
			*c = color.RGBA{R: 255, A: 255}
		case "orange":
			*c = color.RGBA{R: 255, G: 155, A: 255}
		case "yellow":
			*c = color.RGBA{R: 255, G: 255, A: 255}
		case "green":
			*c = color.RGBA{G: 255, A: 255}
		case "aqua":
			*c = color.RGBA{G: 255, B: 255, A: 255}
		case "blue":
			*c = color.RGBA{R: 70, B: 255, A: 255}
		case "pink":
			*c = color.RGBA{R: 255, B: 255, A: 255}
		case "white":
			*c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case "gray":
			*c = color.RGBA{R: 155, G: 155, B: 155, A: 255}
		case "black":
			*c = color.RGBA{R: 0, G: 0, B: 0, A: 255}
		}
	}
}
