package form

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/role"
)

// matchmaking is the form that handles the modification of matchmaking settings.
type matchmaking struct {
	// MatchWithMobile is a toggle allowing the user to enable or disable matching with mobile players.
	MatchWithMobile form.Toggle
	// MatchWithController is a toggle allowing the user to enable or disable matching with controller players.
	MatchWithController form.Toggle
	// MatchWithKeyboard is a toggle allowing the user to enable or disable matching with keyboard players.
	MatchWithKeyboard form.Toggle
	// PingRange is a slider allowing the user to set the ping range opponents must fall within.
	PingRange form.StepSlider
	// u is the user that is using the form.
	u *user.User
}

// NewMatchmaking creates a new form for the player to modify their matchmaking settings.
func NewMatchmaking(u *user.User) form.Form {
	s := u.Settings()
	return form.New(matchmaking{
		MatchWithMobile:     form.NewToggle("Queue Against Touch Players", s.Matchmaking.MatchWithMobile),
		MatchWithController: form.NewToggle("Queue Against Controller Players", s.Matchmaking.MatchWithController),
		MatchWithKeyboard:   form.NewToggle("Queue Against Mouse/Keyboard Players", s.Matchmaking.MatchWithKeyboard),
		PingRange: form.NewStepSlider("Ping Range", []string{
			text.Colourf("<red>Unrestricted</red>"),
			"25",
			"50",
			"75",
			"100",
			"125",
			"150",
		}, int(s.Matchmaking.PingRange)),
		u: u,
	}, "Matchmaking Settings")
}

// Submit ...
func (d matchmaking) Submit(form.Submitter) {
	s := d.u.Settings()
	p := d.u.Player()
	if !d.u.Roles().Contains(role.Plus{}) && !d.u.Roles().Contains(role.Operator{}) {
		d.u.Message("setting.plus")
		return
	}
	if !d.MatchWithMobile.Value() && d.u.DeviceGroup().Compare(user.DeviceGroupMobile()) {
		d.u.Message("form.matchmaking.forced.queueing")
		p.SendForm(NewMatchmaking(d.u))
		return
	}
	s.Matchmaking.MatchWithMobile = d.MatchWithMobile.Value()
	if !d.MatchWithController.Value() && d.u.DeviceGroup().Compare(user.DeviceGroupController()) {
		d.u.Message("form.matchmaking.forced.queueing")
		p.SendForm(NewMatchmaking(d.u))
		return
	}
	s.Matchmaking.MatchWithController = d.MatchWithController.Value()
	if !d.MatchWithKeyboard.Value() && d.u.DeviceGroup().Compare(user.DeviceGroupKeyboardMouse()) {
		d.u.Message("form.matchmaking.forced.queueing")
		p.SendForm(NewMatchmaking(d.u))
		return
	}
	s.Matchmaking.MatchWithKeyboard = d.MatchWithKeyboard.Value()
	s.Matchmaking.PingRange = uint8(d.PingRange.Value())
	d.u.SetSettings(s)
	d.u.Player().SendForm(NewMatchmaking(d.u))
}

// Close ...
func (d matchmaking) Close(form.Submitter) {
	d.u.Player().SendForm(NewSettings(d.u))
}
