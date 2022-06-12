package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/user"
)

// gameplay is the form that handles the modification of gameplay settings.
type gameplay struct {
	// ToggleSprint is a dropdown that allows the user to toggle automatic sprinting.
	ToggleSprint form.Dropdown
	// AutoReapplyKit is a dropdown that allows the user to automatically reapply their kit after killing an entity.
	AutoReapplyKit form.Dropdown
	// PreventInterference is a dropdown that allows the user to enable or disable anti-interference.
	PreventInterference form.Dropdown
	// PreventClutter is a dropdown that allows the user to enable or disable anti-clutter.
	PreventClutter form.Dropdown
	// InstantRespawn is a dropdown that allows the user to enable or disable instant respawn.
	InstantRespawn form.Dropdown
	// u is the user that is using the form.
	u *user.User
}

// NewGameplay creates a new form for the player to modify their gameplay settings.
func NewGameplay(u *user.User) form.Form {
	s := u.Settings()
	return form.New(gameplay{
		ToggleSprint:        newToggleDropdown("Toggle Sprint:", s.Gameplay.ToggleSprint),
		AutoReapplyKit:      newToggleDropdown("Auto-Rekit:", s.Gameplay.AutoReapplyKit),
		PreventInterference: newToggleDropdown("Anti-Interference:", s.Gameplay.PreventInterference),
		PreventClutter:      newToggleDropdown("Anti-Clutter (Anti-Interference Required):", s.Gameplay.PreventClutter),
		InstantRespawn:      newToggleDropdown("Instant Respawn:", s.Gameplay.InstantRespawn),
		u:                   u,
	}, "Gameplay Settings")
}

// Submit ...
func (d gameplay) Submit(form.Submitter) {
	s := d.u.Settings()
	s.Gameplay.ToggleSprint = indexBool(d.ToggleSprint)
	s.Gameplay.AutoReapplyKit = indexBool(d.AutoReapplyKit)
	s.Gameplay.PreventInterference = indexBool(d.PreventInterference)
	s.Gameplay.PreventClutter = indexBool(d.PreventClutter)
	s.Gameplay.InstantRespawn = indexBool(d.InstantRespawn)
	d.u.SetSettings(s)
	d.u.Player().SendForm(NewGameplay(d.u))
}

// Close ...
func (d gameplay) Close(form.Submitter) {
	d.u.Player().SendForm(NewSettings(d.u))
}
