package form

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/webhook"
	"golang.org/x/exp/maps"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// mute is a form that allows a user to issue a mute.
type mute struct {
	// Reason is a dropdown that allows a user to select a mute reason.
	Reason form.Dropdown
	// OnlinePlayer is a dropdown that allows a user to select an online player.
	OnlinePlayer form.Dropdown
	// OfflinePlayer is an input field that allows a user to enter an offline player.
	OfflinePlayer form.Input
	// online is a list of online players' XUIDs indexed by their names.
	online map[string]string
	// p is the player that is using the form.
	p *player.Player
}

// NewMute creates a new form to issue a mute.
func NewMute(p *player.Player) form.Form {
	online := make(map[string]string)
	for _, u := range user.All() {
		online[u.Player().Name()] = u.Player().XUID()
	}
	names := [...]string{"Steve Harvey", "Elon Musk", "Bill Gates", "Mark Zuckerberg", "Jeff Bezos", "Warren Buffet", "Larry Page", "Sergey Brin", "Larry Ellison", "Tim Cook", "Steve Ballmer", "Daniel Larson", "Steve"}
	list := maps.Keys(online)
	sort.Strings(list)
	return form.New(mute{
		Reason:        form.NewDropdown("Reason", []string{"Spam", "Toxicity", "Advertisement"}, 0),
		OnlinePlayer:  form.NewDropdown("Online Player", list, 0),
		OfflinePlayer: form.NewInput("Offline Player", "", names[rand.Intn(len(names)-1)]),
		online:        online,
		p:             p,
	}, "Mute")
}

// Submit ...
func (m mute) Submit(form.Submitter) {
	u, ok := user.Lookup(m.p)
	if !ok {
		// User somehow left midway through the form.
		return
	}
	if !u.Roles().Contains(role.Trial{}, role.Operator{}) {
		// In case the user's role was removed while the form was open.
		return
	}
	var length time.Duration
	reason := m.Reason.Options[m.Reason.Value()]
	switch reason {
	case "Spam":
		length = time.Hour * 6
	case "Toxicity":
		length = time.Hour * 9
	case "Advertising":
		length = time.Hour * 24 * 3
	default:
		panic("should never happen")
	}

	mu := user.Punishment{
		Staff:      m.p.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	}
	if offlineName := strings.TrimSpace(m.OfflinePlayer.Value()); offlineName != "" {
		if strings.EqualFold(offlineName, m.p.Name()) {
			u.Message("command.mute.self")
			return
		}
		t, err := data.LoadOfflineUser(offlineName)
		if err != nil {
			u.Message("command.target.unknown")
			return
		}
		if t.Roles.Contains(role.Operator{}) {
			u.Message("command.mute.operator")
			return
		}
		if !t.Mute.Expired() {
			u.Message("command.mute.already")
			return
		}
		t.Mute = mu
		_ = data.SaveOfflineUser(t)

		user.Alert(m.p, "staff.alert.mute", t.DisplayName(), reason)
		webhook.SendPunishment(m.p.Name(), t.DisplayName(), reason, "Mute")
		u.Message("command.mute.success", t.DisplayName(), reason)
		return
	}
	t, ok := user.LookupXUID(m.online[m.OnlinePlayer.Options[m.OnlinePlayer.Value()]])
	if !ok {
		u.Message("command.target.unknown")
		return
	}
	if t.Roles().Contains(role.Operator{}) {
		u.Message("command.mute.operator")
		return
	}
	if _, ok := t.Mute(); ok {
		u.Message("command.mute.already")
		return
	}
	t.SetMute(mu)
	_ = data.SaveUser(t) // Save in case of a server crash or anything that may cause the data to not get saved.

	user.Alert(m.p, "staff.alert.mute", t.Player().Name(), reason)
	webhook.SendPunishment(m.p.Name(), t.Player().Name(), reason, "Mute")
	u.Message("command.mute.success", t.Player().Name(), reason)
}
