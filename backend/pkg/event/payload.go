package event

import (
	"sudojo/pkg/sudoku"
)

type PlayerStatus struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// Represents arbitrary data carried within an event.
type Payload interface {
	Current() sudoku.Sudoku
	Initial() sudoku.Sudoku
	Row() *int
	Column() *int
	Value() *int
	Conflict() string
	Players() []*PlayerStatus
}

type payload struct {
	current  sudoku.Sudoku
	initial  sudoku.Sudoku
	row      *int
	column   *int
	value    *int
	conflict string
	players  []*PlayerStatus
}

var _ Payload = &payload{}

func (p *payload) Current() sudoku.Sudoku {
	return p.current
}

func (p *payload) Initial() sudoku.Sudoku {
	return p.initial
}

func (p *payload) Row() *int {
	return p.row
}

func (p *payload) Column() *int {
	return p.column
}

func (p *payload) Value() *int {
	return p.value
}

func (p *payload) Conflict() string {
	return p.conflict
}

func (p *payload) Players() []*PlayerStatus {
	return p.players
}

func NewPlayerPayload(players []*PlayerStatus) *payload {
	return &payload{players: players}
}

func NewInsertPayload(row, col, val int, current sudoku.Sudoku, conflict string) *payload {
	return &payload{row: &row, column: &col, value: &val, current: current, conflict: conflict}
}

func NewPingPayload(row, col int) *payload {
	return &payload{row: &row, column: &col}
}

func NewStatePayload(current, initial sudoku.Sudoku) *payload {
	return &payload{current: current, initial: initial}
}
