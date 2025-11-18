package lobby

import (
	"errors"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLobbyFull = errors.New("lobby is already full")
)

// Manages game sessions, coordinates player actions, and enforces game rules
// by delegating to game and player pool components.
type Lobby interface {
	// Unique identifier of the lobby.
	Id() string
	// Complete game state of the lobby.
	Game() game.Game
	// Maximum number of players allowed in the lobby.
	MaxPlayer() int
	// Flag wheter lobby is in strict mode.
	Strict() bool
	// Creates a new inactive player with the given name and returns
	// the secret token for authentication. Returns ErrLobbyFull if not
	// successful.
	Create(name string) (string, error)
	// Marks the player as active and returns the updated player list
	// as payload. Returns player.ErrPlayerNotFound if not successful.
	Join(token string) (event.Payload, error)
	// Marks the player as inactive and returns the updated player list
	// as payload. Returns player.ErrPlayerNotFound if not successful.
	Leave(token string) (event.Payload, error)
	// Inserts a value into the board at the specified position according
	// to the lobby's mode. Returns a payload with the updated board state
	// and any conflicts, or an error if the operation fails.
	Insert(row, col, val int) (event.Payload, error)
	// Validates the cell position and returns a payload with the coordinates.
	// Returns game.ErrOutOfBounds if not successful.
	Ping(row, col int) (event.Payload, error)
	// Returns a payload containing the current and initial board states.
	State() event.Payload
}

type lobby struct {
	id     string
	strict bool
	game   game.Game
	pool   player.Pool
}

var _ Lobby = &lobby{}

// Returns a new lobby with the provided id, mode, game, and player pool.
func New(id string, strict bool, game game.Game, pool player.Pool) *lobby {
	return &lobby{
		id:     id,
		strict: strict,
		game:   game,
		pool:   pool,
	}
}

// Creates and starts a new lobby with a generated game, empty player pool,
// and unique identifier.
func Open(maxPlayer int, strict bool) *lobby {
	game := game.Generate(time.Now().UTC().UnixNano())
	game.Start()
	pool := player.NewPool(make(map[string]string), maxPlayer)
	return New(uuid.NewString(), strict, game, pool)
}

func (l *lobby) Id() string {
	return l.id
}

func (l *lobby) Game() game.Game {
	return l.game
}

func (l *lobby) MaxPlayer() int {
	return l.pool.Size()
}

func (l *lobby) Strict() bool {
	return l.strict
}

func (l *lobby) Create(name string) (string, error) {
	token, err := l.pool.Create(name)
	if err == player.ErrPoolFull {
		return token, ErrLobbyFull
	}
	return token, err
}

func (l *lobby) Join(token string) (event.Payload, error) {
	players, err := l.pool.Join(token)
	if err != nil {
		return nil, err
	}

	return event.NewPayload().SetPlayers(players), nil
}

func (l *lobby) Leave(token string) (event.Payload, error) {
	players, err := l.pool.Leave(token)
	if err != nil {
		return nil, err
	}

	return event.NewPayload().SetPlayers(players), nil
}

func (l *lobby) Insert(row, col, val int) (event.Payload, error) {
	payload := event.NewPayload().SetRow(row).SetColumn(col).SetValue(val)

	var current sudoku.Sudoku
	var err error
	if l.strict {
		current, err = l.game.Strict(row, col, val)
		if err == game.ErrIncorrect {
			payload.SetConflict(err.Error())
			err = nil
		}
	} else {
		current, err = l.game.Lax(row, col, val)
		if err == game.ErrRowConflict ||
			err == game.ErrColConflict ||
			err == game.ErrBoxConflict {
			payload.SetConflict(err.Error())
			err = nil
		}
	}
	payload.SetCurrent(current)

	return payload, err
}

func (l *lobby) Ping(row, col int) (event.Payload, error) {
	payload := event.NewPayload().SetRow(row).SetColumn(col)

	if !sudoku.ValidBounds(row, col) {
		return payload, game.ErrOutOfBounds
	}

	return payload, nil
}

func (l *lobby) State() event.Payload {
	return event.NewPayload().
		SetCurrent(l.game.Current()).
		SetInitial(l.game.Initial())
}
