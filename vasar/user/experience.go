package user

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
)

// ResetExperienceProgress resets the user's experience progress.
func (u *User) ResetExperienceProgress() {
	u.SendExperienceProgress(0, 0)
}

// SendExperienceProgress sends the experience progress to the user.
func (u *User) SendExperienceProgress(level int, progress float64) {
	session_writePacket(u.s, &packet.UpdateAttributes{
		EntityRuntimeID: 1,
		Attributes: []protocol.Attribute{
			{
				Name:  "minecraft:player.level",
				Value: float32(level),
				Max:   float32(math.MaxInt32),
			},
			{
				Name:  "minecraft:player.experience",
				Value: float32(progress),
				Max:   1.0,
			},
		},
	})
}
