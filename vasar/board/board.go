package board

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/vasar-network/vails/lang"
)

// Provider is an interface for providers that provide scoreboards.
type Provider interface {
	// SendScoreboard will send a scoreboard to the player.
	SendScoreboard(p *player.Player)
}

// NopProvider is a provider that does not provide a scoreboard.
type NopProvider struct{}

// SendScoreboard ...
func (NopProvider) SendScoreboard(*player.Player) {}

// Send sends the specified board type and arguments to the player.
func Send(p *player.Player, board string, args ...any) {
	b := scoreboard.New(lang.Translatef(p.Locale(), "scoreboard.title"))
	_, _ = b.WriteString(lang.Translatef(p.Locale(), board, args...))
	b.RemovePadding()
	p.SendScoreboard(b)
}
