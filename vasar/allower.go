package vasar

import (
	"github.com/hako/durafmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/unickorn/strcenter"
	"github.com/upper/db/v4"
	"github.com/vasar-network/practice/vasar/data"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"golang.org/x/text/language"
	"net"
	"strings"
)

// allower ensures that all players who join are whitelisted if whitelisting is enabled.
type allower struct {
	v *Vasar
}

// Allow ...
func (a *allower) Allow(_ net.Addr, identity login.IdentityData, client login.ClientData) (string, bool) {
	l, _ := language.Parse(strings.Replace(client.LanguageCode, "_", "-", 1))
	users, err := data.SearchOfflineUsers(db.Or(db.Cond{"did": client.DeviceID}, db.Cond{"ssid": client.SelfSignedID}, db.Cond{"xuid": identity.XUID}))
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		if !u.Ban.Expired() {
			reason := strings.TrimSpace(u.Ban.Reason)
			if u.Ban.Permanent {
				description := lang.Translatef(l, "user.blacklist.description", reason)
				if u.XUID() == identity.XUID {
					return strcenter.CenterLine(lang.Translatef(l, "user.blacklist.header") + "\n" + description), false
				}
				return strcenter.CenterLine(lang.Translatef(l, "user.blacklist.header.alt") + "\n" + description), false
			}
			description := lang.Translatef(l, "user.ban.description", reason, durafmt.ParseShort(u.Ban.Remaining()))
			if u.XUID() == identity.XUID {
				return strcenter.CenterLine(lang.Translatef(l, "user.ban.header") + "\n" + description), false
			}
			return strcenter.CenterLine(lang.Translatef(l, "user.ban.header.alt") + "\n" + description), false
		}
	}

	if a.v.config.Vasar.Whitelisted {
		u, err := data.LoadOfflineUser(identity.DisplayName)
		if err != nil {
			return strcenter.CenterLine(lang.Translatef(l, "user.server.whitelist")), false
		}
		return strcenter.CenterLine(lang.Translatef(l, "user.server.whitelist")), u.Whitelisted || u.Roles.Contains(role.Trial{}, role.Operator{})
	}
	return "", true
}
