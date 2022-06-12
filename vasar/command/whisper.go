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

// Whisper is a command that allows a player to send a private message to another player.
type Whisper struct {
	Target  []cmd.Target `cmd:"target"`
	Message cmd.Varargs  `cmd:"message"`
}

// Run ...
func (w Whisper) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// The user somehow left in the middle of this, so just stop in our tracks.
		return
	}
	if !u.Settings().Privacy.PrivateMessages {
		o.Error(lang.Translatef(l, "user.whisper.disabled"))
		return
	}
	msg := strings.TrimSpace(string(w.Message))
	if len(msg) <= 0 {
		o.Error(lang.Translatef(l, "message.empty"))
		return
	}
	if len(w.Target) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}

	tP, ok := w.Target[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	t, ok := user.Lookup(tP)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
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
func (Whisper) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}
