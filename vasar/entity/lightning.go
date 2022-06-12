package entity

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Lightning is a lethal element to thunderstorms. Lightning momentarily increases the skylight's brightness to slightly
// greater than full daylight.
type Lightning struct {
	pos atomic.Value[mgl64.Vec3]
}

// NewLightning creates a lightning entity. The lightning entity will be positioned at the position passed.
func NewLightning(pos mgl64.Vec3) *Lightning {
	return &Lightning{
		pos: *atomic.NewValue(pos),
	}
}

// Position returns the current position of the lightning entity.
func (l *Lightning) Position() mgl64.Vec3 {
	return l.pos.Load()
}

// World returns the world that the lightning entity is currently in, or nil if it is not added to a world.
func (l *Lightning) World() *world.World {
	w, _ := world.OfEntity(l)
	return w
}

// BBox ...
func (*Lightning) BBox() cube.BBox {
	return cube.Box(0, 0, 0, 0, 0, 0)
}

// Close closes the lighting.
func (l *Lightning) Close() error {
	l.World().RemoveEntity(l)
	return nil
}

// OnGround ...
func (*Lightning) OnGround() bool {
	return false
}

// Rotation ...
func (*Lightning) Rotation() (yaw, pitch float64) {
	return 0, 0
}

// EncodeEntity ...
func (*Lightning) EncodeEntity() string {
	return "minecraft:lightning_bolt"
}

// Name ...
func (*Lightning) Name() string {
	return "Lightning Bolt"
}
