package vasar

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/vasar-network/practice/vasar/user"

	_ "unsafe"
)

// PearlHandler ...
type PearlHandler struct{}

// HandleTeleport ...
func (PearlHandler) HandleTeleport(ctx *event.Context, p *player.Player, pos mgl64.Vec3) {
	if u, ok := user.Lookup(p); ok && u.Settings().Visual.PearlAnimation {
		ctx.Cancel()

		yaw, pitch := p.Rotation()
		session_writePacket(player_session(p), &packet.MovePlayer{
			EntityRuntimeID: 1,
			Position:        mgl32.Vec3{float32(pos[0]), float32(pos[1] + 1.621), float32(pos[2])},
			Pitch:           float32(pitch),
			Yaw:             float32(yaw),
			HeadYaw:         float32(yaw),
			Mode:            packet.MoveModeNormal,
		})
		p.Move(pos.Sub(p.Position()), 0, 0)
	}
}

//go:linkname player_session github.com/df-mc/dragonfly/server/player.(*Player).session
//noinspection ALL
func player_session(*player.Player) *session.Session

//go:linkname session_writePacket github.com/df-mc/dragonfly/server/session.(*Session).writePacket
//noinspection ALL
func session_writePacket(*session.Session, packet.Packet)
