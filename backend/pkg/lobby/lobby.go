package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sync"
	"unicode"
)

var (
	ErrNameTooLong    = errors.New("player name too long")
	ErrInvalidChar    = errors.New("player name contains invalid character")
	ErrLobbyFull      = errors.New("lobby is already full")
	ErrPlayerNotFound = errors.New("player not found")
)

type Lobby interface {
	Id() string
	Game() game.Game
	Player(token string) error
	Create(name string) (string, error)
	Join(token string)
	Leave(token string)
	Players() []event.Player
	Strict() bool
	Size() int
}

type lobby struct {
	id      string
	game    game.Game
	players map[string]string
	active  map[string]struct{}
	strict  bool
	size    int
	lock    sync.RWMutex
}

var _ Lobby = &lobby{}

func New(
	id string,
	game game.Game,
	players map[string]string,
	strict bool,
	size int,
) *lobby {
	return &lobby{
		id:      id,
		game:    game,
		players: players,
		active:  make(map[string]struct{}),
		strict:  strict,
		size:    size,
	}
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) Game() game.Game {
	return l.game
}

func (l *lobby) Player(token string) error {
	l.lock.RLock()
	defer l.lock.RUnlock()

	if _, ok := l.players[token]; !ok {
		return ErrPlayerNotFound
	}
	return nil
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

func (l *lobby) Create(name string) (string, error) {
	if l.game.Finished() != nil {
		return "", game.ErrFinished
	}

	if err := validName(name); err != nil {
		return "", err
	}
	token := newToken()

	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.players) >= l.size {
		return "", ErrLobbyFull
	}

	l.players[token] = name
	return token, nil
}

func (l *lobby) Join(token string) {
	l.active[token] = struct{}{}
}

func (l *lobby) Leave(token string) {
	delete(l.active, token)
}

func (l *lobby) Players() []event.Player {
	l.lock.RLock()
	defer l.lock.RUnlock()

	players := []event.Player{}
	for token, name := range l.players {
		_, ok := l.active[token]
		players = append(players, event.NewPlayer(name, ok))
	}

	return players
}

func (l *lobby) Strict() bool {
	return l.strict
}

func (l *lobby) Size() int {
	return l.size
}
