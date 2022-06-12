package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/hako/durafmt"
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

// BanForm is a command that is used to ban a player through a punishment form.
type BanForm struct{}

// BanList is a command that outputs a list of banned players.
type BanList struct {
	Sub banList
}

// BanInfoOffline is a command that displays the ban information of an offline player.
type BanInfoOffline struct {
	Sub    banInfo
	Target string `cmd:"target"`
}

// BanLiftOffline is a command that is used to lift the ban of an offline player.
type BanLiftOffline struct {
	Sub    banLift
	Target string `cmd:"target"`
}

// Ban is a command that is used to ban an online player.
type Ban struct {
	Targets []cmd.Target `cmd:"target"`
	Reason  banReason    `cmd:"reason"`
}

// BanOffline is a command that is used to ban an offline player.
type BanOffline struct {
	Target string    `cmd:"target"`
	Reason banReason `cmd:"reason"`
}

// Run ...
func (BanList) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(
		db.And(
			db.Cond{"punishments.ban.expiration": db.NotEq(time.Time{})},
			db.Cond{"punishments.ban.expiration": db.After(time.Now())},
		),
	)
	if err != nil {
		panic(err)
	}
	if len(users) == 0 {
		o.Error(lang.Translatef(l, "command.ban.none"))
		return
	}
	o.Print(lang.Translatef(l, "command.ban.list", len(users), strings.Join(names(users, false), ", ")))
}

// Run ...
func (b BanInfoOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, _ := data.LoadOfflineUser(b.Target)
	if u.Ban.Expired() || u.Ban.Permanent {
		o.Error(lang.Translatef(l, "command.ban.not"))
		return
	}
	o.Print(lang.Translatef(l, "punishment.details",
		u.DisplayName(),
		u.Ban.Reason,
		durafmt.ParseShort(u.Ban.Remaining()),
		u.Ban.Staff,
		u.Ban.Occurrence.Format("01/02/2006"),
	))
}

// Run ...
func (b BanLiftOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(b.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Ban.Expired() || u.Ban.Permanent {
		o.Error(lang.Translatef(l, "command.ban.not"))
		return
	}
	u.Ban = user.Punishment{}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.unban", u.DisplayName())
	webhook.SendPunishment(s.Name(), u.DisplayName(), "", "Unban")
	o.Print(lang.Translatef(l, "command.ban.lift", u.DisplayName()))
}

// Run ...
func (BanForm) Run(s cmd.Source, _ *cmd.Output) {
	p := s.(*player.Player)
	p.SendForm(form.NewBan())
}

// Run ...
func (b Ban) Run(s cmd.Source, o *cmd.Output) {
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
		o.Error(lang.Translatef(l, "command.ban.self"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles().Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.ban.operator"))
		return
	}
	reason, length := parseBanReason(b.Reason)
	u.SetBan(user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	})
	t.Disconnect(strcenter.CenterLine(strings.Join([]string{
		lang.Translatef(t.Locale(), "user.ban.header"),
		lang.Translatef(t.Locale(), "user.ban.description", reason, durafmt.ParseShort(length)),
	}, "\n")))

	user.Alert(s, "staff.alert.ban", t.Name(), reason)
	user.Broadcast("command.ban.broadcast", s.Name(), t.Name(), reason)
	webhook.SendPunishment(s.Name(), t.Name(), reason, "Ban")
	o.Print(lang.Translatef(l, "command.ban.success", t.Name(), reason))
}

// Run ...
func (b BanOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if strings.EqualFold(s.Name(), b.Target) {
		o.Error(lang.Translatef(l, "command.ban.self"))
		return
	}
	u, err := data.LoadOfflineUser(b.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Roles.Contains(role.Operator{}) {
		o.Error(lang.Translatef(l, "command.ban.operator"))
		return
	}
	if !u.Ban.Expired() {
		o.Error(lang.Translatef(l, "command.ban.already"))
		return
	}

	reason, length := parseBanReason(b.Reason)
	u.Ban = user.Punishment{
		Staff:      s.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	}
	_ = data.SaveOfflineUser(u)

	user.Alert(s, "staff.alert.ban", u.DisplayName(), reason)
	user.Broadcast("command.ban.broadcast", s.Name(), u.DisplayName(), reason)
	webhook.SendPunishment(s.Name(), u.DisplayName(), reason, "Ban")
	o.Print(lang.Translatef(l, "command.ban.success", u.DisplayName(), reason))
}

// Allow ...
func (BanList) Allow(s cmd.Source) bool {
	return allow(s, true, role.Admin{})
}

// Allow ...
func (BanInfoOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

// Allow ...
func (BanForm) Allow(s cmd.Source) bool {
	return allow(s, false, role.Mod{})
}

// Allow ...
func (Ban) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

// Allow ...
func (BanOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

// Allow ...
func (BanLiftOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

type (
	banReason string
	banList   string
	banInfo   string
	banLift   string
)

// SubName ...
func (banList) SubName() string {
	return "list"
}

// SubName ...
func (banInfo) SubName() string {
	return "info"
}

// SubName ...
func (banLift) SubName() string {
	return "lift"
}

// Type ...
func (banReason) Type() string {
	return "banReason"
}

// Options ...
func (banReason) Options(cmd.Source) []string {
	return []string{
		"advantage",
		"ranked_advantage",
		"interference",
		"exploitation",
		"abuse",
		"skin",
		"advertisement",
		"evasion",
	}
}

// parseBanReason returns the formatted BanReason and ban duration.
func parseBanReason(r banReason) (string, time.Duration) {
	switch r {
	case "advantage":
		return "Unfair Advantage", time.Hour * 24 * 30
	case "ranked_advantage":
		return "Unfair Advantage in Ranked", time.Hour * 24 * 90
	case "interference":
		return "Interference", time.Hour * 12
	case "exploitation":
		return "Exploitation", time.Hour * 24 * 9
	case "abuse":
		return "Permission Abuse", time.Hour * 24 * 30
	case "skin":
		return "Invalid Skin", time.Hour * 24 * 3
	case "evasion":
		return "Evasion", time.Hour * 24 * 120
	case "advertisement":
		return "Advertisement", time.Hour * 24 * 6
	}
	panic("should never happen")
}
