package event

import (
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
)

// Represents arbitrary data carried within an event.
type Payload interface {
	Current() sudoku.Sudoku
	SetCurrent(current sudoku.Sudoku)
	Initial() sudoku.Sudoku
	SetInitial(initial sudoku.Sudoku)
	Row() *int
	SetRow(int)
	Column() *int
	SetColumn(int)
	Value() *int
	SetValue(int)
	Conflict() string
	SetConflict(conflict string)
	Players() []player.Player
	SetPlayers(players []player.Player)
}

type payload struct {
	current  sudoku.Sudoku
	initial  sudoku.Sudoku
	row      *int
	column   *int
	value    *int
	conflict string
	players  []player.Player
}

var _ Payload = &payload{}

func (p *payload) Current() sudoku.Sudoku {
	return p.current
}

func (p *payload) SetCurrent(current sudoku.Sudoku) {
	p.current = current
}

func (p *payload) Initial() sudoku.Sudoku {
	return p.initial
}

func (p *payload) SetInitial(initial sudoku.Sudoku) {
	p.initial = initial
}

func (p *payload) Row() *int {
	return p.row
}

func (p *payload) SetRow(row int) {
	p.row = &row
}

func (p *payload) Column() *int {
	return p.column
}

func (p *payload) SetColumn(column int) {
	p.column = &column
}

func (p *payload) Value() *int {
	return p.value
}

func (p *payload) SetValue(value int) {
	p.value = &value
}

func (p *payload) Conflict() string {
	return p.conflict
}

func (p *payload) SetConflict(conflict string) {
	p.conflict = conflict
}

func (p *payload) Players() []player.Player {
	return p.players
}

func (p *payload) SetPlayers(players []player.Player) {
	p.players = players
}

func NewPayload() *payload {
	return &payload{}
}
