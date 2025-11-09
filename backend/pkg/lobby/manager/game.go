package manager

import (
	"encoding/json"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
)

type gamePayload struct {
	Current  sudoku.Sudoku `json:"current,omitempty"`
	Initial  sudoku.Sudoku `json:"initial,omitempty"`
	Row      *int          `json:"row,omitempty"`
	Column   *int          `json:"column,omitempty"`
	Value    *int          `json:"value,omitempty"`
	Conflict string        `json:"conflict,omitempty"`
}

// Serializes the game payload into bytes.
func (p *gamePayload) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing game payload: %v", err)
	}
	return b, nil
}

// Represents a manager for handling game operations including state
// synchronization, insertions, and player interactions.
type Game interface {
	// Returns the underlying game instance.
	Game() game.Game
	// Returns whether strict mode is enabled.
	Strict() bool
	// Sends the current and initial board state to the player.
	State(p player.Player) error
	// Processes a cell insertion request and broadcasts the result or sends
	// an error event to the player if validation fails.
	Insert(row, col, val int, p player.Player, trace string) error
	// Broadcasts a ping event for the specified cell position or sends an
	// error event to the player if coordinates are out of bounds.
	Ping(row, col int, p player.Player, trace string) error
}

type gameManager struct {
	game   game.Game
	bus    event.EventBus
	strict bool
}

var _ Game = &gameManager{}

// Returns a new game manager initialized with the provided game instance,
// event bus, and strict mode flag.
func NewGameManager(game game.Game, bus event.EventBus, strict bool) *gameManager {
	return &gameManager{game: game, bus: bus, strict: strict}
}

func (m *gameManager) Game() game.Game {
	return m.game
}

func (m *gameManager) Strict() bool {
	return m.strict
}

func (m *gameManager) State(p player.Player) error {
	payload := &gamePayload{Current: m.game.Current(), Initial: m.game.Initial()}
	e := event.New(event.StateEvent, p.Token(), "", "", payload)
	return p.Send(e)
}

// Processes an insertion in lax mode, allowing incorrect but not conflicting
// inserts. Broadcasts the result or sends an error event to the player.
func (m *gameManager) handleLax(row, col, val int, p player.Player, trace string) error {
	var err error
	payload := &gamePayload{Row: &row, Column: &col, Value: &val}

	payload.Current, err = m.game.Lax(row, col, val)
	if err == game.ErrRowConflict ||
		err == game.ErrColConflict ||
		err == game.ErrBoxConflict {
		payload.Conflict = err.Error()
	} else if err != nil {
		e := event.New(event.InsertEvent, p.Token(), trace, err.Error(), nil)
		return p.Send(e)
	}
	e := event.New(event.InsertEvent, p.Token(), trace, "", payload)
	return m.bus.Send(e)
}

// Processes an insertion in strict mode, placing the correct solution value
// and marking conflicts. Broadcasts the result or sends an error event to the player.
func (m *gameManager) handleStrict(row, col, val int, p player.Player, trace string) error {
	var err error
	payload := &gamePayload{Row: &row, Column: &col, Value: &val}

	payload.Current, err = m.game.Strict(row, col, val)
	if err == game.ErrIncorrect {
		payload.Conflict = game.ErrIncorrect.Error()
	} else if err != nil {
		e := event.New(event.InsertEvent, p.Token(), trace, err.Error(), nil)
		return p.Send(e)
	}
	e := event.New(event.InsertEvent, p.Token(), trace, "", payload)
	return m.bus.Send(e)
}

func (m *gameManager) Insert(row, col, val int, p player.Player, trace string) error {
	if m.Strict() {
		return m.handleStrict(row, col, val, p, trace)
	}
	return m.handleLax(row, col, val, p, trace)
}

func (m *gameManager) Ping(row, col int, p player.Player, trace string) error {
	if !sudoku.ValidBounds(row, col) {
		return p.Send(event.New(
			event.PingEvent,
			p.Token(),
			trace,
			game.ErrOutOfBounds.Error(),
			nil,
		))
	}

	return m.bus.Send(event.New(
		event.PingEvent,
		p.Token(),
		trace,
		"",
		&gamePayload{
			Row: &row, Column: &col,
		},
	))
}
