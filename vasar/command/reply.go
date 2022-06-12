package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
)

// Reply is a command that allows a player to reply to their most recent private message.
type Reply struct {
	Message cmd.Varargs `cmd:"message"`
}

// Run ...
func (r Reply) Run(s cmd.Source, o *cmd.Output) {
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// The user somehow left in the middle of this, so just stop in our tracks.
		return
	}
	l := u.Player().Locale()
	if !u.Settings().Privacy.PrivateMessages {
		o.Error(lang.Translatef(l, "user.whisper.disabled"))
		return
	}
	msg := strings.TrimSpace(string(r.Message))
	if len(msg) <= 0 {
		o.Error(lang.Translatef(l, "message.empty"))
		return
	}

	t, ok := u.LastMessageFrom()
	if !ok {
		o.Error(lang.Translatef(l, "command.reply.none"))
		return
	}
	if !t.Settings().Privacy.PrivateMessages {
		o.Error(lang.Translatef(l, "target.whisper.disabled"))
		return
	}

	uTag, uMsg := text.Colourf("<white>%s</white>", u.DisplayName()), text.Colourf("<white>%s</white>", msg)
	tTag, tMsg := text.Colourf("<white>%s</white>", t.DisplayName()), text.Colourf("<white>%s</white>", msg)
	if _, ok := u.Roles().Highest().(role.Default); !ok {
		uMsg = t.Roles().Highest().Tag(msg)
		uTag = u.Roles().Highest().Tag(u.DisplayName())
	}
	if _, ok := t.Roles().Highest().(role.Default); !ok {
		tMsg = u.Roles().Highest().Tag(msg)
		tTag = t.Roles().Highest().Tag(t.DisplayName())
	}

	t.SetLastMessageFrom(u.Player())
	t.SendCustomSound("random.orb", 1, 1, false)
	u.Message("command.whisper.to", tTag, tMsg)
	t.Message("command.whisper.from", uTag, uMsg)
}

// Allow ...
func (Reply) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
