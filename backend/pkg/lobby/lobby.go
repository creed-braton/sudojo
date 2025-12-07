package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
)

var (
	ErrNameTooLong    = errors.New("player name too long")
	ErrInvalidChar    = errors.New("player name contains invalid character")
	ErrLobbyFull      = errors.New("lobby is already full")
	ErrPlayerNotFound = errors.New("player not found")
)

// Holds game state, player state and metadata.
type Lobby interface {
	// Returns the ID of the lobby.
	Id() string
	// Returns the shared game state.
	Game() game.Game
	// Creates a player with provided name and returns the generated
	// token. Returns an error if lobby is full or name is invalid.
	Create(name string) (string, error)
	// Sets the player associated with the provided token to active.
	// Returns updated list of player state or an error if no player
	// is associated with the token.
	Join(token string) ([]event.Player, error)
	// Sets the player associated with the provided token to inactive.
	// Returns updated list of player state.
	Leave(token string) []event.Player
	// Returns a list of all player names in the lobby with their
	// and whether they are active or not.
	Players() []event.Player
	// Whether the lobby is in strict mode or not.
	Strict() bool
	// Maximum amount of players the lobby can hold.
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

// Returns a lobby with the provided states and settings.
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
		active:  make(map[string]struct{}, size),
		strict:  strict,
		size:    size,
	}
}

// Creates a new lobby with the provided settings.
func Open(strict bool, size int) *lobby {
	now := time.Now().UTC().UnixNano()
	game := game.Generate(now)
	game.Start()
	return &lobby{
		id:      uuid.NewString(),
		game:    game,
		players: make(map[string]string, size),
		active:  make(map[string]struct{}, size),
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

func (l *lobby) Join(token string) ([]event.Player, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if _, ok := l.players[token]; !ok {
		return nil, ErrPlayerNotFound
	}
	l.active[token] = struct{}{}

	players := []event.Player{}
	for token, name := range l.players {
		_, ok := l.active[token]
		players = append(players, event.NewPlayer(name, ok))
	}

	return players, nil
}

func (l *lobby) Leave(token string) []event.Player {
	l.lock.Lock()
	defer l.lock.Unlock()

	delete(l.active, token)
	players := []event.Player{}
	for token, name := range l.players {
		_, ok := l.active[token]
		players = append(players, event.NewPlayer(name, ok))
	}

	return players
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
