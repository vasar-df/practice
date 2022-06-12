package command

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	hook "github.com/justtaldevelops/webhook"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/webhook"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"time"
)

// Report is a command used to report other players.
type Report struct {
	Targets []cmd.Target `cmd:"target"`
	Reason  reason       `cmd:"reason"`
}

// Run ...
func (r Report) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	u, ok := user.Lookup(s.(*player.Player))
	if !ok {
		// User somehow left midway through, just stop in our tracks.
		return
	}
	if len(r.Targets) < 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := r.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if s == t {
		o.Error(lang.Translatef(l, "command.report.self"))
		return
	}
	if exp := u.ReportSince().Add(time.Minute); exp.After(time.Now()) {
		o.Error(lang.Translatef(l, "command.report.cooldown", time.Until(exp).Round(time.Millisecond*10)))
		return
	}
	u.RenewReportSince()
	o.Print(lang.Translatef(l, "command.report.success"))
	user.Alert(s, "staff.alert.report", t.Name(), r.Reason)
	webhook.Send(webhook.Report, hook.Webhook{
		Embeds: []hook.Embed{{
			Title: "Report (Practice)",
			Color: 0xFFFFFF,
			Description: strings.Join([]string{
				fmt.Sprintf("**Player:** %v", t.Name()),
				fmt.Sprintf("**Reporter:** %v", s.Name()),
				fmt.Sprintf("**Reason:** %v", cases.Title(language.English).String(string(r.Reason))),
			}, "\n"),
		}},
	})
}

// Allow ...
func (Report) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

type reason string

// Type ...
func (reason) Type() string {
	return "reason"
}

// Options ...
func (reason) Options(cmd.Source) []string {
	return []string{
		"cheating",
		"teaming",
		"interfering",
		"spam",
		"threats",
		"glitching",
		"exploiting",
		"toxic",
	}
}
