package form

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/game"
	"github.com/vasar-network/practice/vasar/user"
	"strings"
)

// competitiveStats is a form that displays the competitive stats of a player.
type competitiveStats struct {
	// id is the xuid or name of the target player.
	id string
}

// NewCompetitiveStats creates a new competitive stats form to send to a player.
func NewCompetitiveStats(id string) form.Form {
	var displayName string
	var stats user.Stats

	if u, ok := user.LookupXUID(id); ok {
		displayName = u.DisplayName()
		stats = u.Stats()
	} else {
		u, _ := data.LoadOfflineUser(id)
		displayName = u.DisplayName()
		stats = u.Stats
	}

	var games []string
	for _, g := range game.Games() {
		games = append(games, text.Colourf("<aqua>%s:</aqua> <white>%v</white>", g.Name(), stats.GameElo[strings.ReplaceAll(strings.ToLower(g.Name()), " ", "_")]))
	}

	return form.NewMenu(competitiveStats{id: id}, displayName+"'s Competitive Stats").WithButtons(
		form.NewButton("View Casual Stats", ""),
	).WithBody(
		text.Colourf("<aqua>Wins:</aqua> <white>%v</white>", stats.RankedWins),
		text.Colourf("\n<aqua>Losses:</aqua> <white>%v</white>\n", stats.RankedLosses),
		text.Colourf("\n<dark-aqua>Elo</dark-aqua>\n"),
		text.Colourf("<aqua>Global:</aqua> <white>%v</white>\n", stats.Elo),
		strings.Join(games, "\n "),
	)
}

// Submit ...
func (c competitiveStats) Submit(s form.Submitter, _ form.Button) {
	p := s.(*player.Player)
	f, err := NewCasualStats(p, c.id)
	if err != nil {
		p.Message(err.Error())
		return
	}
	p.SendForm(f)
}
