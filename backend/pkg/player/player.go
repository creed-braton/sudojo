package player

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"unicode"
)

var (
	ErrNameTooLong    = errors.New("player name too long")
	ErrInvalidChar    = errors.New("player name contains invalid character")
	ErrPoolFull       = errors.New("player pool is already full")
	ErrPlayerNotFound = errors.New("player not found")
)

// Represents a player holding metadata.
type Player interface {
	// Secret token of the player also used for identification.
	Token() string
	// In-game name of the player.
	Name() string
	// Activity status of the player.
	Active() bool
	// Set activity status of the player.
	SetActive(a bool)
}

type player struct {
	token  string
	name   string
	active atomic.Bool
}

var _ Player = &player{}

// Creates a new player with the given token and name.
func New(token, name string) *player {
	return &player{token: token, name: name}
}

func (p *player) Token() string {
	return p.token
}

func (p *player) Name() string {
	return p.name
}

func (p *player) Active() bool {
	return p.active.Load()
}

func (p *player) SetActive(a bool) {
	p.active.Store(a)
}

// Generates a random 16 bytes hex string.
func newToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Validates the length and characters of the name. Returns ErrNameTooLong
// or ErrInvalidChar if name is invalid.
func validName(name string) error {
	if len(name) > 16 {
		return ErrNameTooLong
	}
	for _, c := range name {
		if unicode.IsDigit(c) || unicode.IsLetter(c) {
			continue
		}
		if string(c) != "-" && string(c) != "_" {
			return ErrInvalidChar
		}
	}
	return nil
}

// Controls player allocation, validates capacity, and updates
// player activity when joining or leaving.
type Pool interface {
	// Returns the maximum number of players allowed in the pool.
	Size() int
	// Creates a new inactive player with the given name and returns
	// his secret token for authentication. Returns an error if the
	// name is invalid or the pool is full.
	Create(name string) (string, error)
	// Marks the player as active and registers its event bus.
	// Returns all players in sorted order or an error if not found.
	Join(token string) ([]Player, error)
	// Marks the player as inactive and removes it from the fanout.
	// Returns all players in sorted order or an error if not found.
	Leave(token string) ([]Player, error)
}

type pool struct {
	size    int
	players map[string]Player
	lock    sync.RWMutex
}

var _ Pool = &pool{}

func NewPool(init map[string]string, size int) *pool {
	players := make(map[string]Player, size)
	for token, name := range init {
		players[token] = New(token, name)
	}
	return &pool{size: size, players: players}
}

func (p *pool) Size() int {
	return p.size
}

func (p *pool) Create(name string) (string, error) {
	if err := validName(name); err != nil {
		return "", err
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.size <= len(p.players) {
		return "", ErrPoolFull
	}

	token := newToken()
	player := New(token, name)
	p.players[token] = player
	return token, nil
}

func (p *pool) sortedPlayers() []Player {
	tokens := make([]string, 0, len(p.players))
	for token := range p.players {
		tokens = append(tokens, token)
	}

	sort.Strings(tokens)

	sorted := make([]Player, 0, len(tokens))
	for _, token := range tokens {
		sorted = append(sorted, p.players[token])
	}

	return sorted
}

func (p *pool) Join(token string) ([]Player, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	player, exist := p.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(true)

	return p.sortedPlayers(), nil
}

func (p *pool) Leave(token string) ([]Player, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	player, exist := p.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(false)

	return p.sortedPlayers(), nil
}
