package user

import (
	"fmt"
)

// EloRange represents an elo range that a user may match within.
type EloRange struct {
	min, max int
}

// NewEloRange creates a new EloRange from a base elo.
func NewEloRange(elo int) EloRange {
	return EloRange{min: elo, max: elo}
}

// Min ...
func (p EloRange) Min() int {
	return p.min
}

// Max ...
func (p EloRange) Max() int {
	return p.max
}

// String ...
func (p EloRange) String() string {
	return fmt.Sprintf("%d - %d", p.min, p.max)
}

const (
	// maxElo is the maximum elo value allowed.
	maxElo = 5000
	// minElo is the minimum elo value allowed.
	minElo = 0
)

// Extend ...
func (p EloRange) Extend(min, max int) EloRange {
	if p.max+max > maxElo {
		p.max = maxElo
	} else {
		p.max += max
	}
	if p.min-min < minElo {
		p.min = minElo
	} else {
		p.min -= min
	}
	return p
}

// Compare ...
func (p EloRange) Compare(elo int) bool {
	return elo >= p.min && elo <= p.max
}
