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

type Lobby interface {
	Id() string
	MaxPlayer() int
	Strict() bool
	Create(name string) (string, error)
	Join(token string) (event.Payload, error)
	Leave(token string) (event.Payload, error)
	Insert(row, col, val int) (event.Payload, error)
	Ping(row, col int) (event.Payload, error)
	State() event.Payload
}

type lobby struct {
	id     string
	pool   player.Pool
	strict bool
	game   game.Game
}

var _ Lobby = &lobby{}

func New(id string, strict bool, game game.Game, pool player.Pool) *lobby {
	return &lobby{
		id:     id,
		pool:   pool,
		strict: strict,
		game:   game,
	}
}

func Open(maxPlayer int, strict bool) *lobby {
	game := game.Generate(time.Now().UTC().UnixNano())
	game.Start()
	pool := player.NewPool(make(map[string]string), maxPlayer)
	return New(uuid.NewString(), strict, game, pool)
}

func (l *lobby) Id() string {
	return l.id
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
