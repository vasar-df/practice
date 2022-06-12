package match

import (
	"sync"
)

// GridPos is the position of a node on the grid. It is composed of two integers.
type GridPos [2]int

// Add adds two GridPos' together.
func (p GridPos) Add(o GridPos) GridPos {
	return GridPos{p[0] + o[0], p[1] + o[1]}
}

// X returns the X coordinate of the node position. It is equivalent to GridPos[0].
func (p GridPos) X() int {
	return p[0]
}

// Y returns the Y coordinate of the node position. It is equivalent to GridPos[1].
func (p GridPos) Y() int {
	return p[1]
}

// Grid is a grid that keeps track of arena placements on a 2D level. It contains all the open positions on the grid,
// and the positions that are occupied by arenas.
type Grid struct {
	nodeMu sync.Mutex
	nodes  map[GridPos]struct{}
}

// NewGrid initializes a new grid with the provided step.
func NewGrid() *Grid {
	return &Grid{nodes: map[GridPos]struct{}{}}
}

// Next returns the next open node position on the grid. The second return value will be false if no open
// node can be found.
func (g *Grid) Next() GridPos {
	g.nodeMu.Lock()
	defer g.nodeMu.Unlock()

	if len(g.nodes) == 0 {
		// Return the origin.
		return GridPos{}
	}

	var (
		pos GridPos
		l   int
		d   = GridPos{0, -1}
	)
	for {
		if _, ok := g.nodes[pos]; !ok {
			return pos
		}

		if pos[0] == pos[1] || (pos[0] < 0 && pos[0] == -pos[1]) || (pos[0] > 0 && pos[0] == 1-pos[1]) {
			l = d[0]
			d[0] = -d[1]
			d[1] = l
		}

		pos = pos.Add(d)
	}
}

// Close closes the node at the provided position.
func (g *Grid) Close(pos GridPos) {
	g.nodeMu.Lock()
	defer g.nodeMu.Unlock()
	g.nodes[pos] = struct{}{}
}

// Open opens the node at the provided position.
func (g *Grid) Open(pos GridPos) {
	g.nodeMu.Lock()
	defer g.nodeMu.Unlock()
	delete(g.nodes, pos)
}
