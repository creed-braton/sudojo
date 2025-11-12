package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sudojo/pkg/event"
	"sudojo/pkg/player"
	"sync"
)

var (
	ErrLobbyFull      = errors.New("lobby is already full")
	ErrPlayerNotFound = errors.New("player not found in lobby")
)

type status struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type playerPayload []*status

// Serializes the player payload into bytes.
func (p *playerPayload) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing player payload: %v", err)
	}
	return b, nil
}

// Represents a manager for handling player operations in a lobby including
// creation, joining, and leaving of players.
type Player interface {
	// Returns the maximum number of players allowed.
	MaxPlayer() int
	// Creates a new player slot with the given name. Returns the player token
	// or ErrLobbyFull if maximum capacity is reached.
	Create(name string) (string, error)
	// Activates the player with the given token, registers them in the fanout,
	// and broadcasts a join event. Returns ErrPlayerNotFound if token is invalid.
	Join(token, name string) (player.Player, error)
	// Deactivates the player, deregisters them from the fanout, closes their
	// bus, and broadcasts a leave event. Returns ErrPlayerNotFound if token is invalid.
	Leave(p player.Player) error
}

type playerManager struct {
	players   map[string]*status
	maxPlayer int
	lock      sync.RWMutex
	bus       event.EventBus
	fanout    event.Fanout
}

var _ Player = &playerManager{}

// Returns a new player manager initialized with the provided players map,
// max player limit, event bus, and fanout router.
func NewPlayerManager(
	players map[string]string,
	maxPlayer int,
	bus event.EventBus,
	fanout event.Fanout,
) *playerManager {
	statusMap := make(map[string]*status)
	for token, name := range players {
		statusMap[token] = &status{
			Name:   name,
			Active: false,
		}
	}

	return &playerManager{
		players:   statusMap,
		maxPlayer: maxPlayer,
		bus:       bus,
		fanout:    fanout,
	}
}

// Returns player status sorted alphabetically by token.
func (m *playerManager) sortedPlayers() *playerPayload {
	tokens := make([]string, 0, len(m.players))
	for token := range m.players {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)

	var players playerPayload
	for _, token := range tokens {
		players = append(players, m.players[token])
	}

	return &players
}

func (m *playerManager) MaxPlayer() int {
	return m.maxPlayer
}

func (m *playerManager) Create(name string) (string, error) {
	if err := player.ValidName(name); err != nil {
		return "", err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.players) >= m.maxPlayer {
		return "", ErrLobbyFull
	}

	token := player.NewToken()
	m.players[token] = &status{
		Name:   name,
		Active: false,
	}

	return token, nil
}

func (m *playerManager) Join(token, name string) (player.Player, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	s := m.players[token]
	if s == nil {
		return nil, ErrPlayerNotFound
	}

	s.Active = true
	bus := event.NewEventBus()
	m.fanout.Register(token, bus)
	p := player.New(token, name, bus)

	e := event.New(event.JoinEvent, p.Token(), "", "", m.sortedPlayers())
	err := m.bus.Send(e)

	return p, err
}

func (m *playerManager) Leave(p player.Player) error {
	m.fanout.Deregister(p.Token())
	p.Close()

	m.lock.Lock()
	defer m.lock.Unlock()

	s := m.players[p.Token()]
	if s == nil {
		return ErrPlayerNotFound
	}
	s.Active = false

	e := event.New(event.LeaveEvent, p.Token(), "", "", m.sortedPlayers())

	return m.bus.Send(e)
}
