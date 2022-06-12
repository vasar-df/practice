package module

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/user"
	"sync"
	"time"
)

// Click is a handler that is used to track player clicks for CPS.
type Click struct {
	player.NopHandler

	u *user.User

	clicks  []time.Time
	clickMu sync.Mutex
}

// NewClick ...
func NewClick(u *user.User) *Click {
	return &Click{u: u}
}

// HandlePunchAir ...
func (c *Click) HandlePunchAir(*event.Context) {
	c.Click()
}

// HandleAttackEntity ...
func (c *Click) HandleAttackEntity(*event.Context, world.Entity, *float64, *float64, *bool) {
	c.Click()
}

// Click adds a click to the click history.
func (c *Click) Click() {
	c.clickMu.Lock()
	c.clicks = append(c.clicks, time.Now())
	if len(c.clicks) >= 100 {
		c.clicks = c.clicks[1:]
	}
	cps := c.calculate()
	c.clickMu.Unlock()
	for _, w := range c.u.ClickWatchers() {
		w.Player().SendTip(text.Colourf("<white>%v CPS</white>", cps))
	}
	if _, ok := c.u.WatchingClicks(); !ok && c.u.Settings().Display.CPS {
		c.u.Player().SendTip(text.Colourf("<white>%v CPS</white>", cps))
	}
}

// CPS returns the current clicks per second.
func (c *Click) CPS() int {
	c.clickMu.Lock()
	defer c.clickMu.Unlock()
	return c.calculate()
}

// calculate uses the click samples to calculate the player's current clicks per second. This does not lock the click
// mutex, as it is expected to only be called by CPS and Click.
func (c *Click) calculate() int {
	var clicks int
	for _, past := range c.clicks {
		if time.Since(past) <= time.Second {
			clicks++
		}
	}
	return clicks
}
