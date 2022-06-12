package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/vails/lang"
	"strings"
)

// Announce is a command that announces a message to the entire server.
type Announce struct {
	Message cmd.Varargs `cmd:"message"`
}

// Run ...
func (a Announce) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	msg := strings.TrimSpace(string(a.Message))
	if len(msg) == 0 {
		o.Error(lang.Translatef(l, "message.empty"))
		return
	}
	_, _ = chat.Global.WriteString(text.Colourf(strings.ReplaceAll(msg, "\\n", "\n")))
}

// Allow ...
func (Announce) Allow(s cmd.Source) bool {
	return allow(s, true)
}
