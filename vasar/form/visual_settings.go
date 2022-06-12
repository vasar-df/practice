package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/user"
)

// visual is the form that handles the modification of visual settings.
type visual struct {
	// Lightning is a dropdown that allows the user to enable or disable lightning.
	Lightning form.Dropdown
	// Splashes is a dropdown that allows the user to enable or disable potion splashes.
	Splashes form.Dropdown
	// PearlAnimation is a dropdown that allows the user to enable or disable the pearl animation.
	PearlAnimation form.Dropdown
	// u is the user that is using the form.
	u *user.User
}

// NewVisual creates a new form for the player to modify their visual settings.
func NewVisual(u *user.User) form.Form {
	s := u.Settings()
	return form.New(visual{
		Lightning:      newToggleDropdown("Lightning:", s.Visual.Lightning),
		Splashes:       newToggleDropdown("Potion Splashes:", s.Visual.Splashes),
		PearlAnimation: newToggleDropdown("Pearl Animation:", s.Visual.PearlAnimation),
		u:              u,
	}, "Visual Settings")
}

// Submit ...
func (d visual) Submit(form.Submitter) {
	s := d.u.Settings()
	s.Visual.Lightning = indexBool(d.Lightning)
	s.Visual.Splashes = indexBool(d.Splashes)
	s.Visual.PearlAnimation = indexBool(d.PearlAnimation)
	d.u.SetSettings(s)
	d.u.Player().SendForm(NewVisual(d.u))
}

// Close ...
func (d visual) Close(form.Submitter) {
	d.u.Player().SendForm(NewSettings(d.u))
}
