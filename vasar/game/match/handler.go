package match

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player"
	"time"
)

// Handler handles events that are called by a match. Implementations of Handler may be used to listen to specific
// events such as when a player quits (Warro) or when a match starts.
type Handler interface {
	// HandlePrepare handles when a match is first initialized.
	HandlePrepare(duration *time.Duration)
	// HandleStart handles when a match starts.
	HandleStart(initial bool)

	// HandleScoreboardUpdate handles when the scoreboard is updated.
	HandleScoreboardUpdate(ctx *event.Context, p *player.Player)

	// HandleUserAdd handles when a user joins a match.
	HandleUserAdd(player *player.Player)
	// HandleUserStartHit handles when a player starts to hit another player. This is used to ensure a hit isn't logged
	// processed if it shouldn't be allowed.
	HandleUserStartHit(attacker, attacked *player.Player, hits int) bool
	// HandleUserHit handles when a player hits another player. This is useful for certain game types, such as boxing,
	// where a winner is determined by the amount of hits a player has.
	HandleUserHit(attacker, attacked *player.Player)
	// HandleUserRemove handles when a player is removed from a match. The forced boolean indicates if the player was
	// removed due to a timeout or a forced quit. This is useful for certain players, such as Warro.
	HandleUserRemove(ctx *event.Context, player *player.Player)
}

// NopHandler implements the Handler interface but does not execute any code when an event is called. The default handler
// of matches is set to NopHandler. Users may embed NopHandler to avoid having to implement each method.
type NopHandler struct{}

// Compile time check to make sure NopHandler implements Handler.
var _ Handler = (*NopHandler)(nil)

func (NopHandler) HandlePrepare(*time.Duration)                                {}
func (NopHandler) HandleStart(bool)                                            {}
func (NopHandler) HandleScoreboardUpdate(*event.Context, *player.Player)       {}
func (NopHandler) HandleUserAdd(*player.Player)                                {}
func (NopHandler) HandleUserStartHit(*player.Player, *player.Player, int) bool { return true }
func (NopHandler) HandleUserHit(*player.Player, *player.Player)                {}
func (NopHandler) HandleUserRemove(*event.Context, *player.Player)             {}
