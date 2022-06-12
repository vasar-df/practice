package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/form"
	"github.com/vasar-network/practice/vasar/game/lobby"
	"github.com/vasar-network/practice/vasar/game/match"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
)

// DuelRequest is a command used to request a duel to a specific player.
type DuelRequest struct {
	Targets []cmd.Target `cmd:"target"`
}

// Run ...
func (d DuelRequest) Run(s cmd.Source, o *cmd.Output) {
	p := s.(*player.Player)
	l := p.Locale()
	if _, ok := lobby.LookupProvider(p); !ok {
		o.Error(lang.Translatef(l, "user.feature.disabled"))
		return
	}
	if len(d.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	t, ok := d.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if p == t {
		o.Error(lang.Translatef(l, "duel.self"))
		return
	}
	if _, ok := lobby.LookupProvider(t); !ok {
		o.Error(lang.Translatef(l, "duel.unavailable"))
		return
	}
	u, ok := user.Lookup(t)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if u.Settings().Privacy.DuelRequests {
		p.SendForm(form.NewDuelRequest(u))
		return
	}
	o.Error(lang.Translatef(l, "duel.requests.disabled"))
}

// Allow ...
func (DuelRequest) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

// DuelRespond is a command used to accept or deny a duel request last sent to the player.
type DuelRespond struct {
	Response response `cmd:"response"`
}

// Run ...
func (d DuelRespond) Run(s cmd.Source, o *cmd.Output) {
	if p := s.(*player.Player); d.Response == "accept" {
		if _, ok := lobby.LookupProvider(p); !ok {
			o.Error(lang.Translatef(p.Locale(), "duel.match"))
			return
		}
		if !match.Unranked().AcceptDuel(p) {
			o.Print(lang.Translatef(p.Locale(), "duel.none"))
		}
	} else if !match.Unranked().DeclineDuel(p) {
		o.Print(lang.Translatef(p.Locale(), "duel.none"))
	}
}

// Allow ...
func (DuelRespond) Allow(s cmd.Source) bool {
	_, ok := s.(*player.Player)
	return ok
}

// response ...
type response string

// Type ...
func (r response) Type() string {
	return "duelResponse"
}

// Options ...
func (r response) Options(cmd.Source) []string {
	return []string{"accept", "decline"}
}
