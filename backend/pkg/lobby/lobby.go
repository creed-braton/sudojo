package lobby

import (
	"errors"
	"sort"
	"sudojo/pkg/game"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
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
	// Returns updated list of player state if successful. Returns
	// ErrPlayerNotFound if no player is associated with the token or
	// game.ErrFinished if the game is already finished.
	Join(token string) ([]Player, error)
	// Sets the player associated with the provided token to inactive.
	// Returns updated list of player state if successful. Returns
	// ErrPlayerNotFound if no player is associated with the token.
	Leave(token string) ([]Player, error)
	// Returns a list of all player names in the lobby with their
	// and whether they are active or not.
	Players() []Player
	// Whether the lobby is in strict mode or not.
	Strict() bool
	// Maximum amount of players the lobby can hold.
	Size() int
}

type lobby struct {
	id      string
	game    game.Game
	players map[string]Player
	strict  bool
	size    int
	lock    sync.RWMutex
}

var _ Lobby = &lobby{}

// Returns a lobby with the provided states and settings.
func New(
	id string,
	game game.Game,
	players map[string]Player,
	strict bool,
	size int,
) *lobby {
	return &lobby{
		id:      id,
		game:    game,
		players: players,
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
		players: make(map[string]Player, size),
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

func (l *lobby) Create(name string) (string, error) {
	if l.game.Finished() != nil {
		return "", game.ErrFinished
	}

	if err := validName(name); err != nil {
		return "", err
	}
	token := newToken()
	player := NewPlayer(token, name)

	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.players) >= l.size {
		return "", ErrLobbyFull
	}

	l.players[token] = player
	return token, nil
}

// Returns all players sorted alphabetically by token (caller must hold lock).
func (l *lobby) sortedPlayers() []Player {
	tokens := make([]string, 0, len(l.players))
	for token := range l.players {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)

	players := make([]Player, 0, len(l.players))
	for _, token := range tokens {
		players = append(players, l.players[token])
	}

	return players
}

func (l *lobby) Join(token string) ([]Player, error) {
	if l.game.Finished() != nil {
		return nil, game.ErrFinished
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	player, exist := l.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(true)

	return l.sortedPlayers(), nil
}

func (l *lobby) Leave(token string) ([]Player, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	player, exist := l.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(false)

	return l.sortedPlayers(), nil
}

func (l *lobby) Players() []Player {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.sortedPlayers()
}

func (l *lobby) Strict() bool {
	return l.strict
}

func (l *lobby) Size() int {
	return l.size
}
