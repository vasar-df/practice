package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/upper/db/v4"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
)

// AliasOnline is a command used to check the alt accounts of an online player.
type AliasOnline struct {
	Targets []cmd.Target `cmd:"target"`
}

// AliasOffline is a command used to check the alt accounts of an offline player.
type AliasOffline struct {
	Target string `cmd:"target"`
}

// Run ...
func (a AliasOnline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(a.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	target, ok := a.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u, ok := user.Lookup(target)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}

	usersIPs, _ := data.SearchOfflineUsers(db.Cond{"address": u.HashedAddress()})
	ipNames := names(usersIPs, true)

	usersDID, _ := data.SearchOfflineUsers(db.Cond{"did": u.DeviceID()})
	deviceNames := names(usersDID, true)

	usersSSID, _ := data.SearchOfflineUsers(db.Cond{"ssid": u.SelfSignedID()})
	ssidNames := names(usersSSID, true)

	g := text.Colourf("<grey> - </grey>")
	o.Print(lang.Translatef(l, "command.alias.accounts",
		target.Name(), strings.Join(ipNames, g),
		target.Name(), strings.Join(deviceNames, g),
		target.Name(), strings.Join(ssidNames, g)),
	)
}

// Run ...
func (a AliasOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, _ := data.SearchOfflineUsers(db.Cond{"name": strings.ToLower(a.Target)})
	if len(users) == 0 {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	u := users[0]

	usersIPs, _ := data.SearchOfflineUsers(db.Cond{"address": u.Address()})
	ipNames := names(usersIPs, true)

	usersDID, _ := data.SearchOfflineUsers(db.Cond{"did": u.DeviceID()})
	deviceNames := names(usersDID, true)

	usersSSID, _ := data.SearchOfflineUsers(db.Cond{"ssid": u.SelfSignedID()})
	ssidNames := names(usersSSID, true)

	g := text.Colourf("<grey> - </grey>")
	o.Print(lang.Translatef(l, "command.alias.accounts",
		u.DisplayName(), strings.Join(ipNames, g),
		u.DisplayName(), strings.Join(deviceNames, g),
		u.DisplayName(), strings.Join(ssidNames, g)),
	)
}

// Allow ...
func (AliasOnline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}

// Allow ...
func (AliasOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Mod{})
}
