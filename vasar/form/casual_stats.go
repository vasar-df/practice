package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"time"
)

// casualStats is a form that displays the casual stats of a player.
type casualStats struct {
	// id is the xuid or name of the target player.
	id string
}

// NewCasualStats creates a new casual stats form to send to a player. An error will be returned if the offline user
// or if the offline user has their stats hidden.
func NewCasualStats(p *player.Player, id string) (form.Form, error) {
	var (
		displayName     string
		playtimeSession time.Duration
		playtimeTotal   time.Duration
		stats           user.Stats
	)

	if u, ok := user.LookupXUID(id); ok {
		displayName = u.DisplayName()
		playtimeSession = time.Since(u.JoinTime()).Round(time.Second)
		playtimeTotal = u.PlayTime().Round(time.Second)
		stats = u.Stats()
	} else {
		u, err := data.LoadOfflineUser(id)
		if err != nil {
			return nil, err
		}
		if !u.Settings.Privacy.PublicStatistics {
			return nil, fmt.Errorf(lang.Translatef(p.Locale(), "command.stats.private"))
		}
		displayName = u.DisplayName()
		playtimeTotal = u.PlayTime().Round(time.Second)
		stats = u.Stats
	}

	return form.NewMenu(casualStats{id: id}, fmt.Sprintf("%v's Casual Stats", displayName)).WithButtons(
		form.NewButton("View Competitive Stats", ""),
	).WithBody(
		text.Colourf(" <aqua>Playtime (Session):</aqua> <white>%s</white>\n", playtimeSession),
		text.Colourf("<aqua>Playtime (All Time):</aqua> <white>%s</white>\n", playtimeTotal),
		text.Colourf("<aqua>Kills:</aqua> <white>%v</white>\n", stats.Kills),
		text.Colourf("<aqua>Killstreak:</aqua> <white>%v</white>\n", stats.KillStreak),
		text.Colourf("<aqua>Best Killstreak:</aqua> <white>%v</white>\n", stats.BestKillStreak),
		text.Colourf("<aqua>Deaths:</aqua> <white>%v</white>\n", stats.Deaths),
	), nil
}

// Submit ...
func (c casualStats) Submit(s form.Submitter, _ form.Button) {
	s.(*player.Player).SendForm(NewCompetitiveStats(c.id))
}
