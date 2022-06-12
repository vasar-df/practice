package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/unickorn/strcenter"
	"github.com/upper/db/v4"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
)

// WhitelistAdd is a command that adds an online user to the whitelist.
type WhitelistAdd struct {
	Sub     add
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (a WhitelistAdd) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(a.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := a.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Whitelisted() {
		o.Error(lang.Translatef(l, "command.whitelist.already.added", t.Name()))
		return
	}
	u.Whitelist()
	_ = data.SaveUser(u)

	o.Print(lang.Translatef(l, "command.whitelist.added", t.Name()))
}

// Allow ...
func (WhitelistAdd) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// WhitelistAddOffline is a command that adds a user to the whitelist.
type WhitelistAddOffline struct {
	Sub    add
	Target string `cmd:"target"`
}

// Run ...
func (a WhitelistAddOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(a.Target)
	if err != nil {
		u = data.NewOfflineUser(a.Target)
	}
	if u.Whitelisted {
		o.Error(lang.Translatef(l, "command.whitelist.already.added", a.Target))
		return
	}
	u.Whitelisted = true
	_ = data.SaveOfflineUser(u)

	o.Print(lang.Translatef(l, "command.whitelist.added", a.Target))
}

// Allow ...
func (WhitelistAddOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// WhitelistRemove is a command that removes an online user from the whitelist.
type WhitelistRemove struct {
	Sub     remove
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (w WhitelistRemove) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(w.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := w.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if !u.Whitelisted() {
		o.Error(lang.Translatef(l, "command.whitelist.already.removed", t.Name()))
		return
	}
	u.Unwhitelist()
	if !u.Roles().Contains(role.Trial{}, role.Operator{}) {
		t.Disconnect(strcenter.CenterLine(lang.Translatef(l, "user.server.whitelist")))
	} else {
		_ = data.SaveUser(u)
	}

	o.Print(lang.Translatef(l, "command.whitelist.removed", t.Name()))
}

// Allow ...
func (WhitelistRemove) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// WhitelistRemoveOffline is a command that removes a user from the whitelist.
type WhitelistRemoveOffline struct {
	Sub    remove
	Target string `cmd:"target"`
}

// Run ...
func (r WhitelistRemoveOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, err := data.LoadOfflineUser(r.Target)
	if err != nil {
		u = data.NewOfflineUser(r.Target)
	}
	if !u.Whitelisted {
		o.Error(lang.Translatef(l, "command.whitelist.already.removed", r.Target))
		return
	}
	u.Whitelisted = false
	_ = data.SaveOfflineUser(u)

	o.Print(lang.Translatef(l, "command.whitelist.removed", r.Target))
}

// Allow ...
func (WhitelistRemoveOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// WhitelistClear is a command that clears the whitelist.
type WhitelistClear struct {
	Sub clear
}

// Run ...
func (w WhitelistClear) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(db.Cond{"whitelisted": true})
	if err != nil {
		panic(err)
	}
	for _, d := range users {
		if u, ok := user.LookupXUID(d.XUID()); ok {
			u.Unwhitelist()
			if !u.Roles().Contains(role.Trial{}, role.Operator{}) {
				u.Player().Disconnect(strcenter.CenterLine(lang.Translatef(l, "user.server.whitelist")))
				continue
			}
			_ = data.SaveUser(u)
			continue
		}
		d.Whitelisted = false
		_ = data.SaveOfflineUser(d)
	}
	o.Print(lang.Translatef(l, "command.whitelist.cleared"))
}

// Allow ...
func (WhitelistClear) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Whitelist is a command that lists all whitelisted users.
type Whitelist struct {
	Sub list
}

// Run ...
func (w Whitelist) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(db.Cond{"whitelisted": true})
	if err != nil {
		panic(err)
	}

	whitelisted := names(users, true)
	o.Print(lang.Translatef(l, "command.whitelist", len(whitelisted), strings.Join(whitelisted, ", ")))
}

// Allow ...
func (Whitelist) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

type (
	clear string
)

// SubName ...
func (clear) SubName() string {
	return "clear"
}
