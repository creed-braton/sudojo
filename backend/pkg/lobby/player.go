package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync/atomic"
	"unicode"
)

var (
	ErrNameTooLong = errors.New("player name too long")
	ErrInvalidChar = errors.New("player name contains invalid character")
)

// Represents a player with token, name and activity status in associated lobby.
type Player interface {
	// Secret token used for player authentification.
	Token() string
	// In-game name of the player.
	Name() string
	// Activity status of the player.
	Active() bool
	// Sets activity status of the player.
	SetActive(active bool)
}

type player struct {
	token  string
	name   string
	active atomic.Bool
}

var _ Player = &player{}

// Initializes a player instance with the given token and name.
func NewPlayer(token, name string) *player {
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
