package player

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sudojo/pkg/event"
	"sync"
	"sync/atomic"
	"unicode"
)

var (
	ErrNameTooLong    = errors.New("player name too long")
	ErrInvalidChar    = errors.New("player name contains invalid character")
	ErrPoolFull       = errors.New("player pool is already full")
	ErrPlayerNotFound = errors.New("player not found in pool")
)

// Represents a player holding metadata and his event bus with events
// directed towards him.
type Player interface {
	// Secret token of the player also used for identification.
	Token() string
	// In-game name of the player.
	Name() string
	// Activity status of the player.
	Active() bool
	// Set activity status of the player.
	SetActive(active bool)
}

type player struct {
	token  string
	name   string
	active atomic.Bool
}

var _ Player = &player{}

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

func (p *player) SetActive(active bool) {
	p.active.Store(active)
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

type Pool interface {
	MaxSize() int
	Create(name string) error
	Join(token string, pub, sub event.EventChan) ([]Player, error)
	Leave(token string) ([]Player, error)
}

type pool struct {
	maxSize int
	players map[string]Player
	bus     event.EventBus
	lock    sync.RWMutex
}

var _ Pool = &pool{}

func NewPlayerPool(maxSize int, bus event.EventBus) *pool {
	return &pool{
		maxSize: maxSize,
		players: make(map[string]Player, maxSize),
		bus:     bus,
	}
}

func (p *pool) MaxSize() int {
	return p.maxSize
}

func (p *pool) Create(name string) error {
	if err := validName(name); err != nil {
		return err
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.maxSize >= len(p.players) {
		return ErrPoolFull
	}

	token := newToken()
	player := New(token, name)
	p.players[token] = player
	return nil
}

func (p *pool) sortedPlayers() []Player {
	p.lock.RLock()
	defer p.lock.RUnlock()

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

func (p *pool) Join(token string, pub, sub event.EventChan) ([]Player, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	player, exist := p.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(true)
	p.bus.Register(token, pub, sub)

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
	p.bus.Deregister(token)

	return p.sortedPlayers(), nil
}
