package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/unickorn/strcenter"
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

// BlacklistForm is a command that is used to blacklist a player through a punishment form.
type BlacklistForm struct{}

// BlacklistList is a command that outputs a list of blacklisted players.
type BlacklistList struct {
	Sub blacklistList
}

// BlacklistInfoOffline is a command that displays the blacklist information of an offline player.
type BlacklistInfoOffline struct {
	Sub    blacklistInfo
	Target string `cmd:"target"`
}

// BlacklistLiftOffline is a command that is used to lift the blacklist of an offline player.
type BlacklistLiftOffline struct {
	Sub    blacklistLift
	Target string `cmd:"target"`
}

// Blacklist is a command that is used to blacklist an online player.
type Blacklist struct {
	Targets []cmd.Target              `cmd:"target"`
	Reason  cmd.Optional[cmd.Varargs] `cmd:"reason"`
}

// BlacklistOffline is a command that is used to blacklist an offline player.
type BlacklistOffline struct {
	Target string                    `cmd:"target"`
	Reason cmd.Optional[cmd.Varargs] `cmd:"reason"`
}

// Run ...
func (b BlacklistList) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(db.And(db.Cond{"punishments.ban.permanent": true}))
	if err != nil {
		panic(err)
	}
	if len(users) == 0 {
		o.Error(lang.Translatef(l, "command.blacklist.none"))
		return
	}
	o.Print(lang.Translatef(l, "command.blacklist.list", len(users), strings.Join(names(users, false), ", ")))
}

// Run ...
func (b BlacklistInfoOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, _ := data.LoadOfflineUser(b.Target)
	if u.Ban.Expired() || !u.Ban.Permanent {
		o.Error(lang.Translatef(l, "command.blacklist.not"))
		return
	}
	o.Print(lang.Translatef(l, "punishment.details",
		u.DisplayName(),
		u.Ban.Reason,
		"Permanent",
		u.Ban.Staff,
		u.Ban.Occurrence.Format("01/02/2006"),
	))
}

// Run ...
func (b BlacklistLiftOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(b.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Ban.Expired() {
		o.Error(lang.Translatef(l, "command.blacklist.not"))
		return
	}
	u.Ban = user.Punishment{}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.unblacklist", u.DisplayName())
	webhook.SendPunishment(s.Name(), u.DisplayName(), "", "Unblacklist")
	o.Print(lang.Translatef(l, "command.blacklist.lift", u.DisplayName()))
}

// Run ...
func (BlacklistForm) Run(s cmd.Source, _ *cmd.Output) {
	p := s.(*player.Player)
	p.SendForm(form.NewBlacklist())
}

// Run ...
func (b Blacklist) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(b.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := b.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if t == s {
		o.Error(lang.Translatef(l, "command.blacklist.self"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles().Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.blacklist.operator"))
		return
	}

	reason := strings.TrimSpace(string(b.Reason.LoadOr("")))
	if len(reason) == 0 {
		reason = "None"
	}
	u.SetBan(user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Permanent:  true,
	})
	t.Disconnect(strcenter.CenterLine(strings.Join([]string{
		lang.Translatef(t.Locale(), "user.blacklist.header"),
		lang.Translatef(t.Locale(), "user.blacklist.description", reason),
	}, "\n")))

	user.Alert(s, "staff.alert.blacklist", t.Name())
	user.Broadcast("command.blacklist.broadcast", s.Name(), t.Name())
	webhook.SendPunishment(s.Name(), t.Name(), reason, "Blacklist")
	o.Print(lang.Translatef(l, "command.blacklist.success", t.Name(), reason))
}

// Run ...
func (b BlacklistOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if s.Name() == b.Target {
		o.Error(lang.Translatef(l, "command.blacklist.self"))
		return
	}
	u, err := data.LoadOfflineUser(b.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles.Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.blacklist.operator"))
		return
	}
	if !u.Ban.Expired() && u.Ban.Permanent {
		o.Error(lang.Translatef(l, "command.blacklist.already"))
		return
	}
	reason := strings.TrimSpace(string(b.Reason.LoadOr("")))
	if len(reason) == 0 {
		reason = "None"
	}
	u.Ban = user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Permanent:  true,
	}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.blacklist", u.DisplayName())
	user.Broadcast("command.blacklist.broadcast", s.Name(), u.DisplayName())
	webhook.SendPunishment(s.Name(), u.DisplayName(), reason, "Blacklist")
	o.Print(lang.Translatef(l, "command.blacklist.success", u.DisplayName(), reason))
}

// Allow ...
func (BlacklistList) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (BlacklistInfoOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (BlacklistForm) Allow(s cmd.Source) bool {
	return allow(s, false, role.Manager{})
}

// Allow ...
func (Blacklist) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (BlacklistOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (BlacklistLiftOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

type (
	blacklistList string
	blacklistInfo string
	blacklistLift string
)

// SubName ...
func (blacklistList) SubName() string {
	return "list"
}

// SubName ...
func (blacklistInfo) SubName() string {
	return "info"
}

// SubName ...
func (blacklistLift) SubName() string {
	return "lift"
}
