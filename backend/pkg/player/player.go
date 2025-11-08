package player

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sudojo/pkg/event"
	"unicode"
)

var (
	ErrNameTooLong = errors.New("player name too long")
	ErrInvalidChar = errors.New("player name contains invalid character")
)

// Represents a player holding metadata and his event bus with events
// directed towards him.
type Player interface {
	// Secret token of the player also used for identification.
	Token() string
	// In-game name of the player.
	Name() string
	// Puts event in the players event bus. Returns event.ErrClosedBus if the
	// event bus is closed and event.ErrFullBus if the buffer is full.
	Send(e event.Event) error
	// Returns an event directed to the player or ErrClosedBus if his event
	// bus is closed.
	Receive() (event.Event, error)
	// Closes the event bus of the player.
	Close()
}

type player struct {
	bus   event.EventBus
	token string
	name  string
}

var _ Player = &player{}

// Generates a random 16 bytes hex string.
func NewToken() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Returns a new player with the provided token, name and event bus.
func New(token, name string, bus event.EventBus) *player {
	return &player{token: token, name: name, bus: bus}
}

func (p *player) Token() string {
	return p.token
}

func (p *player) Name() string {
	return p.name
}

func (p *player) Send(e event.Event) error {
	return p.bus.Send(e)
}

func (p *player) Receive() (event.Event, error) {
	return p.bus.Receive()
}

func (p *player) Close() {
	p.bus.Close()
}

// Validates the length and characters of the name. Returns ErrNameTooLong
// or ErrInvalidChar if name is invalid.
func ValidName(name string) error {
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
