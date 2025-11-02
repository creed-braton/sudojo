package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"unicode"
)

var (
	ErrLobbyFull   = errors.New("lobby is already full")
	ErrPlayerMiss  = errors.New("player does not exist")
	ErrNameTooLong = errors.New("player name too long")
	ErrNameChar    = errors.New("player name contains invalid character")
)

type Player struct {
	Token  string
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type PlayerManager struct {
	Players   map[string]*Player
	MaxPlayer int
	lock      sync.RWMutex
}

type playerPayload struct {
	Players []Player
}

func newPlayerManager(maxPlayer int) *PlayerManager {
	return &PlayerManager{
		Players:   make(map[string]*Player),
		MaxPlayer: maxPlayer,
		lock:      sync.RWMutex{},
	}
}

func (m *PlayerManager) player(token string) (*Player, error) {
	p := m.Players[token]
	if p == nil {
		return nil, ErrPlayerMiss
	}
	return p, nil
}

func validName(name string) error {
	if len(name) > 16 {
		return ErrNameTooLong
	}
	for _, c := range name {
		if unicode.IsDigit(c) || unicode.IsLetter(c) {
			continue
		}
		if string(c) != "-" && string(c) != "_" && string(c) != " " {
			return ErrNameChar
		}
	}
	return nil
}

func newToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (m *PlayerManager) state() *playerPayload {
	players := make([]Player, 0, len(m.Players))
	for _, p := range m.Players {
		players = append(players, *p)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Token < players[j].Token
	})

	return &playerPayload{
		Players: players,
	}
}

func (m *PlayerManager) create(name string) (string, error) {
	if err := validName(name); err != nil {
		return "", err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if len(m.Players) >= m.MaxPlayer {
		return "", ErrLobbyFull
	}

	p := &Player{
		Token:  newToken(),
		Name:   name,
		Active: false,
	}
	m.Players[p.Token] = p

	return p.Token, nil
}

func (m *PlayerManager) join(token string) (*playerPayload, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if p, err := m.player(token); err == nil {
		p.Active = true
		return m.state(), nil
	} else {
		return nil, err
	}
}

func (m *PlayerManager) leave(token string) (*playerPayload, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if p, err := m.player(token); err == nil {
		p.Active = false
		return m.state(), nil
	} else {
		return nil, err
	}
}
