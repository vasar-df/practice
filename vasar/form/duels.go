package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"strings"
	"time"
)

// duels is a form that allows players to join games of Duels.
type duels struct {
	p match.Provider
}

// NewRankedDuels creates a new duels form with the provided host and user.
func NewRankedDuels() form.Menu {
	prov := match.Ranked()
	buttons := make([]form.Button, 0, len(game.Games()))
	for _, g := range game.Games() {
		buttons = append(buttons, form.NewButton(
			g.Name()+"\n"+fmt.Sprintf("Queued: %v Playing: %v", prov.QueuedUsers(g), prov.PlayingUsers(g)),
			g.Texture()),
		)
	}
	return form.NewMenu(duels{p: prov}, "Join Ranked Queue").WithButtons(buttons...)
}

// NewUnrankedDuels creates a new duels form with the provided host and user.
func NewUnrankedDuels() form.Menu {
	prov := match.Unranked()
	buttons := make([]form.Button, 0, len(game.Games()))
	for _, g := range game.Games() {
		buttons = append(buttons, form.NewButton(
			g.Name()+"\n"+fmt.Sprintf("Queued: %v Playing: %v", prov.QueuedUsers(g), prov.PlayingUsers(g)),
			g.Texture()),
		)
	}
	return form.NewMenu(duels{p: prov}, "Join Unranked Queue").WithButtons(buttons...)
}

// Submit ...
func (d duels) Submit(submitter form.Submitter, pressed form.Button) {
	p := submitter.(*player.Player)
	if u, ok := user.Lookup(p); ok {
		if exp := u.QueuedSince().Add(time.Second * 3); exp.After(time.Now()) {
			u.Message("queue.message.cooldown", time.Until(exp).Round(time.Millisecond*10))
			return
		}
	}
	g := game.ByName(strings.Split(pressed.Text, "\n")[0])
	d.p.EnterQueue(g, p)
}
