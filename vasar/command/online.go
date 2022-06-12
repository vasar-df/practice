package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
)

// Online is a command that displays the number of players online and their names.
type Online struct{}

// Run ...
func (Online) Run(s cmd.Source, o *cmd.Output) {
	var users []string
	for _, u := range user.All() {
		name := u.Player().Name()
		if name != u.DisplayName() {
			name += fmt.Sprintf("(%s)", u.DisplayName())
		}
		highest := u.Roles().Highest()
		tag := highest.Tag(u.DisplayName())
		if _, ok := highest.(role.Plus); ok {
			tag = strings.ReplaceAll(tag, "§0", u.Settings().Advanced.VasarPlusColour)
		}
		users = append(users, tag)
	}
	o.Printf(lang.Translatef(locale(s), "command.online.users", len(users), strings.Join(users, ", ")))
}

// Allow ...
func (Online) Allow(s cmd.Source) bool {
	return allow(s, true)
}
