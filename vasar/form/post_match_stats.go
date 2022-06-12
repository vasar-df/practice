package form

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"github.com/vasar-network/practice/vasar/game/match"
	it "github.com/vasar-network/practice/vasar/item"
	"github.com/vasar-network/practice/vasar/user"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
)

// postMatchStats is a form that displays the post-match stats of a match
type postMatchStats struct {
	// stats maps between a player and their stats.
	stats map[string]any
	// index maps between a form.Button and a player.
	index map[form.Button]string
	// u is the user that is using the form.
	u *user.User
}

// specificPostMatchStats is a form that displays the stats of a specific player.
type specificPostMatchStats struct {
	// name is the name of the player.
	name string
	// stats is the stats to be displayed.
	stats match.Stats
	// u is the user that is using the form.
	u *user.User
}

// NewPostMatchStats creates a new form for the player to view the post-match stats of a match.
func NewPostMatchStats(u *user.User) form.Form {
	p := u.Player()
	stats, _ := u.PostMatchStats()

	selections := maps.Keys(stats)

	ind := slices.Index(selections, p.Name())
	selections = slices.Delete(selections, ind, ind+1)

	sort.Strings(selections)
	selections = append([]string{p.Name()}, selections...)

	s := postMatchStats{
		index: make(map[form.Button]string),
		stats: stats,
		u:     u,
	}
	m := form.NewMenu(s, "Post Match Details")
	for _, v := range selections {
		button := form.NewButton(fmt.Sprintf("%v's Statistics", v), "")
		if v == p.Name() {
			button = form.NewButton("My Statistics", "")
		}

		s.index[button] = v
		m = m.WithButtons(button)
	}
	return m
}

// NewSpecificPostMatchStats creates a new form for the player to view the stats of a specific player.
func NewSpecificPostMatchStats(u *user.User, name string, stats match.Stats) form.Form {
	s := specificPostMatchStats{
		name:  name,
		stats: stats,
		u:     u,
	}

	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("Total Hits: %d\n", stats.Hits))
	b.WriteString(fmt.Sprintf("Damage Dealt: %.2f\n", stats.Damage))

	var potions, apples int
	for _, i := range stats.Items {
		switch i.Item().(type) {
		case it.VasarPotion:
			potions += i.Count()
		case item.GoldenApple:
			apples += i.Count()
		}
	}
	if potions > 0 {
		b.WriteString(fmt.Sprintf("Splash Potions: %d\n", potions))
	}
	if apples > 0 {
		b.WriteString(fmt.Sprintf("Golden Apples: %d\n", apples))
	}

	m := form.NewMenu(s, fmt.Sprintf("%v's Statistics", name)).WithBody(text.Colourf(b.String()))
	return m
}

// Submit ...
func (p postMatchStats) Submit(_ form.Submitter, pressed form.Button) {
	name := p.index[pressed]
	stats := p.stats[name].(match.Stats)
	p.u.Player().SendForm(NewSpecificPostMatchStats(p.u, name, stats))
}

// Submit ...
func (s specificPostMatchStats) Submit(form.Submitter, form.Button) {
	s.u.Player().SendForm(NewSpecificPostMatchStats(s.u, s.name, s.stats))
}
