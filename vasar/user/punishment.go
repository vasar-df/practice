package user

import "time"

// Punishment stores the punishments of a user.
type Punishment struct {
	// Staff is the staff member who issued the punishment.
	Staff string `bson:"staff"`
	// Reason is the reason for the punishment.
	Reason string `bson:"reason"`
	// Occurrence is the time the punishment was issued.
	Occurrence time.Time `bson:"occurrence"`
	// Permanent is true if the punishment doesn't expire.
	Permanent bool `bson:"permanent"`
	// Expiration is the expiration time of the punishment.
	Expiration time.Time `bson:"expiration"`
}

// Remaining returns the remaining duration of the punishment.
func (p Punishment) Remaining() time.Duration {
	return time.Until(p.Expiration).Round(time.Second)
}

// Expired checks if the punishment has expired.
func (p Punishment) Expired() bool {
	return !p.Permanent && time.Now().After(p.Expiration)
}
