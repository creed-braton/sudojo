package lobby

import (
	"errors"
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"sync"
)

var (
	ErrLobbyFull      = errors.New("lobby is already full")
	ErrPlayerNotFound = errors.New("player not found")
)

// Represents a multiplayer game session that manages players, game state,
// and insert history.
type Lobby interface {
	// Returns the ID of the lobby.
	Id() string
	// Returns the config of the lobby.
	Config() Config
	// Returns concurrency-safe game state of the lobby.
	Game() game.Game
	// Returns concurrency-safe game history of the lobby.
	History() history.History
	// Returns a list of all players in the lobby with their
	// name and whether they are active or not.
	Players() []player.Player
	// Creates a player with provided name and returns the generated
	// token. Returns ErrLobbyFull if lobby is full, player.ErrNameTooLong
	// or player.ErrInvalidChar if the player name is invalid.
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
	// Inserts the value at the specified cell position and adds the
	// insertion to the lobby history. Returns the new board state if
	// changed, otherwise nil. Returns game.ErrOutOfBounds if row or col
	// is invalid, game.ErrFinished if the game is already finished,
	// game.ErrNotStarted if not yet started, or game.ErrInitialClue if
	// the position is an initial clue. Strict lobbies may additionally
	// return game.ErrStrictValRange or game.ErrIncorrect. Lax lobbies
	// may return game.ErrLaxValRange, game.ErrRowConflict,
	// game.ErrColConflict, or game.ErrBoxConflict.
	Insert(row, col, val int, token string, now int64) (sudoku.Sudoku, error)
}

type lobby struct {
	id      string
	config  Config
	game    game.Game
	history history.History
	players map[string]player.Player
	lock    sync.RWMutex
}

var _ Lobby = &lobby{}

// Creates a new lobby with the specified id, config, game, history,
// and initial players.
func New(
	id string,
	config Config,
	game game.Game,
	history history.History,
	players map[string]player.Player,
) *lobby {
	return &lobby{
		id:      id,
		config:  config,
		game:    game,
		history: history,
		players: players,
	}
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) Config() Config {
	return l.config
}

func (l *lobby) Game() game.Game {
	return l.game
}

func (l *lobby) Players() []player.Player {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return player.Sort(l.players)
}

func (l *lobby) History() history.History {
	return l.history
}

func (l *lobby) Create(name string) (string, error) {
	if l.game.FinishedAt() != nil {
		return "", game.ErrFinished
	}

	if err := player.ValidName(name); err != nil {
		return "", err
	}
	token := player.NewToken()
	player := player.New(token, name)

	l.lock.Lock()
	defer l.lock.Unlock()

	if len(l.players) >= l.config.MaxSize() {
		return "", ErrLobbyFull
	}

	l.players[token] = player
	return token, nil
}

func (l *lobby) Join(token string) ([]player.Player, error) {
	if l.game.FinishedAt() != nil {
		return nil, game.ErrFinished
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	p, exist := l.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	p.SetActive(true)

	return player.Sort(l.players), nil
}

func (l *lobby) Leave(token string) ([]player.Player, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	p, exist := l.players[token]
	if !exist {
		return nil, ErrPlayerNotFound
	}
	p.SetActive(false)

	return player.Sort(l.players), nil
}

func (l *lobby) Insert(row, col, val int, token string, now int64) (sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, game.ErrOutOfBounds
	}

	l.history.Append(history.NewArtifact(row, col, val, token, now))

	if l.config.Strict() {
		return l.Game().Strict(row, col, val, now)
	} else {
		return l.Game().Lax(row, col, val, now)
	}
}
