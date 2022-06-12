package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails"
	"github.com/vasar-network/vails/role"
	"golang.org/x/text/language"
)

// locale returns the locale of a cmd.Source.
func locale(s cmd.Source) language.Tag {
	if p, ok := s.(*player.Player); ok {
		return p.Locale()
	}
	return language.English
}

// allow is a helper function for command allowers. It allows users to easily check for the specified roles.
func allow(src cmd.Source, console bool, roles ...vails.Role) bool {
	p, ok := src.(*player.Player)
	if !ok {
		return console
	}
	u, ok := user.Lookup(p)
	return ok && u.Roles().Contains(append(roles, role.Operator{})...)
}

// names returns a list of formatted names from a list of users. If tag is true a tag will be shown next to the name
// indicating if the user is banned or blacklisted.
func names(users []data.User, tag bool) (names []string) {
	for _, u := range users {
		if u.Ban.Permanent {
			if tag {
				names = append(names, text.Colourf("<dark-red>%s [BLACKLISTED]</dark-red>", u.DisplayName()))
			} else {
				names = append(names, text.Colourf("<dark-red>%s</dark-red>", u.DisplayName()))
			}
		} else if !u.Ban.Expired() {
			if tag {
				names = append(names, text.Colourf("<red>%s [BANNED]</red>", u.DisplayName()))
			} else {
				names = append(names, text.Colourf("<red>%s</red>", u.DisplayName()))
			}
		} else {
			names = append(names, text.Colourf("<green>%s</green>", u.DisplayName()))
		}
	}
	return
}
