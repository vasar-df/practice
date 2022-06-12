package lobby

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/vasar-network/practice/vasar/board"
	"github.com/vasar-network/practice/vasar/game/kit"
	"github.com/vasar-network/practice/vasar/user"
	"github.com/vasar-network/vails/sets"
	"sync"
)

// providers maps between a *player.Player and a *Provider.
var providers sync.Map

// LookupProvider looks up the *Provider of the *player.Player passed.
func LookupProvider(p *player.Player) (*Provider, bool) {
	h, ok := providers.Load(p)
	if ok {
		return h.(*Provider), ok
	}
	return nil, false
}

// lobby ...
var lobby *Provider

// Lobby ...
func Lobby() *Provider {
	return lobby
}

// Provider is the provider for the lobby.
type Provider struct {
	w *world.World

	playerMu sync.Mutex
	players  sets.Set[*player.Player]
}

// NewProvider creates a new lobby provider.
func NewProvider(w *world.World) *Provider {
	lobby = &Provider{
		w:       w,
		players: make(sets.Set[*player.Player]),
	}
	return lobby
}

// AddPlayer ...
func (s *Provider) AddPlayer(p *player.Player) {
	providers.Store(p, s)
	p.Inventory().Handle(inventoryHandler{})

	kit.Apply(kit.Lobby{}, p)

	s.playerMu.Lock()
	s.players.Add(p)
	s.playerMu.Unlock()

	if p.World() != s.w {
		s.w.AddEntity(p)
	}
	p.Teleport(s.w.Spawn().Vec3Middle())

	if u, ok := user.Lookup(p); ok {
		yaw, pitch := p.Rotation()
		u.Rotate(180-yaw, -pitch)

		u.EnableProjectiles()
		if u.PearlCoolDown() {
			u.TogglePearlCoolDown()
		}
		if u.Tagged() {
			u.RemoveTag()
		}
		u.SetBoard(s)
	}

	players := s.Players()
	for _, otherP := range players {
		if otherU, ok := user.Lookup(otherP); ok {
			otherU.Board().SendScoreboard(otherP)
		}
	}
}

// RemovePlayer ...
func (s *Provider) RemovePlayer(p *player.Player, force bool) {
	providers.Delete(p)
	p.Inventory().Handle(nil)

	s.playerMu.Lock()
	s.players.Delete(p)
	s.playerMu.Unlock()

	if u, ok := user.Lookup(p); ok {
		u.SetBoard(board.NopProvider{})
	}
	if !force {
		for _, otherP := range s.Players() {
			if otherU, ok := user.Lookup(otherP); ok {
				otherU.Board().SendScoreboard(otherP)
			}
		}
	}
}

// Players ...
func (s *Provider) Players() []*player.Player {
	s.playerMu.Lock()
	defer s.playerMu.Unlock()
	return s.players.Values()
}

// PlayerCount ...
func (s *Provider) PlayerCount() int {
	s.playerMu.Lock()
	defer s.playerMu.Unlock()
	return len(s.players)
}

// SendScoreboard ...
func (s *Provider) SendScoreboard(p *player.Player) {
	users := user.Count()
	board.Send(p, "scoreboard.lobby", users, users-s.PlayerCount())
}
