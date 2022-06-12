package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/hako/durafmt"
	"github.com/upper/db/v4"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/webhook"
	"strings"
	"time"
)

// MuteForm is a command that is used to mute an online player through a punishment form.
type MuteForm struct{}

// MuteList is a command that outputs a list of muted players.
type MuteList struct {
	Sub muteList
}

// MuteInfo is a command that displays the mute information of an online player.
type MuteInfo struct {
	Sub     muteInfo
	Targets []cmd.Target `cmd:"target"`
}

// MuteInfoOffline is a command that displays the mute information of an offline player.
type MuteInfoOffline struct {
	Sub    muteInfo
	Target string `cmd:"target"`
}

// MuteLift is a command that is used to lift the mute of an online player.
type MuteLift struct {
	Sub     muteLift
	Targets []cmd.Target `cmd:"target"`
}

// MuteLiftOffline is a command that is used to lift the mute of an offline player.
type MuteLiftOffline struct {
	Sub    muteLift
	Target string `cmd:"target"`
}

// Mute is a command that is used to mute an online player.
type Mute struct {
	Targets []cmd.Target `cmd:"target"`
	Reason  muteReason   `cmd:"reason"`
}

// MuteOffline is a command that is used to mute an offline player.
type MuteOffline struct {
	Target string     `cmd:"target"`
	Reason muteReason `cmd:"reason"`
}

// Run ...
func (MuteList) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(
		db.And(
			db.Cond{"punishments.mute.expiration": db.NotEq(time.Time{})},
			db.Cond{"punishments.mute.expiration": db.After(time.Now())},
		),
	)
	if err != nil {
		panic(err)
	}
	if len(users) == 0 {
		o.Error(lang.Translatef(l, "command.mute.none"))
		return
	}
	o.Print(lang.Translatef(l, "command.mute.list", len(users), strings.Join(names(users, true), ", ")))
}

// Run ...
func (m MuteInfo) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	p, ok := m.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(p)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if _, ok := u.Mute(); !ok {
		o.Error(lang.Translatef(l, "command.mute.not"))
		return
	}
	mute, _ := u.Mute()
	o.Print(lang.Translatef(l, "punishment.details",
		p.Name(),
		mute.Reason,
		durafmt.Parse(mute.Remaining()),
		mute.Staff,
		mute.Occurrence.Format("01/02/2006"),
	))
}

// Run ...
func (m MuteInfoOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, _ := data.LoadOfflineUser(m.Target)
	if u.Mute.Expired() {
		o.Error(lang.Translatef(l, "command.mute.not"))
		return
	}
	o.Print(lang.Translatef(l, "punishment.details",
		u.DisplayName(),
		u.Mute.Reason,
		durafmt.Parse(u.Mute.Remaining()),
		u.Mute.Staff,
		u.Mute.Occurrence.Format("01/02/2006"),
	))
}

// Run ...
func (m MuteLift) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	p, ok := m.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(p)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if _, ok := u.Mute(); !ok {
		o.Error(lang.Translatef(l, "command.mute.not"))
		return
	}
	u.SetMute(user.Punishment{})

	user.Alert(s, "staff.alert.unmute", p.Name())
	webhook.SendPunishment(s.Name(), u.DisplayName(), "", "Unmute")
	o.Print(lang.Translatef(l, "command.mute.lift", p.Name()))
}

// Run ...
func (m MuteLiftOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(m.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Mute.Expired() {
		o.Error(lang.Translatef(l, "command.mute.not"))
		return
	}
	u.Mute = user.Punishment{}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.unmute", u.DisplayName())
	webhook.SendPunishment(s.Name(), u.DisplayName(), "", "Unmute")
	o.Print(lang.Translatef(l, "command.mute.lift", u.DisplayName()))
}

// Run ...
func (m MuteForm) Run(s cmd.Source, _ *cmd.Output) {
	p := s.(*player.Player)
	p.SendForm(form.NewMute(p))
}

// Run ...
func (m Mute) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(m.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := m.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if t == s {
		o.Error(lang.Translatef(l, "command.mute.self"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles().Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.mute.operator"))
		return
	}
	if _, ok := u.Mute(); ok {
		o.Error(lang.Translatef(l, "command.mute.already"))
		return
	}

	reason, length := parseMuteReason(m.Reason)
	u.SetMute(user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	})

	user.Alert(s, "staff.alert.mute", t.Name(), reason)
	webhook.SendPunishment(s.Name(), t.Name(), reason, "Mute")
	o.Print(lang.Translatef(l, "command.mute.success", t.Name(), reason))
}

// Run ...
func (m MuteOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if s.Name() == m.Target {
		o.Error(lang.Translatef(l, "command.mute.self"))
		return
	}
	u, err := data.LoadOfflineUser(m.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles.Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.mute.operator"))
		return
	}
	if !u.Mute.Expired() {
		o.Error(lang.Translatef(l, "command.mute.already"))
		return
	}

	reason, length := parseMuteReason(m.Reason)
	u.Mute = user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.mute", u.DisplayName(), reason)
	webhook.SendPunishment(s.Name(), u.DisplayName(), reason, "Mute")
	o.Print(lang.Translatef(l, "command.mute.success", u.DisplayName(), reason))
}

// Allow ...
func (MuteList) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}

// Allow ...
func (MuteInfo) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}

// Allow ...
func (MuteInfoOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}

// Allow ...
func (MuteForm) Allow(s cmd.Source) bool {
	return allow(s, false, role.Trial{})
}

// Allow ...
func (Mute) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}

// Allow ...
func (MuteOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}

// Allow ...
func (MuteLift) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

// Allow ...
func (MuteLiftOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

type (
	muteReason string
	muteList   string
	muteInfo   string
	muteLift   string
)

// Type ...
func (muteReason) Type() string {
	return "muteReason"
}

// Options ...
func (muteReason) Options(cmd.Source) []string {
	return []string{
		"spam",
		"toxic",
		"advertising",
	}
}

// SubName ...
func (muteList) SubName() string {
	return "list"
}

// SubName ...
func (muteInfo) SubName() string {
	return "info"
}

// SubName ...
func (muteLift) SubName() string {
	return "lift"
}

// parseMuteReason returns the formatted muteReason and mute duration.
func parseMuteReason(r muteReason) (string, time.Duration) {
	switch r {
	case "spam":
		return "Spam", time.Hour * 6
	case "toxic":
		return "Toxicity", time.Hour * 9
	case "advertising":
		return "Advertising", time.Hour * 24 * 3
	}
	panic("should never happen")
}
