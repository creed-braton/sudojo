package event

import (
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
)

const (
	SystemEvent = "system"
	LeaveEvent  = "leave"
	JoinEvent   = "join"
	StateEvent  = "state"
	InsertEvent = "insert"
	PingEvent   = "ping"
)

// Represents an occurrence within a lobby, used to notify clients about
// changes to the lobby state. Each event has a type identifying the kind of
// action (e.g., join, leave, insert), a timestamp indicating when it was
// created, and optionally a trace ID for end-to-end tracking. The remaining
// fields serve as a flexible payload whose relevant values vary depending on
// the event type.
type Event interface {
	// Type of the event.
	Type() string
	// Creation timestamp of the event.
	Timestamp() int64
	// Trace ID of the event to identify and keep track of it.
	Trace() string
	// Error message if something went wrong while processing this event.
	Error() string
	// Sets the error message.
	SetError(msg string) Event
	// Configuration of the lobby.
	Config() lobby.Config
	// Sets the lobby configuration.
	SetConfig(config lobby.Config) Event
	// The current Sudoku board state.
	Current() sudoku.Sudoku
	// Updates the current Sudoku board state.
	SetCurrent(current sudoku.Sudoku) Event
	// The initial Sudoku board state.
	Initial() sudoku.Sudoku
	// Updates the initial Sudoku board state.
	SetInitial(initial sudoku.Sudoku) Event
	// Row position of the event, or nil if not specified.
	Row() *int
	// Sets the row position of the event.
	SetRow(row int) Event
	// Column position of the event, or nil if not specified.
	Column() *int
	// Sets the column position of the event.
	SetColumn(column int) Event
	// Value being placed in the Sudoku board, or nil if not specified.
	Value() *int
	// Sets the value of the event.
	SetValue(value int) Event
	// Conflict description, if any exists.
	Conflict() string
	// Sets the conflict description.
	SetConflict(conflict string) Event
	// Players in the lobby, including their names and activity state.
	Players() []player.Player
	// Sets the players list.
	SetPlayers(players []player.Player) Event
}

type event struct {
	eventType        string
	timestamp        int64
	trace            string
	error, conflict  string
	config           lobby.Config
	row, col, val    *int
	current, initial sudoku.Sudoku
	players          []player.Player
}

var _ Event = &event{}

func New(t string, now int64, trace string) *event {
	return &event{eventType: t, timestamp: now, trace: trace}
}

func (e *event) Type() string {
	return e.eventType
}

func (e *event) Timestamp() int64 {
	return e.timestamp
}

func (e *event) Trace() string {
	return e.trace
}

func (e *event) Error() string {
	return e.error
}

func (e *event) SetError(msg string) Event {
	e.error = msg
	return e
}

func (e *event) Config() lobby.Config {
	return e.config
}

func (e *event) SetConfig(config lobby.Config) Event {
	e.config = config
	return e
}

func (e *event) Current() sudoku.Sudoku {
	return e.current
}

func (e *event) SetCurrent(current sudoku.Sudoku) Event {
	e.current = current
	return e
}

func (e *event) Initial() sudoku.Sudoku {
	return e.initial
}

func (e *event) SetInitial(initial sudoku.Sudoku) Event {
	e.initial = initial
	return e
}

func (e *event) Row() *int {
	return e.row
}

func (e *event) SetRow(row int) Event {
	e.row = &row
	return e
}

func (e *event) Column() *int {
	return e.col
}

func (e *event) SetColumn(col int) Event {
	e.col = &col
	return e
}

func (e *event) Value() *int {
	return e.val
}

func (e *event) SetValue(val int) Event {
	e.val = &val
	return e
}

func (e *event) Conflict() string {
	return e.conflict
}

func (e *event) SetConflict(conflict string) Event {
	e.conflict = conflict
	return e
}

func (e *event) Players() []player.Player {
	return e.players
}

func (e *event) SetPlayers(players []player.Player) Event {
	e.players = players
	return e
}
