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

func (p *playerPayload) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing player payload: %v", err)
	}
	return b, nil
}

type Player interface {
	MaxPlayer() int
	Create(name string) (string, error)
	Join(token, name string) (player.Player, error)
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
