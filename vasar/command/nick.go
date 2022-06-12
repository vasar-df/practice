package command

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	hook "github.com/justtaldevelops/webhook"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/webhook"
	"golang.org/x/exp/slices"
	"os"
	"regexp"
	"strings"
)

var (
	// regex is the regex used to make sure the nicknames are valid.
	regex = regexp.MustCompile(`^[a-zA-Z\d ]+$`)
	// forbiddenNames is a list of names that are forbidden.
	forbiddenNames []string
)

// init initializes forbiddenNames.
func init() {
	d, err := os.ReadFile("./assets/forbidden_names.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(d, &forbiddenNames)
	if err != nil {
		panic(err)
	}
}

// Nick is a command used to change your displayed nickname.
type Nick struct {
	Nickname string                     `cmd:"nickname"`
	Targets  cmd.Optional[[]cmd.Target] `cmd:"target"`
}

// Run ...
func (n Nick) Run(s cmd.Source, o *cmd.Output) {
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// User does not exist, so just return.
		return
	}

	l := locale(s)

	if strings.TrimSpace(n.Nickname) == "" {
		o.Error(lang.Translatef(l, "command.nick.invalid"))
		return
	}
	if slices.Contains(forbiddenNames, strings.ToLower(n.Nickname)) {
		o.Error(lang.Translatef(l, "command.nick.forbidden"))
		return
	}
	targets := n.Targets.LoadOr(nil)
	if len(targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}

	t := u
	if len(targets) == 1 {
		if !u.Roles().Contains(role.Manager{}, role.Operator{}) {
			o.Error(lang.Translatef(l, "command.nick.other"))
			return
		}
		p, ok := targets[0].(*player.Player)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		t, ok = user.Lookup(p)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
	}
	if _, ok := match.Lookup(t.Player()); ok {
		o.Error(lang.Translatef(l, "user.feature.disabled"))
		return
	}

	for _, u := range user.All() {
		if strings.EqualFold(u.DisplayName(), n.Nickname) || strings.EqualFold(u.Player().Name(), n.Nickname) {
			o.Error(lang.Translatef(l, "command.nick.used"))
			return
		}
	}

	if !regex.MatchString(n.Nickname) {
		o.Error(lang.Translatef(l, "command.nick.invalid"))
		return
	}
	if len(n.Nickname) < 3 {
		o.Error(lang.Translatef(l, "command.nick.short"))
		return
	}
	if len(n.Nickname) > 13 {
		o.Error(lang.Translatef(l, "command.nick.long"))
		return
	}

	t.SetDisplayName(n.Nickname)
	t.SetNameTagFromRole()
	if t == u {
		user.Alert(s, "staff.alert.name.change", n.Nickname)
		o.Print(lang.Translatef(l, "command.nick.reminder"))
	} else {
		user.Alert(s, "staff.alert.name.change.other", t.Player().Name(), n.Nickname)
		o.Print(lang.Translatef(l, "command.nick.target.nicked", t.Player().Name(), n.Nickname))
	}
	webhook.Send(webhook.Nick, hook.Webhook{
		Embeds: []hook.Embed{
			{
				Title: "Nick (Practice)",
				Description: strings.Join([]string{
					fmt.Sprintf("Requester: **`%s`**", s.Name()),
					fmt.Sprintf("Target: **`%s`**", t.Player().Name()),
					fmt.Sprintf("Nickname: **`%s`**", n.Nickname),
				}, "\n"),
				Color: 0xFFFFFF,
			},
		},
	})
	t.Player().Message(lang.Translatef(t.Player().Locale(), "command.nick.nicked", n.Nickname))
}

// NickReset is a command used to reset the nickname of a user.
type NickReset struct {
	Sub     reset
	Targets cmd.Optional[[]cmd.Target]
}

// Run ...
func (n NickReset) Run(s cmd.Source, o *cmd.Output) {
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// User does not exist, so just return.
		return
	}

	t, l := u, u.Player().Locale()
	targets := n.Targets.LoadOr(nil)
	if len(targets) >= 1 {
		if !u.Roles().Contains(role.Manager{}, role.Operator{}) {
			o.Error(lang.Translatef(l, "command.nick.other"))
			return
		}

		p, ok := targets[0].(*player.Player)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		t, ok = user.Lookup(p)
		if !ok {
			o.Error(lang.Translatef(l, "command.target.unknown"))
			return
		}
		o.Print(lang.Translatef(l, "command.nick.target.reset", t.Player().Name()))
	}
	t.SetDisplayName(t.Player().Name())
	t.SetNameTagFromRole()
	t.Player().Message(lang.Translatef(t.Player().Locale(), "command.nick.reset"))
}

// Allow ...
func (Nick) Allow(s cmd.Source) bool {
	return allow(s, false, role.Plus{}, role.Mod{})
}

// Allow ...
func (NickReset) Allow(s cmd.Source) bool {
	return allow(s, false, role.Plus{}, role.Mod{})
}

// reset ...
type reset string

// SubName ...
func (reset) SubName() string {
	return "reset"
}
