package game

import "sudojo/pkg/sudoku"

type mockGame struct {
	start    func()
	lax      func(row, col, val int) (sudoku.Sudoku, error)
	strict   func(row, col, val int) (sudoku.Sudoku, error)
	current  func() sudoku.Sudoku
	initial  func() sudoku.Sudoku
	solution func() sudoku.Sudoku
	started  func() *int64
	finished func() *int64
}

var _ Game = &mockGame{}

func NewMockGame(
	start func(),
	lax func(row, col, val int) (sudoku.Sudoku, error),
	strict func(row, col, val int) (sudoku.Sudoku, error),
	current func() sudoku.Sudoku,
	initial func() sudoku.Sudoku,
	solution func() sudoku.Sudoku,
	started func() *int64,
	finished func() *int64,
) *mockGame {
	return &mockGame{
		start:    start,
		lax:      lax,
		strict:   strict,
		current:  current,
		initial:  initial,
		solution: solution,
		started:  started,
		finished: finished,
	}
}

func (g *mockGame) Start() {
	g.start()
}

func (g *mockGame) Lax(row, col, val int) (sudoku.Sudoku, error) {
	return g.lax(row, col, val)
}

func (g *mockGame) Strict(row, col, val int) (sudoku.Sudoku, error) {
	return g.strict(row, col, val)
}

func (g *mockGame) Current() sudoku.Sudoku {
	return g.current()
}

func (g *mockGame) Initial() sudoku.Sudoku {
	return g.initial()
}

func (g *mockGame) Solution() sudoku.Sudoku {
	return g.solution()
}

func (g *mockGame) Started() *int64 {
	return g.started()
}

func (g *mockGame) Finished() *int64 {
	return g.finished()
}
