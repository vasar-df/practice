package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"strings"
)

// Chat is a command used to check the available chats a player can switch to.
type Chat struct{}

// ChatSwitch is a command used to switch to available chats.
type ChatSwitch struct {
	ChatType chatType `cmd:"chat"`
}

// Run ...
func (c Chat) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	t := []string{"Global"}
	if u, ok := user.Lookup(p); ok && u.Roles().Staff() {
		t = append(t, "Staff")
	}
	// TODO: Append a party option if the user is in a party, when parties are implemented.
	o.Print(lang.Translatef(p.Locale(), "command.chat.available", strings.Join(t, ", ")))
}

// Run ...
func (c ChatSwitch) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	u, ok := user.Lookup(p)
	if !ok {
		return
	}

	var t user.ChatType
	switch c.ChatType {
	case "global":
		t = user.ChatTypeGlobal()
	case "staff":
		t = user.ChatTypeStaff()
	case "party":
		t = user.ChatTypeParty()
	}
	if u.ChatType() == t {
		o.Errorf(lang.Translatef(p.Locale(), "command.chat.already", t.String()))
		return
	}
	o.Print(lang.Translatef(p.Locale(), "command.chat.switch", t.String()))
	u.UpdateChatType(t)
}

// Allow ...
func (c Chat) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

// Allow ...
func (c ChatSwitch) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

type chatType string

// Type ...
func (c chatType) Type() string {
	return "chatType"
}

// Options ...
func (c chatType) Options(s cmd.Source) []string {
	o := []string{"global"}
	if u, ok := user.Lookup(s.(*player.Player)); ok && u.Roles().Staff() {
		o = append(o, "staff")
	}
	// TODO: Append a party option if the user is in a party, when parties are implemented.
	return o
}
