package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/user"
)

// privacy is the form that handles the modification of privacy settings.
type privacy struct {
	// PrivateMessages is a dropdown that allows the user to enable or disable private messages from others.
	PrivateMessages form.Dropdown
	// PublicStatistics is a dropdown that allows the user to enable or disable public statistics.
	PublicStatistics form.Dropdown
	// DuelRequests is a dropdown that allows the user to enable or disable duel requests from others.
	DuelRequests form.Dropdown
	// u is the user that is using the form.
	u *user.User
}

// NewPrivacy creates a new form for the player to modify their privacy settings.
func NewPrivacy(u *user.User) form.Form {
	s := u.Settings()
	return form.New(privacy{
		PrivateMessages:  newToggleDropdown("Allow others to private message me:", s.Privacy.PrivateMessages),
		PublicStatistics: newToggleDropdown("Allow others to view my stats:", s.Privacy.PublicStatistics),
		DuelRequests:     newToggleDropdown("Allow others to send me duel requests:", s.Privacy.DuelRequests),
		u:                u,
	}, "Privacy Settings")
}

// Submit ...
func (d privacy) Submit(form.Submitter) {
	s := d.u.Settings()
	s.Privacy.PrivateMessages = indexBool(d.PrivateMessages)
	s.Privacy.PublicStatistics = indexBool(d.PublicStatistics)
	s.Privacy.DuelRequests = indexBool(d.DuelRequests)
	d.u.SetSettings(s)
	d.u.Player().SendForm(NewPrivacy(d.u))
}

// Close ...
func (d privacy) Close(form.Submitter) {
	d.u.Player().SendForm(NewSettings(d.u))
}
