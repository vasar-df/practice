package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/game/ffa"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/vails/lang"
	"strings"
)

// freeForAll is a form that allows players to join games of Free For All.
type freeForAll struct{}

// NewFFA creates a new FFA form with the provided host and user.
func NewFFA() form.Menu {
	var buttons []form.Button
	for _, prov := range ffa.Providers() {
		status := text.Colourf("<red>Closed</red>")
		if prov.Open() {
			status = fmt.Sprintf("%v/%v playing", len(prov.Players()), prov.Game().Cap())
		}
		buttons = append(buttons, form.NewButton(
			prov.Game().Name()+"\n"+status,
			prov.Game().Texture()),
		)
	}
	return form.NewMenu(freeForAll{}, "Free For All").WithButtons(buttons...)
}

// Submit ...
func (f freeForAll) Submit(submitter form.Submitter, pressed form.Button) {
	p := submitter.(*player.Player)
	if _, ok := lobby.LookupProvider(p); !ok {
		p.Message(lang.Translatef(p.Locale(), "user.feature.disabled"))
		return
	}
	g := game.ByName(strings.Split(pressed.Text, "\n")[0])
	for _, prov := range ffa.Providers() {
		if prov.Game() == g {
			if !prov.Open() {
				p.Message(lang.Translatef(p.Locale(), "arena.closed"))
				return
			}
			if len(prov.Players()) >= prov.Game().Cap() {
				p.Message(lang.Translatef(p.Locale(), "arena.full"))
				return
			}
			lobby.Lobby().RemovePlayer(p, false)
			prov.AddPlayer(p)
			break
		}
	}
}
