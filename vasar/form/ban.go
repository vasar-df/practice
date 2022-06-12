package form

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/hako/durafmt"
	"github.com/unickorn/strcenter"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"github.com/vasar-network/vails/webhook"
	"golang.org/x/exp/maps"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// ban is a form that allows a user to issue a ban.
type ban struct {
	// Reason is a dropdown that allows a user to select a ban reason.
	Reason form.Dropdown
	// OnlinePlayer is a dropdown that allows a user to select an online player.
	OnlinePlayer form.Dropdown
	// OfflinePlayer is an input field that allows a user to enter an offline player.
	OfflinePlayer form.Input
	// online is a list of online players' XUIDs indexed by their names.
	online map[string]string
}

// NewBan creates a new form to issue a ban.
func NewBan() form.Form {
	online := make(map[string]string)
	for _, u := range user.All() {
		online[u.Player().Name()] = u.Player().XUID()
	}
	names := [...]string{"Steve Harvey", "Elon Musk", "Bill Gates", "Mark Zuckerberg", "Jeff Bezos", "Warren Buffet", "Larry Page", "Sergey Brin", "Larry Ellison", "Tim Cook", "Steve Ballmer", "Daniel Larson", "Steve"}
	list := maps.Keys(online)
	sort.Strings(list)
	return form.New(ban{
		Reason:        form.NewDropdown("Reason", []string{"Unfair Advantage", "Unfair Advantage in Ranked", "Interference", "Exploitation", "Permission Abuse", "Invalid Skin", "Evasion", "Advertising"}, 0),
		OnlinePlayer:  form.NewDropdown("Online Player", list, 0),
		OfflinePlayer: form.NewInput("Offline Player", "", names[rand.Intn(len(names)-1)]),
		online:        online,
	}, "Ban")
}

// Submit ...
func (b ban) Submit(s form.Submitter) {
	p := s.(*player.Player)
	u, ok := user.Lookup(p)
	if !ok {
		// User somehow left midway through the form.
		return
	}
	if !u.Roles().Contains(role.Trial{}, role.Operator{}) {
		// In case the user's role was removed while the form was open.
		return
	}
	var length time.Duration
	reason := b.Reason.Options[b.Reason.Value()]
	switch reason {
	case "Unfair Advantage":
		length = time.Hour * 24 * 30
	case "Unfair Advantage in Ranked":
		length = time.Hour * 24 * 90
	case "Interference":
		length = time.Hour * 12
	case "Exploitation":
		length = time.Hour * 24 * 9
	case "Permission Abuse":
		length = time.Hour * 24 * 30
	case "Invalid Skin":
		length = time.Hour * 24 * 3
	case "Evasion":
		length = time.Hour * 24 * 120
	case "Advertising":
		length = time.Hour * 24 * 6
	}

	punishment := user.Punishment{
		Staff:      p.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Expiration: time.Now().Add(length),
	}
	var name string
	if offlineName := strings.TrimSpace(b.OfflinePlayer.Value()); offlineName != "" {
		if strings.EqualFold(offlineName, p.Name()) {
			u.Message("command.ban.self")
			return
		}
		t, err := data.LoadOfflineUser(offlineName)
		if err != nil {
			u.Message("command.target.unknown")
			return
		}
		if t.Roles.Contains(role.Operator{}) {
			u.Message("command.ban.operator")
			return
		}
		if !t.Ban.Expired() {
			u.Message("command.ban.already")
			return
		}
		t.Ban = punishment
		_ = data.SaveOfflineUser(t)
		name = t.DisplayName()
	} else {
		t, ok := user.LookupXUID(b.online[b.OnlinePlayer.Options[b.OnlinePlayer.Value()]])
		if !ok {
			u.Message("command.target.unknown")
			return
		}
		if t.Roles().Contains(role.Operator{}) {
			u.Message("command.ban.operator`")
			return
		}

		tP := t.Player()
		t.SetBan(punishment)
		tP.Disconnect(strcenter.CenterLine(strings.Join([]string{
			lang.Translatef(tP.Locale(), "user.ban.header"),
			lang.Translatef(tP.Locale(), "user.ban.description", reason, durafmt.ParseShort(length)),
		}, "\n")))
		name = t.Player().Name()
	}

	user.Alert(p, "staff.alert.ban", name, reason)
	user.Broadcast("command.ban.broadcast", p.Name(), name, reason)
	webhook.SendPunishment(p.Name(), name, reason, "Ban")
	u.Message("command.ban.success", name, reason)
}
