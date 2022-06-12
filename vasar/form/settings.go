package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/user"
)

// settings is a form menu which contains all possible player setting categories, allowing the user to configure them
// at will.
type settings struct {
	u *user.User
}

// NewSettings creates a new settings form to send to a player.
func NewSettings(u *user.User) form.Menu {
	return form.NewMenu(settings{u: u}, text.Colourf("Settings")).WithButtons(
		form.NewButton("Display", ""),
		form.NewButton("Visual", ""),
		form.NewButton("Gameplay", ""),
		form.NewButton("Privacy", ""),
		form.NewButton("Matchmaking", ""),
		form.NewButton("Advanced", ""),
	)
}

// Submit ...
func (s settings) Submit(_ form.Submitter, pressed form.Button) {
	p := s.u.Player()
	switch pressed.Text {
	case "Display":
		p.SendForm(NewDisplay(s.u))
	case "Visual":
		p.SendForm(NewVisual(s.u))
	case "Gameplay":
		p.SendForm(NewGameplay(s.u))
	case "Privacy":
		p.SendForm(NewPrivacy(s.u))
	case "Matchmaking":
		p.SendForm(NewMatchmaking(s.u))
	case "Advanced":
		p.SendForm(NewAdvanced(s.u))
	}
}
