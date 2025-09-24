package sudoku

import (
	"errors"
	"fmt"
	"math/rand"
)

const (
	BoardSize = 9
	BoxSize   = 3
	EmptyCell = 0
	MinValue  = 1
	MaxValue  = 9
	MinClues  = 17 // https://arxiv.org/abs/1201.0749
)

type Sudoku [BoardSize][BoardSize]int

func New() *Sudoku {
	return &Sudoku{}
}

func (org *Sudoku) Is(comp *Sudoku) bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if org[row][col] != comp[row][col] {
				return false
			}
		}
	}
	return true
}

func validVal(val int) bool {
	if val < MinValue {
		return false
	}
	if val > MaxValue {
		return false
	}
	return true
}

func (s *Sudoku) validRow(row, val int) bool {
	for col := 0; col < BoardSize; col++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

func (s *Sudoku) validCol(col, val int) bool {
	for row := 0; row < BoardSize; row++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

func (s *Sudoku) validBox(row, col, val int) bool {
	boxRow := (row / BoxSize) * BoxSize
	boxCol := (col / BoxSize) * BoxSize

	for r := 0; r < BoxSize; r++ {
		for c := 0; c < BoxSize; c++ {
			if s[boxRow+r][boxCol+c] == val {
				return false
			}
		}
	}

	return true
}

func (s *Sudoku) Complete() bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s[row][col] == EmptyCell {
				return false
			}
		}
	}

	return true
}

func (s *Sudoku) fill() bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s[row][col] == EmptyCell {
				nums := rand.Perm(MaxValue)
				for _, n := range nums {
					val := n + 1
					if s.validRow(row, val) && s.validCol(col, val) && s.validBox(row, col, val) {
						s[row][col] = val
						if s.fill() {
							return true
						}
						s[row][col] = EmptyCell // backtrack
					}
				}
				return false
			}
		}
	}
	return true
}

func (s *Sudoku) Fill(seed int64) {
	rand.New(rand.NewSource(seed))
	s.fill()
}

func (s *Sudoku) Insert(row, col, val int) (bool, error) {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return false, errors.New("position out of bounds")
	}

	if s[row][col] == val {
		return false, nil
	}

	if val != EmptyCell && s[row][col] != val {
		if !validVal(val) {
			return false, fmt.Errorf("value must be between %d and %d", MinValue, MaxValue)
		}

		if !s.validRow(row, val) {
			return false, errors.New("value already exists in this row")
		}

		if !s.validCol(col, val) {
			return false, errors.New("value already exists in this column")
		}

		if !s.validBox(row, col, val) {
			return false, errors.New("value already exists in this box")
		}
	}

	s[row][col] = val
	return true, nil
}

func (s *Sudoku) Copy(c *Sudoku) {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			c[row][col] = s[row][col]
		}
	}
}

func (s *Sudoku) solve(emptyCells [][2]int, index int, count *int, solution *Sudoku) {
	if *count > 1 {
		return
	}

	if index >= len(emptyCells) {
		(*count)++
		if *count == 1 {
			s.Copy(solution)
		}
		return
	}

	row, col := emptyCells[index][0], emptyCells[index][1]
	for val := 1; val <= 9; val++ {
		if s.validRow(row, val) && s.validCol(col, val) && s.validBox(row, col, val) {
			s[row][col] = val
			s.solve(emptyCells, index+1, count, solution)
			s[row][col] = EmptyCell
		}
	}
}

func (s *Sudoku) Clues() int {
	clues := 0
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s[row][col] != EmptyCell {
				clues++
			}
		}
	}
	return clues
}

func (s *Sudoku) UniqueSolution() bool {
	if s.Clues() < MinClues {
		return false
	}

	var emptyCells [][2]int
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s[row][col] == EmptyCell {
				emptyCells = append(emptyCells, [2]int{row, col})
			}
		}
	}

	count := 0
	solution := &Sudoku{}
	s.solve(emptyCells, 0, &count, solution)

	if count == 1 {
		solution.Copy(s)
		return true
	}
	return false
}

func (s *Sudoku) GeneratePuzzle(seed int64) {
	rand.New(rand.NewSource(seed))

	cells := make([][2]int, 0, BoardSize*BoardSize)
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			cells = append(cells, [2]int{row, col})
		}
	}
	rand.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })

	for _, cell := range cells {
		row, col := cell[0], cell[1]
		backup := s[row][col]
		s[row][col] = EmptyCell
		c := New()
		s.Copy(c)
		if !c.UniqueSolution() {
			s[row][col] = backup
		}
	}
}
