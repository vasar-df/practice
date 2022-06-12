package vasar

import (
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

// Provider is a dummy player data provider to change the direction the player is looking towards.
type Provider struct {
	srv *server.Server
	player.NopProvider
}

// Load ...
func (p *Provider) Load(uuid.UUID) (player.Data, error) {
	return player.Data{
		Position:        p.srv.World().Spawn().Vec3Middle(),
		GameMode:        world.GameModeSurvival,
		Yaw:             180,
		Health:          20,
		MaxHealth:       20,
		Hunger:          20,
		SaturationLevel: 5,
	}, nil
}
