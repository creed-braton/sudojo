package game

import "sudojo/pkg/sudoku"

func NewMock(start bool) *game {
	initial := sudoku.NewFromInts(
		[][]int{
			{0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 0, 0, 0, 2, 0, 0, 3},
			{0, 0, 0, 4, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 5, 0, 0},
			{4, 0, 1, 6, 0, 0, 0, 0, 0},
			{0, 0, 7, 1, 0, 0, 0, 0, 0},
			{0, 5, 0, 0, 0, 0, 2, 0, 0},
			{0, 0, 0, 0, 8, 0, 0, 4, 0},
			{0, 3, 0, 9, 1, 0, 0, 0, 0},
		},
	)
	current := sudoku.New()
	initial.Copy(current)
	solution := sudoku.NewFromInts(
		[][]int{
			{7, 4, 5, 3, 6, 8, 9, 1, 2},
			{8, 1, 9, 5, 7, 2, 4, 6, 3},
			{3, 6, 2, 4, 9, 1, 8, 5, 7},
			{6, 9, 3, 8, 2, 4, 5, 7, 1},
			{4, 2, 1, 6, 5, 7, 3, 9, 8},
			{5, 8, 7, 1, 3, 9, 6, 2, 4},
			{1, 5, 8, 7, 4, 6, 2, 3, 9},
			{9, 7, 6, 2, 8, 3, 1, 4, 5},
			{2, 3, 4, 9, 1, 5, 7, 8, 6},
		},
	)
	// insert one value so it differs from initial board
	current[8][8] = solution[8][8]

	g := &game{
		initial:    initial,
		current:    current,
		solution:   solution,
		difficulty: Joker,
	}

	if start {
		g.Start(int64(42))
	}

	return g
}
