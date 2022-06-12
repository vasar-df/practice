package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	hook "github.com/justtaldevelops/webhook"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/webhook"
	"strings"
)

// Kick is a command that disconnects another player from the server.
type Kick struct {
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (k Kick) Run(s cmd.Source, o *cmd.Output) {
	l, single := locale(s), true
	if len(k.Targets) > 1 {
		if p, ok := s.(*player.Player); ok {
			if u, ok := user.Lookup(p); ok && !u.Roles().Contains(role.Operator{}) {
				o.Error(lang.Translatef(l, "command.targets.exceed"))
				return
			}
		}
		single = false
	}

	var kicked int
	for _, p := range k.Targets {
		if p, ok := p.(*player.Player); ok {
			u, ok := user.Lookup(p)
			if !ok || u.Roles().Contains(role.Operator{}) {
				o.Print(lang.Translatef(l, "command.kick.fail"))
				continue
			}
			p.Disconnect(lang.Translatef(p.Locale(), "command.kick.reason"))
			if single {
				webhook.SendPunishment(s.Name(), p.Name(), "", "Kick")
				o.Print(lang.Translatef(l, "command.kick.success", p.Name()))
				return
			}
			kicked++
		} else if single {
			o.Print(lang.Translatef(l, "command.target.unknown"))
			return
		}
	}
	if !single {
		return
	}
	webhook.Send(webhook.Punishments, hook.Webhook{
		Embeds: []hook.Embed{
			{
				Title: "Kick (Practice)",
				Description: strings.Join([]string{
					fmt.Sprintf("Staff: **`%s`**", s.Name()),
					fmt.Sprintf("Kicked: **`%v players`**", kicked),
				}, "\n"),
				Color: 0xFF0000,
			},
		},
	})
	o.Print(lang.Translatef(l, "command.kick.multiple", kicked))
}

// Allow ...
func (Kick) Allow(s cmd.Source) bool {
	return allow(s, true, role.Trial{})
}
