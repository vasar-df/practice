package vasar

import (
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/lang"
	"github.com/vasar-network/vails/role"
	"strings"
	"time"
)

// startBroadcasts starts sending a new broadcast every five minutes.
func (v *Vasar) startBroadcasts() {
	broadcasts := []string{
		"vasar.broadcast.discord",
		"vasar.broadcast.store",
		"vasar.broadcast.emojis",
		"vasar.broadcast.settings",
		"vasar.broadcast.duels",
		"vasar.broadcast.feedback",
		"vasar.broadcast.report",
	}

	var cursor int
	t := time.NewTicker(time.Minute * 5)
	defer t.Stop()
	for {
		select {
		case <-v.c:
			return
		case <-t.C:
			message := broadcasts[cursor]
			for _, u := range user.All() {
				u.Message("vasar.broadcast.notice", lang.Translate(u.Player().Locale(), message))
			}

			if cursor++; cursor == len(broadcasts) {
				cursor = 0
			}
		}
	}
}

// startPlayerBroadcasts starts sending a new player broadcast every five minutes.
func (v *Vasar) startPlayerBroadcasts() {
	t := time.NewTicker(time.Minute * 10)
	defer t.Stop()
	for {
		select {
		case <-v.c:
			return
		case <-t.C:
			users := user.All()
			var plus []string
			for _, u := range users {
				if u.Roles().Contains(role.Plus{}) {
					plus = append(plus, u.DisplayName())
				}
			}

			for _, u := range users {
				u.Message("vasar.broadcast.plus", len(plus), strings.Join(plus, ", "))
			}
		}
	}
}
