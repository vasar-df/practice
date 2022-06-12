package user

import (
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SendSound sends a built-in sound to the user with a custom position.
func (u *User) SendSound(pos mgl64.Vec3, sound world.Sound) {
	u.s.ViewSound(pos, sound)
}

// SendCustomParticle sends a custom particle to the user, or it's viewers as well.
func (u *User) SendCustomParticle(id, data int32, pos mgl64.Vec3, public bool) {
	pk := &packet.LevelEvent{
		EventType: packet.LevelEventParticleLegacyEvent | id,
		EventData: data,
		Position:  vec64To32(pos),
	}

	viewers := []world.Viewer{u.s}
	if public {
		viewers = u.viewers()
	}

	for _, v := range viewers {
		if s, ok := v.(*session.Session); ok {
			session_writePacket(s, pk)
		}
	}
}

// SendCustomSound sends a custom sound to the user, or it's viewers as well.
func (u *User) SendCustomSound(sound string, volume, pitch float64, public bool) {
	pos := u.Player().Position()
	pk := &packet.PlaySound{
		SoundName: sound,
		Position:  vec64To32(pos),
		Volume:    float32(volume),
		Pitch:     float32(pitch),
	}

	viewers := []world.Viewer{u.s}
	if public {
		viewers = u.viewers()
	}

	for _, v := range viewers {
		if s, ok := v.(*session.Session); ok {
			session_writePacket(s, pk)
		}
	}
}
