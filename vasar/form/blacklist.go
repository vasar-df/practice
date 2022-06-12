package form

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
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

// blacklist is a form that allows a user to issue a blacklist.
type blacklist struct {
	// Reason is a dropdown that allows a user to select a blacklist reason.
	Reason form.Input
	// OnlinePlayer is a dropdown that allows a user to select an online player.
	OnlinePlayer form.Dropdown
	// OfflinePlayer is an input field that allows a user to enter an offline player.
	OfflinePlayer form.Input
	// online is a list of online players' XUIDs indexed by their names.
	online map[string]string
}

// NewBlacklist creates a new form to issue a blacklist.
func NewBlacklist() form.Form {
	online := make(map[string]string)
	for _, u := range user.All() {
		online[u.Player().Name()] = u.Player().XUID()
	}
	names := [...]string{"Steve Harvey", "Elon Musk", "Bill Gates", "Mark Zuckerberg", "Jeff Bezos", "Warren Buffet", "Larry Page", "Sergey Brin", "Larry Ellison", "Tim Cook", "Steve Ballmer", "Daniel Larson", "Steve"}
	list := maps.Keys(online)
	sort.Strings(list)
	return form.New(blacklist{
		Reason:        form.NewInput("Reason", "", "Enter a reason for the blacklist."),
		OnlinePlayer:  form.NewDropdown("Online Player", list, 0),
		OfflinePlayer: form.NewInput("Offline Player", "", names[rand.Intn(len(names)-1)]),
		online:        online,
	}, "Blacklist")
}

// Submit ...
func (b blacklist) Submit(s form.Submitter) {
	p := s.(*player.Player)
	u, ok := user.Lookup(p)
	if !ok {
		// User somehow left midway through the form.
		return
	}
	if !u.Roles().Contains(role.Manager{}, role.Operator{}) {
		// In case the user's role was removed while the form was open.
		return
	}
	reason := strings.TrimSpace(b.Reason.Value())
	if len(reason) == 0 {
		reason = "None"
	}

	punishment := user.Punishment{
		Staff:      p.Name(),
		Reason:     reason,
		Occurrence: time.Now(),
		Permanent:  true,
	}
	var name string
	if offlineName := strings.TrimSpace(b.OfflinePlayer.Value()); offlineName != "" {
		if strings.EqualFold(offlineName, p.Name()) {
			u.Message("command.blacklist.self")
			return
		}
		t, err := data.LoadOfflineUser(offlineName)
		if err != nil {
			u.Message("command.target.unknown")
			return
		}
		if t.Roles.Contains(role.Operator{}) {
			u.Message("command.blacklist.operator")
			return
		}
		if !t.Ban.Expired() && t.Ban.Permanent {
			u.Message("command.blacklist.already")
			return
		}
		t.Ban = punishment
		name = t.DisplayName()
		_ = data.SaveOfflineUser(t)
	} else {
		t, ok := user.LookupXUID(b.online[b.OnlinePlayer.Options[b.OnlinePlayer.Value()]])
		if !ok {
			u.Message("command.target.unknown")
			return
		}
		if t.Roles().Contains(role.Operator{}) {
			u.Message("command.blacklist.operator")
			return
		}

		tP := t.Player()
		t.SetBan(punishment)
		tP.Disconnect(strcenter.CenterLine(strings.Join([]string{
			lang.Translatef(tP.Locale(), "user.blacklist.header"),
			lang.Translatef(tP.Locale(), "user.blacklist.description", reason),
		}, "\n")))
		name = tP.Name()
	}

	user.Alert(p, "staff.alert.blacklist", name)
	user.Broadcast("command.ban.broadcast", p.Name(), name, reason)
	webhook.SendPunishment(p.Name(), name, reason, "Blacklist")
	u.Message("command.blacklist.success", name, reason)
}
