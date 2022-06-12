package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/user"
)

// display is the form that handles the modification of display settings.
type display struct {
	// Scoreboard is a dropdown that allows the user to enable or disable scoreboards.
	Scoreboard form.Dropdown
	// CPS is a dropdown that allows the user to enable or disable the CPS counter.
	CPS form.Dropdown
	// u is the user that is using the form.
	u *user.User
}

// NewDisplay creates a new form for the player to modify their display settings.
func NewDisplay(u *user.User) form.Form {
	s := u.Settings()
	return form.New(display{
		Scoreboard: newToggleDropdown("Scoreboard:", s.Display.Scoreboard),
		CPS:        newToggleDropdown("CPS Counter:", s.Display.CPS),
		u:          u,
	}, "Display Settings")
}

// Submit ...
func (d display) Submit(form.Submitter) {
	s := d.u.Settings()
	s.Display.CPS = indexBool(d.CPS)
	s.Display.Scoreboard = indexBool(d.Scoreboard)
	d.u.SetSettings(s)
	if s.Display.Scoreboard {
		d.u.Board().SendScoreboard(d.u.Player())
	} else if !s.Display.Scoreboard {
		d.u.Player().RemoveScoreboard()
	}
	d.u.Player().SendForm(NewDisplay(d.u))
}

// Close ...
func (d display) Close(form.Submitter) {
	d.u.Player().SendForm(NewSettings(d.u))
}
