package user

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/vasar-network/vails/lang"
	"time"
)

// Tagged checks if the player is in combat.
func (u *User) Tagged() bool {
	return time.Now().Before(u.TagExpiration())
}

// TagExpiration returns the expiration time of the combat time.
func (u *User) TagExpiration() time.Time {
	u.tagMu.Lock()
	defer u.tagMu.Unlock()
	return u.tagExpiration
}

// Attacker returns the last attacker of the player.
func (u *User) Attacker() *player.Player {
	u.tagMu.Lock()
	defer u.tagMu.Unlock()
	return u.attacker
}

// Tag starts the combat tag and notifies the player if specified.
func (u *User) Tag(attacker *player.Player, kill, notify bool) {
	u.tagMu.Lock()
	defer u.tagMu.Unlock()

	now := time.Now()
	seconds := time.Second * 20
	if kill && u.tagExpiration.Sub(now) > time.Second*5 {
		seconds /= 5
		if notify {
			u.p.SendTitle(title.New().WithActionText(lang.Translatef(u.p.Locale(), "combat.tag.reduced")))
		}
	}
	if attacker != nil {
		u.attacker = attacker
	}

	if !now.Before(u.tagExpiration) {
		if notify {
			u.p.SendTitle(title.New().WithActionText(lang.Translatef(u.p.Locale(), "combat.tag.added")))
		}

		go func() {
			t := time.NewTicker(time.Second)
			defer t.Stop()
			for {
				select {
				case <-u.tagC:
					return
				case <-t.C:
					if !u.Tagged() {
						u.tagMu.Lock()
						if notify {
							u.p.SendTitle(title.New().WithActionText(lang.Translatef(u.p.Locale(), "combat.tag.expired")))
						}
						u.attacker = nil
						u.tagMu.Unlock()
						return
					}
				}
			}
		}()
	}

	u.tagExpiration = now.Add(seconds)
}

// RemoveTag removes the existing combat tag without notifying the user.
func (u *User) RemoveTag() {
	u.tagMu.Lock()
	u.tagC <- struct{}{}
	u.tagExpiration = time.Time{}
	u.attacker = nil
	u.tagMu.Unlock()
}
