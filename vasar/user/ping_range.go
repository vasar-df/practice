package user

import (
	"fmt"
)

// PingRange represents a ping range that a user may match within.
type PingRange struct {
	unrestricted bool
	min, max     int
}

// NewPingRange creates a new PingRange from a base ping.
func NewPingRange(ping int) PingRange {
	return PingRange{min: ping, max: ping}
}

// PingRangeUnrestricted returns the ping group with no restrictions.
func PingRangeUnrestricted() PingRange {
	return PingRange{unrestricted: true}
}

// PingRangeTwentyFive returns the ping group of twenty-five or below.
func PingRangeTwentyFive() PingRange {
	return PingRange{max: 25}
}

// PingRangeFifty returns the ping group of fifty or below.
func PingRangeFifty() PingRange {
	return PingRange{min: 25, max: 50}
}

// PingRangeSeventyFive returns the ping group of seventy-five or below.
func PingRangeSeventyFive() PingRange {
	return PingRange{min: 50, max: 75}
}

// PingRangeHundred returns the ping group of a hundred or below.
func PingRangeHundred() PingRange {
	return PingRange{min: 75, max: 100}
}

// PingRangeHundredTwentyFive returns the ping group of a hundred and twenty-five or below.
func PingRangeHundredTwentyFive() PingRange {
	return PingRange{min: 100, max: 125}
}

// PingRangeHundredFifty returns the ping group of a hundred and fifty or below.
func PingRangeHundredFifty() PingRange {
	return PingRange{min: 125, max: 150}
}

// PingRanges returns all possible ping groups.
func PingRanges() []PingRange {
	return []PingRange{
		PingRangeUnrestricted(),
		PingRangeTwentyFive(),
		PingRangeFifty(),
		PingRangeSeventyFive(),
		PingRangeHundred(),
		PingRangeHundredTwentyFive(),
		PingRangeHundredFifty(),
	}
}

// Unrestricted ...
func (p PingRange) Unrestricted() bool {
	return p.unrestricted
}

// Min ...
func (p PingRange) Min() int {
	return p.min
}

// Max ...
func (p PingRange) Max() int {
	return p.max
}

// String ...
func (p PingRange) String() string {
	if p.unrestricted {
		return "Unrestricted"
	}
	return fmt.Sprintf("%d - %d", p.min, p.max)
}

const (
	// maxPing is the maximum ping value allowed.
	maxPing = 5000
	// minPing is the minimum ping value allowed.
	minPing = 0
)

// Extend ...
func (p PingRange) Extend(min, max int) PingRange {
	if p.unrestricted {
		return p
	}
	if p.max+max > maxPing {
		p.max = maxPing
	} else {
		p.max += max
	}
	if p.min-min < minPing {
		p.min = minPing
	} else {
		p.min -= min
	}
	return p
}

// Compare ...
func (p PingRange) Compare(ping int) bool {
	if p.unrestricted {
		// Unrestricted, so any range is fine.
		return true
	}
	return ping > p.min || ping < p.max
}
