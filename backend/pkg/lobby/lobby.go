package lobby

import (
	"errors"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLobbyFull = errors.New("lobby is already full")
)

type Lobby interface {
	Close()
	Id() string
	MaxPlayer() int
	Strict() bool
	LastEvent() int64
	Pump() error
	CreatePlayer(name string) (string, error)
	Join(token string, bus event.EventBus) error
	Leave(token string) error
	Insert(row, col, val int, token, trace string) (event.Event, error)
	Ping(row, col int, token, trace string) (event.Event, error)
	State(token, trace string) event.Event
}

type lobby struct {
	id        string
	pool      player.Pool
	strict    bool
	lastEvent atomic.Int64
	game      game.Game
	broadcast event.EventBus
	fanout    event.Fanout
}

var _ Lobby = &lobby{}

func (l *lobby) event() {
	l.lastEvent.Store(time.Now().UTC().UnixNano())
}

func New(
	id string,
	pool player.Pool,
	strict bool,
	game game.Game,
	bus event.EventBus,
	fanout event.Fanout,
) *lobby {
	l := &lobby{
		id:        id,
		pool:      pool,
		strict:    strict,
		game:      game,
		broadcast: bus,
		fanout:    fanout,
	}
	l.event()
	return l
}

func Open(maxPlayer int, strict bool) *lobby {
	bus := event.NewEventBus()
	fanout := event.NewFanout(bus)
	game := game.Generate(time.Now().UTC().UnixNano())
	game.Start()
	pool := player.NewPool(make(map[string]string), maxPlayer)
	return New(uuid.NewString(), pool, strict, game, bus, fanout)
}

func (l *lobby) Close() {
	l.broadcast.Close()
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

func (l *lobby) LastEvent() int64 {
	return l.lastEvent.Load()
}

func (l *lobby) Pump() error {
	return l.fanout.Pump()
}

func (l *lobby) CreatePlayer(name string) (string, error) {
	l.event()

	token, err := l.pool.Create(name)
	if err == player.ErrPoolFull {
		return token, ErrLobbyFull
	}
	return token, err
}

func (l *lobby) Join(token string, bus event.EventBus) error {
	l.event()

	players, err := l.pool.Join(token)
	if err != nil {
		return err
	}

	l.fanout.Register(token, bus)

	payload := event.NewPayload()
	payload.SetPlayers(players)
	if err := l.broadcast.Send(
		event.New(event.JoinEvent, token, "", "", payload),
	); err != nil {
		return err
	}

	return nil
}

func (l *lobby) Leave(token string) error {
	l.event()

	players, err := l.pool.Leave(token)
	if err != nil {
		return err
	}

	l.fanout.Deregister(token)

	payload := event.NewPayload()
	payload.SetPlayers(players)
	if err := l.broadcast.Send(
		event.New(event.JoinEvent, token, "", "", payload),
	); err != nil {
		return err
	}

	return nil
}

func (l *lobby) Insert(row, col, val int, token, trace string) (event.Event, error) {
	l.event()

	payload := event.NewPayload()
	payload.SetRow(row)
	payload.SetRow(col)
	payload.SetRow(val)

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

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	event := event.New(event.InsertEvent, token, trace, errMsg, payload)

	if err != nil {
		return event, nil
	}

	return nil, l.broadcast.Send(event)
}

func (l *lobby) Ping(row, col int, token, trace string) (event.Event, error) {
	l.event()

	payload := event.NewPayload()
	payload.SetRow(row)
	payload.SetColumn(col)

	if !sudoku.ValidBounds(row, col) {
		return event.New(
			event.PingEvent, token, trace,
			game.ErrOutOfBounds.Error(), payload,
		), nil
	}

	event := event.New(event.PingEvent, token, trace, "", payload)
	return nil, l.broadcast.Send(event)
}

func (l *lobby) State(token, trace string) event.Event {
	l.event()

	payload := event.NewPayload()
	payload.SetCurrent(l.game.Current())
	payload.SetInitial(l.game.Initial())
	return event.New(event.StateEvent, token, trace, "", payload)
}
