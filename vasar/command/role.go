package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/hako/durafmt"
	"github.com/upper/db/v4"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
	"time"
)

// RoleAdd is a command to add a role to a player.
type RoleAdd struct {
	Sub      add
	Targets  []cmd.Target         `cmd:"target"`
	Role     roles                `cmd:"role"`
	Duration cmd.Optional[string] `cmd:"duration"`
}

// RoleRemove is a command to remove a role from a player.
type RoleRemove struct {
	Sub     remove
	Targets []cmd.Target `cmd:"target"`
	Role    roles        `cmd:"role"`
}

// RoleAddOffline is a command to remove a role from an offline user.
type RoleAddOffline struct {
	Sub      add
	Target   string               `cmd:"target"`
	Role     roles                `cmd:"role"`
	Duration cmd.Optional[string] `cmd:"duration"`
}

// RoleRemoveOffline is a command to remove a role from an offline user.
type RoleRemoveOffline struct {
	Sub    remove
	Target string `cmd:"target"`
	Role   roles  `cmd:"role"`
}

// RoleList is a command to list all users with a role.
type RoleList struct {
	Sub  list
	Role roles `cmd:"role"`
}

// Run ...
func (a RoleAdd) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(a.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	p, isPlayer := s.(*player.Player)
	if !isPlayer && len(a.Targets) < 1 {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}

	var t *user.User
	if len(a.Targets) > 0 {
		tP, ok := a.Targets[0].(*player.Player)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		t, ok = user.Lookup(tP)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		if isPlayer && tP != p && t.Roles().Contains(role.Operator{}) {
			o.Error(lang.Translatef(l, "command.role.modify.other"))
			return
		}
	} else {
		var ok bool
		t, ok = user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
	}

	r, _ := role.ByName(string(a.Role))
	if isPlayer {
		u, ok := user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
		if !u.Roles().Contains(role.Operator{}) {
			if u == t {
				o.Error(lang.Translatef(l, "command.role.modify.self"))
				return
			}
			if role.Tier(u.Roles().Highest()) < role.Tier(r) {
				o.Error(lang.Translatef(l, "command.role.higher"))
				return
			}
		}
	}

	duration, hasDuration := a.Duration.Load()
	if t.Roles().Contains(r) {
		e, ok := t.Roles().Expiration(r)
		if !ok {
			o.Error(lang.Translatef(l, "command.role.has", r.Name()))
			return
		}
		if hasDuration {
			duration, err := vails.ParseDuration(duration)
			if err != nil {
				o.Error(lang.Translatef(l, "command.duration.invalid"))
				return
			}
			if e.After(time.Now().Add(duration)) {
				o.Error(lang.Translatef(l, "command.role.has", r.Name()))
				return
			}
		}
		t.Roles().Remove(r)
	}
	t.Roles().Add(r)
	d := "infinity and beyond"
	if hasDuration {
		duration, err := vails.ParseDuration(duration)
		if err != nil {
			o.Error(lang.Translatef(l, "command.duration.invalid"))
			return
		}
		d = durafmt.ParseShort(duration).String()
		t.Roles().Expire(r, time.Now().Add(duration))
	}
	if _, ok := match.Lookup(t.Player()); !ok {
		t.SetNameTagFromRole()
	}
	user.Alert(s, "staff.alert.role.add", r.Name(), t.Player().Name(), d)
	o.Print(lang.Translatef(l, "command.role.add", r.Name(), t.Player().Name(), d))
}

// Run ...
func (d RoleRemove) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(d.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	p, isPlayer := s.(*player.Player)
	if !isPlayer && len(d.Targets) < 1 {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}

	var t *user.User
	if len(d.Targets) > 0 {
		tP, ok := d.Targets[0].(*player.Player)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		t, ok = user.Lookup(tP)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		if isPlayer && tP != p && t.Roles().Contains(role.Operator{}) {
			o.Error(lang.Translatef(l, "command.role.modify.other"))
			return
		}
	} else {
		var ok bool
		t, ok = user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
	}

	r, _ := role.ByName(string(d.Role))
	if isPlayer {
		u, ok := user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
		if !u.Roles().Contains(role.Operator{}) {
			if u == t {
				o.Error(lang.Translatef(l, "command.role.modify.self"))
				return
			}
			if role.Tier(u.Roles().Highest()) < role.Tier(r) {
				o.Error(lang.Translatef(l, "command.role.higher"))
				return
			}
		}
	}

	if !t.Roles().Contains(r) {
		o.Error(lang.Translatef(l, "command.role.missing", r.Name()))
		return
	}
	t.Roles().Remove(r)
	if _, ok := match.Lookup(t.Player()); !ok {
		t.SetNameTagFromRole()
	}

	user.Alert(s, "staff.alert.role.remove", r.Name(), t.Player().Name())
	o.Print(lang.Translatef(l, "command.role.remove", r.Name(), t.Player().Name()))
}

// Run ...
func (a RoleAddOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	t, err := data.LoadOfflineUser(a.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}

	r, _ := role.ByName(string(a.Role))
	if p, ok := s.(*player.Player); ok {
		u, ok := user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
		if role.Tier(u.Roles().Highest()) < role.Tier(r) {
			o.Error(lang.Translatef(l, "command.role.higher"))
			return
		}
		if t.Roles.Contains(role.Operator{}) {
			o.Error(lang.Translatef(l, "command.role.modify.other"))
			return
		}
	}

	if t.Roles.Contains(r) {
		o.Error(lang.Translatef(l, "command.role.has", r.Name()))
		return
	}
	t.Roles.Add(r)
	d := "infinity and beyond"
	duration, hasDuration := a.Duration.Load()
	if hasDuration {
		duration, err := vails.ParseDuration(duration)
		if err != nil {
			o.Error(lang.Translatef(l, "command.duration.invalid"))
			return
		}
		d = durafmt.ParseShort(duration).String()
		t.Roles.Expire(r, time.Now().Add(duration))
	}
	_ = data.SaveOfflineUser(t)

	user.Alert(s, "staff.alert.role.add", r.Name(), t.DisplayName(), d)
	o.Print(lang.Translatef(l, "command.role.add", r.Name(), t.DisplayName(), d))
}

// Run ...
func (d RoleRemoveOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	t, err := data.LoadOfflineUser(d.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}

	r, _ := role.ByName(string(d.Role))
	if p, ok := s.(*player.Player); ok {
		u, ok := user.Lookup(p)
		if !ok {
			// The user somehow left in the middle of this, so just stop in our tracks.
			return
		}
		if role.Tier(u.Roles().Highest()) < role.Tier(r) {
			o.Error(lang.Translatef(l, "command.role.higher"))
			return
		}
		if t.Roles.Contains(role.Operator{}) {
			o.Error(lang.Translatef(l, "command.role.modify.other"))
			return
		}
	}

	if !t.Roles.Contains(r) {
		o.Error(lang.Translatef(l, "command.role.missing", r.Name()))
		return
	}
	t.Roles.Remove(r)
	_ = data.SaveOfflineUser(t)

	user.Alert(s, "staff.alert.role.remove", r.Name(), t.DisplayName())
	o.Print(lang.Translatef(l, "command.role.remove", r.Name(), t.DisplayName()))
}

// Run ...
func (r RoleList) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	users, err := data.SearchOfflineUsers(db.Cond{"roles.name": r.Role})
	if err != nil {
		panic(err)
	}
	if len(users) <= 0 {
		o.Error(lang.Translatef(l, "command.role.list.empty"))
		return
	}

	var usernames []string
	for _, u := range users {
		usernames = append(usernames, u.DisplayName())
	}
	o.Print(lang.Translatef(l, "command.role.list", r.Role, len(users), strings.Join(usernames, ", ")))
}

// Allow ...
func (RoleAdd) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (RoleRemove) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (RoleAddOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (RoleRemoveOffline) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

// Allow ...
func (RoleList) Allow(s cmd.Source) bool {
	return allow(s, true, role.Manager{})
}

type (
	add    string
	remove string
	list   string
	roles  string
)

// SubName ...
func (add) SubName() string {
	return "add"
}

// SubName ...
func (remove) SubName() string {
	return "remove"
}

func (list) SubName() string {
	return "list"
}

// Type ...
func (roles) Type() string {
	return "role"
}

// Options ...
func (roles) Options(s cmd.Source) (roles []string) {
	_, disallow := s.(*player.Player)
	if disallow {
		u, ok := user.Lookup(s.(*player.Player))
		if ok {
			disallow = !u.Roles().Contains(role.Operator{})
		}
	}
	for _, r := range role.All() {
		if _, ok := r.(role.Operator); ok && disallow {
			continue
		}
		roles = append(roles, r.Name())
	}
	return roles
}
