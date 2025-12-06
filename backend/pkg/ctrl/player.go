package ctrl

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/sudoku"
	"sync"
	"time"
)

type Player interface {
	Receive() (event.Event, error)
	Leave()
	Ping(row, col int, trace string) error
	Insert(row, col, val int, trace string) error
	State(trace string) error
}

type player struct {
	token     string
	lobby     lobby.Lobby
	buffer    event.Buffer
	broadcast func(e event.Event) error
	leave     func(token string)
	once      sync.Once
}

var _ Player = &player{}

func NewPlayer(
	token string,
	lobby lobby.Lobby,
	buffer event.Buffer,
	broadcast func(e event.Event) error,
	leave func(token string),
) *player {
	return &player{
		token:     token,
		lobby:     lobby,
		buffer:    buffer,
		broadcast: broadcast,
		leave:     leave,
	}
}

func (p *player) Receive() (event.Event, error) {
	return p.buffer.Receive()
}

func (p *player) Leave() {
	p.once.Do(func() {
		p.leave(p.token)
		p.buffer.Close()
	})
}

func (p *player) Ping(row, col int, trace string) error {
	event := event.New().SetType(event.PingEvent).SetSender(p.token).
		SetTrace(trace).SetTimestamp(time.Now().UTC().UnixNano()).
		SetPayload(event.NewPayload().SetRow(row).SetColumn(col))

	if !sudoku.ValidBounds(row, col) {
		event.SetError(game.ErrOutOfBounds.Error())
		return p.buffer.Send(event)
	}

	return p.broadcast(event)
}

func (p *player) Insert(row, col, val int, trace string) error {
	event := event.New().SetType(event.InsertEvent).SetSender(p.token).
		SetTrace(trace).SetTimestamp(time.Now().UTC().UnixNano()).
		SetPayload(event.NewPayload().SetRow(row).SetColumn(col).SetValue(val))

	var current sudoku.Sudoku
	var err error
	if p.lobby.Strict() {
		current, err = p.lobby.Game().Strict(row, col, val)
		if err == game.ErrIncorrect {
			event.Payload().SetConflict(err.Error())
			err = nil
		}
	} else {
		current, err = p.lobby.Game().Lax(row, col, val)
		if err == game.ErrRowConflict ||
			err == game.ErrColConflict ||
			err == game.ErrBoxConflict {
			event.Payload().SetConflict(err.Error())
			err = nil
		}
	}
	event.Payload().SetCurrent(current)

	if err != nil {
		return p.buffer.Send(event)
	}

	return p.broadcast(event)
}

func (p *player) State(trace string) error {
	now := time.Now().UTC().UnixNano()
	payload := event.NewPayload().
		SetCurrent(p.lobby.Game().Current()).
		SetInitial(p.lobby.Game().Initial()).
		SetStrict(p.lobby.Strict()).
		SetPlayers(p.lobby.Players())
	event := event.New().SetType(event.StateEvent).SetSender(p.token).
		SetTrace(trace).SetTimestamp(now).SetPayload(payload)

	return p.buffer.Send(event)
}
