package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/game/match"
	"golang.org/x/exp/maps"
)

// spectateDuels ...
type spectateDuels struct {
	m map[form.Button]*match.Match
}

// NewSpectate ...
func NewSpectate() form.Menu {
	var matches = make(map[form.Button]*match.Match)
	for _, m := range match.Unranked().RunningMatches() {
		players := m.Players()
		b := form.NewButton(fmt.Sprintf("%v vs %v\n Unranked %v | Spectating: %v", players[0].DisplayName(), players[1].DisplayName(), m.Game().Name(), len(m.Spectators())), "")
		matches[b] = m
	}
	for _, m := range match.Ranked().RunningMatches() {
		players := m.Players()
		b := form.NewButton(fmt.Sprintf("%v vs %v\n Ranked %v | Spectating: %v", players[0].DisplayName(), players[1].DisplayName(), m.Game().Name(), len(m.Spectators())), "")
		matches[b] = m
	}
	return form.NewMenu(spectateDuels{m: matches}, "Spectate").WithButtons(maps.Keys(matches)...)
}

// Submit ...
func (s spectateDuels) Submit(submitter form.Submitter, pressed form.Button) {
	s.m[pressed].AddSpectator(submitter.(*player.Player), false)
}
