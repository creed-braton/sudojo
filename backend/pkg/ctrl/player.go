package ctrl

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/sudoku"
	"sync"
	"time"
)

// Represents a player in the lobby. Provides methods to interact
// with the lobby as a player.
type Player interface {
	// Returns an event directed to the player or event.ErrClosedBuffer
	// if the buffer is closed.
	Receive() (event.Event, error)
	// Sets player inactive in lobby state and closes his event buffer.
	// Broadcasts an leave event to all players in the lobby.
	Leave()
	// Broadcasts a ping event to all players in the lobby if valid
	// row and column bounds are provided, otherwise sends an error
	// event to the player. Returns event.ErrClosedBuffer if event is
	// broadcast and the lobby event buffer is closed or if the event
	// is only directed to the player and his event buffer is closed.
	Ping(row, col int, trace string) error
	// Broadcasts an insert event to all players in the lobby if input
	// is valid based on the game state and mode. Sends an error event
	// directed to the player if input is invalid. Returns
	// event.ErrClosedBuffer if event is broadcast and the lobby event
	// buffer is closed or if the event is only directed to the player
	// and his event buffer is closed.
	Insert(row, col, val int, trace string) error
	// Sends lobby state to the player. Returns event.ErrClosedBuffer
	// if his event buffer is closed.
	State(trace string) error
}

type player struct {
	token       string
	lobby       lobby.Lobby
	buffer      event.Buffer
	broadcast   func(e event.Event) error
	leave       func(token string)
	updateEvent func()
	once        sync.Once
}

var _ Player = &player{}

// Returns an initialized player instance to interact with the lobby
// as a player.
func NewPlayer(
	token string,
	lobby lobby.Lobby,
	buffer event.Buffer,
	broadcast func(e event.Event) error,
	leave func(token string),
	updateEvent func(),
) *player {
	return &player{
		token:       token,
		lobby:       lobby,
		buffer:      buffer,
		broadcast:   broadcast,
		leave:       leave,
		updateEvent: updateEvent,
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
	p.updateEvent()

	return p.broadcast(event)
}

func (p *player) Insert(row, col, val int, trace string) error {
	p.updateEvent()
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
		SetPlayers(p.lobby.Players()).
		SetMaxPlayer(p.lobby.Size())
	event := event.New().SetType(event.StateEvent).SetSender(p.token).
		SetTrace(trace).SetTimestamp(now).SetPayload(payload)

	return p.buffer.Send(event)
}
