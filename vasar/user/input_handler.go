package user

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerAuthInputHandler ...
type PlayerAuthInputHandler struct {
	u *User
}

// Handle ...
func (h PlayerAuthInputHandler) Handle(p packet.Packet, s *session.Session) error {
	pk := p.(*packet.PlayerAuthInput)
	set := h.u.Settings()

	var deviceGroup DeviceGroup
	switch pk.InputMode {
	case packet.InputModeMouse:
		deviceGroup = DeviceGroupKeyboardMouse()
		set.Matchmaking.MatchWithKeyboard = true
	case packet.InputModeTouch:
		deviceGroup = DeviceGroupMobile()
		set.Matchmaking.MatchWithMobile = true
	case packet.InputModeGamePad:
		deviceGroup = DeviceGroupController()
		set.Matchmaking.MatchWithController = true
	default:
		return fmt.Errorf("unexpected input mode %d for current device group: %v", pk.InputMode, h.u.DeviceGroup())
	}

	h.u.SetSettings(set)
	h.u.deviceGroup.Store(deviceGroup)
	return (session.PlayerAuthInputHandler{}).Handle(p, s)
}
