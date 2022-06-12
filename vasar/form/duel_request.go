package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vap"
	"strings"
)

// duelRequest ...
type duelRequest struct {
	to *user.User
}

// NewDuelRequest ...
func NewDuelRequest(t *user.User) form.Menu {
	buttons := make([]form.Button, 0, len(game.Games()))
	for _, g := range game.Games() {
		buttons = append(buttons, form.NewButton(g.Name(), g.Texture()))
	}
	return form.NewMenu(duelRequest{to: t}, fmt.Sprintf("Duel %v", t.DisplayName())).WithButtons(buttons...)
}

// Submit ...
func (d duelRequest) Submit(submitter form.Submitter, pressed form.Button) {
	g := game.ByName(strings.Split(pressed.Text, "\n")[0])
	submitter.SendForm(newDuelMaps(g, d.to))
}

// duelMaps ...
type duelMaps struct {
	g  game.Game
	to *user.User

	m map[form.Button]*vap.Vap
}

// newDuelMaps ...
func newDuelMaps(g game.Game, t *user.User) form.Menu {
	a := match.Arenas(g)
	m := make(map[form.Button]*vap.Vap, len(a))
	buttons := make([]form.Button, 0, len(a))
	buttons = append(buttons, form.NewButton("Random", ""))
	for _, v := range a {
		n, _, _ := v.Arena()
		b := form.NewButton(n, "")
		m[b] = v
		buttons = append(buttons, b)
	}
	return form.NewMenu(duelMaps{g: g, to: t, m: m}, "Pick a Map").WithButtons(buttons...)
}

// Submit ...
func (d duelMaps) Submit(submitter form.Submitter, pressed form.Button) {
	p := submitter.(*player.Player)
	if _, ok := lobby.LookupProvider(p); !ok {
		p.Message(lang.Translatef(p.Locale(), "user.feature.disabled"))
		return
	}

	var v *vap.Vap
	if pressed.Text == "Random" {
		v = match.RandomArena(d.g)
	} else {
		v = d.m[pressed]
	}

	match.Unranked().RequestDuel(d.g, v, p, d.to.Player())
}
