package lobby

import (
	"errors"
	"sort"
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/player"
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
	// Returns thread-safe insert history of the lobby.
	History() history.History
	// Creates a player with provided name and returns the generated
	// token. Returns ErrLobbyFull if lobby is full, player.ErrNameTooLong
	// if name is too long and player.ErrInvalidChar if name contains
	// an illegal character.
	Create(name string) (string, error)
	// Sets the player associated with the provided token to active.
	// Returns updated list of player state if successful. Returns
	// ErrPlayerNotFound if no player is associated with the token or
	// game.ErrFinished if the game is already finished.
	Join(token string) ([]player.Player, error)
	// Sets the player associated with the provided token to inactive.
	// Returns updated list of player state if successful. Returns
	// ErrPlayerNotFound if no player is associated with the token.
	Leave(token string) ([]player.Player, error)
	// Returns a list of all player names in the lobby with their
	// and whether they are active or not.
	Players() []player.Player
	// Whether the lobby is in strict mode or not.
	Strict() bool
	// Maximum amount of players the lobby can hold.
	Size() int
}

type lobby struct {
	id      string
	game    game.Game
	players map[string]player.Player
	history history.History
	strict  bool
	size    int
	lock    sync.RWMutex
}

var _ Lobby = &lobby{}

// Returns a lobby instance with the provided states and settings.
func New(
	id string,
	game game.Game,
	players map[string]player.Player,
	history history.History,
	strict bool,
	size int,
) *lobby {
	return &lobby{
		id:      id,
		game:    game,
		players: players,
		history: history,
		strict:  strict,
		size:    size,
	}
}

// Creates a new lobby instance with the provided settings.
func Open(strict bool, size int) *lobby {
	now := time.Now().UTC().UnixNano()
	game := game.Generate(now)
	game.Start()
	return &lobby{
		id:      uuid.NewString(),
		game:    game,
		players: make(map[string]player.Player, size),
		history: history.New([]history.Artifact{}),
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

func (l *lobby) History() history.History {
	return l.history
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

	if err := player.ValidName(name); err != nil {
		return "", err
	}
	token := player.NewToken()
	player := player.New(token, name)

	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.players) >= l.size {
		return "", ErrLobbyFull
	}

	l.players[token] = player
	return token, nil
}

// Returns all players sorted alphabetically by token (caller must hold lock).
func (l *lobby) sortedPlayers() []player.Player {
	tokens := make([]string, 0, len(l.players))
	for token := range l.players {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)

	players := make([]player.Player, 0, len(l.players))
	for _, token := range tokens {
		players = append(players, l.players[token])
	}

	return players
}

func (l *lobby) Join(token string) ([]player.Player, error) {
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

func (l *lobby) Leave(token string) ([]player.Player, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	player, exist := l.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	player.SetActive(false)

	return l.sortedPlayers(), nil
}

func (l *lobby) Players() []player.Player {
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
