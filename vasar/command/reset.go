package command

import (
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"strings"
)

// ResetElo is a command used to reset the elo of a player.
type ResetElo struct {
	Sub     elo
	Targets []cmd.Target   `cmd:"target"`
	Mode    statisticsMode `cmd:"mode"`
}

// Run ...
func (r ResetElo) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(r.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	target, ok := r.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	t, ok := user.Lookup(target)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	stats := t.Stats()
	if r.Mode == "global" {
		stats.Elo = 1000
	} else {
		stats.GameElo[string(r.Mode)] = 1000
	}
	t.SetStats(stats)
	_ = data.SaveUser(t)
	o.Print(lang.Translatef(l, "command.reset.elo", target.Name(), r.Mode))
}

// ResetEloOffline is a command used to reset the elo of an offline player.
type ResetEloOffline struct {
	Sub    elo
	Target string         `cmd:"target"`
	Mode   statisticsMode `cmd:"mode"`
}

// Run ...
func (r ResetEloOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	t, err := data.LoadOfflineUser(r.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	if r.Mode == "global" {
		t.Stats.Elo = 1000
	} else {
		t.Stats.GameElo[string(r.Mode)] = 1000
	}
	_ = data.SaveOfflineUser(t)
	o.Print(lang.Translatef(l, "command.reset.elo", t.DisplayName(), r.Mode))
}

// ResetStats is a command used to reset the stats of a player.
type ResetStats struct {
	Statistic statistics   `cmd:"statistic"`
	Targets   []cmd.Target `cmd:"target"`
}

// Run ...
func (r ResetStats) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	if len(r.Targets) > 1 {
		o.Error(lang.Translatef(l, "command.targets.exceed"))
		return
	}
	target, ok := r.Targets[0].(*player.Player)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	t, ok := user.Lookup(target)
	if !ok {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	stats := t.Stats()
	switch r.Statistic {
	case "kills":
		stats.Kills = 0
	case "deaths":
		stats.Deaths = 0
	case "kill-streak":
		stats.KillStreak = 0
	case "best-kill-streak":
		stats.BestKillStreak = 0
	case "stats":
		stats.Kills = 0
		stats.Deaths = 0
		stats.KillStreak = 0
		stats.BestKillStreak = 0
	case "matches":
		stats.UnrankedWins = 0
		stats.UnrankedLosses = 0
		stats.RankedWins = 0
		stats.RankedLosses = 0
		stats.Elo = 1000
		for _, g := range game.Games() {
			stats.GameElo[g.Name()] = 1000
		}
	case "all":
		stats.Kills = 0
		stats.Deaths = 0
		stats.KillStreak = 0
		stats.BestKillStreak = 0
		stats.UnrankedWins = 0
		stats.UnrankedLosses = 0
		stats.RankedWins = 0
		stats.RankedLosses = 0
		stats.Elo = 1000
		for _, g := range game.Games() {
			stats.GameElo[g.Name()] = 1000
		}
	}
	t.SetStats(stats)
	_ = data.SaveUser(t)
	o.Print(lang.Translatef(l, "command.reset.stats", target.Name(), strings.ToUpper(string(r.Statistic))))
	target.Message(lang.Translatef(l, "command.reset.complete"))
}

// ResetStatsOffline is a command used to reset the stats of an offline player.
type ResetStatsOffline struct {
	Statistic statistics `cmd:"statistic"`
	Target    string     `cmd:"target"`
}

// Run ...
func (r ResetStatsOffline) Run(s cmd.Source, o *cmd.Output) {
	l := locale(s)
	t, err := data.LoadOfflineUser(r.Target)
	if err != nil {
		o.Error(lang.Translatef(l, "command.target.unknown"))
		return
	}
	switch r.Statistic {
	case "kills":
		t.Stats.Kills = 0
	case "deaths":
		t.Stats.Deaths = 0
	case "kill-streak":
		t.Stats.KillStreak = 0
	case "best-kill-streak":
		t.Stats.BestKillStreak = 0
	case "stats":
		t.Stats.Kills = 0
		t.Stats.Deaths = 0
		t.Stats.KillStreak = 0
		t.Stats.BestKillStreak = 0
	case "matches":
		t.Stats.UnrankedWins = 0
		t.Stats.UnrankedLosses = 0
		t.Stats.RankedWins = 0
		t.Stats.RankedLosses = 0
		t.Stats.Elo = 1000
		for _, g := range game.Games() {
			t.Stats.GameElo[g.Name()] = 1000
		}
	case "all":
		t.Stats.Kills = 0
		t.Stats.Deaths = 0
		t.Stats.KillStreak = 0
		t.Stats.BestKillStreak = 0
		t.Stats.UnrankedWins = 0
		t.Stats.UnrankedLosses = 0
		t.Stats.RankedWins = 0
		t.Stats.RankedLosses = 0
		t.Stats.Elo = 1000
		for _, g := range game.Games() {
			t.Stats.GameElo[g.Name()] = 1000
		}
	}
	_ = data.SaveOfflineUser(t)
	o.Print(lang.Translatef(l, "command.reset.stats", t.DisplayName(), strings.ToUpper(string(r.Statistic))))
}

// ResetSeason is a command used to reset match stats of all players.
type ResetSeason struct {
	Sub season
}

// Run ...
func (ResetSeason) Run(cmd.Source, *cmd.Output) {
	players, _ := data.SearchOfflineUsers()
	for _, p := range players {
		p.Stats.RankedWins = 0
		p.Stats.RankedLosses = 0
		p.Stats.Elo = 1000
		for _, g := range game.Games() {
			p.Stats.GameElo[g.Name()] = 1000
		}
		_ = data.SaveOfflineUser(p)
	}
}

// Allow ...
func (ResetElo) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (ResetEloOffline) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (ResetStats) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (ResetStatsOffline) Allow(s cmd.Source) bool {
	return allow(s, true)
}

// Allow ...
func (ResetSeason) Allow(s cmd.Source) bool {
	return allow(s, true)
}

type (
	elo            string
	season         string
	statistics     string
	statisticsMode string
)

func (elo) SubName() string {
	return "elo"
}
func (season) SubName() string {
	return "season"
}

func (statistics) Type() string {
	return "statistics"
}
func (statistics) Options(cmd.Source) []string {
	return []string{"kills", "deaths", "kill-streak", "best-kill-streak", "stats", "matches"}
}

// Type ...
func (statisticsMode) Type() string {
	return "statisticMode"
}

// Options ...
func (statisticsMode) Options(cmd.Source) []string {
	modes := []string{"global"}
	for _, m := range game.Games() {
		modes = append(modes, strings.ReplaceAll(strings.ToLower(m.Name()), " ", "_"))
	}
	return modes
}
