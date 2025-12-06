package event

import (
	"sudojo/pkg/sudoku"
)

// Represents a player with name and status.
type Player interface {
	// In-game name of the player.
	Name() string
	// Activity status of the player.
	Active() bool
}

type player struct {
	name   string
	active bool
}

var _ Player = &player{}

// Creates a new player with the given token and name.
func NewPlayer(name string, active bool) *player {
	return &player{name: name, active: active}
}

func (p *player) Name() string {
	return p.name
}

func (p *player) Active() bool {
	return p.active
}

// Carries arbitrary data within an event, including Sudoku board states,
// cell updates, conflict information, and participating players.
type Payload interface {
	// The current Sudoku board state.
	Current() sudoku.Sudoku
	// Updates the current Sudoku board state.
	SetCurrent(current sudoku.Sudoku) Payload
	// The initial Sudoku board state.
	Initial() sudoku.Sudoku
	// Updates the initial Sudoku board state.
	SetInitial(initial sudoku.Sudoku) Payload
	// Row index of the event, or nil if not specified.
	Row() *int
	// Sets the row index of the event.
	SetRow(row int) Payload
	// Column index of the event, or nil if not specified.
	Column() *int
	// Sets the column index of the event.
	SetColumn(column int) Payload
	// Value being placed in the Sudoku board, or nil if not specified.
	Value() *int
	// Sets the value of the event.
	SetValue(value int) Payload
	// Conflict description, if any exists.
	Conflict() string
	// Sets the conflict description.
	SetConflict(conflict string) Payload
	// Players in the session, including their names and activity state.
	Players() []Player
	// Sets the players.
	SetPlayers(players []Player) Payload
	// Maximum amount of players in the session.
	MaxPlayer() int
	// Sets the maximum player amount.
	SetMaxPlayer(max int) Payload
	// Strict mode of the session.
	Strict() *bool
	// Sets the strict mode.
	SetStrict(strict bool) Payload
}

type payload struct {
	current   sudoku.Sudoku
	initial   sudoku.Sudoku
	row       *int
	column    *int
	value     *int
	conflict  string
	players   []Player
	maxPlayer int
	strict    *bool
}

var _ Payload = &payload{}

func NewPayload() *payload {
	return &payload{}
}

func (p *payload) Current() sudoku.Sudoku {
	return p.current
}

func (p *payload) SetCurrent(current sudoku.Sudoku) Payload {
	p.current = current
	return p
}

func (p *payload) Initial() sudoku.Sudoku {
	return p.initial
}

func (p *payload) SetInitial(initial sudoku.Sudoku) Payload {
	p.initial = initial
	return p
}

func (p *payload) Row() *int {
	return p.row
}

func (p *payload) SetRow(row int) Payload {
	p.row = &row
	return p
}

func (p *payload) Column() *int {
	return p.column
}

func (p *payload) SetColumn(column int) Payload {
	p.column = &column
	return p
}

func (p *payload) Value() *int {
	return p.value
}

func (p *payload) SetValue(value int) Payload {
	p.value = &value
	return p
}

func (p *payload) Conflict() string {
	return p.conflict
}

func (p *payload) SetConflict(conflict string) Payload {
	p.conflict = conflict
	return p
}

func (p *payload) Players() []Player {
	return p.players
}

func (p *payload) SetPlayers(players []Player) Payload {
	p.players = players
	return p
}

func (p *payload) MaxPlayer() int {
	return p.maxPlayer
}

func (p *payload) SetMaxPlayer(max int) Payload {
	p.maxPlayer = max
	return p
}

func (p *payload) Strict() *bool {
	return p.strict
}

func (p *payload) SetStrict(strict bool) Payload {
	p.strict = &strict
	return p
}
