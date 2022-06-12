package user

import (
	"github.com/vasar-network/practice/vasar/game"
)

// Stats contains all the stats of a user.
type Stats struct {
	// Elo is the amount of elo that the user has altogether.
	Elo int32 `bson:"elo"`
	// GameElo maps between a game and the amount of elo that the user has in that game.
	GameElo map[string]int32 `bson:"game_elo"`

	// Kills is the amount of players the user has killed.
	Kills uint32 `bson:"kills"`
	// Deaths is the amount of times the user has died.
	Deaths uint32 `bson:"deaths"`
	// RankedWins is the amount of ranked wins the user has.
	UnrankedWins uint32 `bson:"unranked_wins"`
	// RankedLosses is the amount of unranked losses the user has.
	UnrankedLosses uint32 `bson:"unranked_losses"`
	// RankedWins is the amount of ranked wins the user has.
	RankedWins uint32 `bson:"ranked_wins"`
	// RankedLosses is the amount of ranked losses the user has.
	RankedLosses uint32 `bson:"ranked_losses"`

	// KillStreak is the current streak of kills the user has without dying.
	KillStreak uint32 `bson:"kill_streak"`
	// BestKillStreak is the highest kill-streak the user has ever gotten.
	BestKillStreak uint32 `bson:"best_kill_streak"`
}

// defaultElo is the default elo all players start with.
const defaultElo = 1000

// DefaultStats returns the default stats of a user.
func DefaultStats() Stats {
	s := Stats{
		Elo:     defaultElo,
		GameElo: make(map[string]int32, len(game.Games())),
	}
	for _, g := range game.Games() {
		s.GameElo[g.String()] = defaultElo
	}
	return s
}
