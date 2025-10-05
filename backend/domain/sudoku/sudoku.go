package sudoku

import (
	"errors"
	"fmt"
	"math/rand"
)

const (
	boardSize = 9
	boxSize   = 3
	emptyCell = 0
	minValue  = 1
	maxValue  = 9
	minClues  = 17 // https://arxiv.org/abs/1201.0749
)

type Sudoku [boardSize][boardSize]int

func New() *Sudoku {
	return &Sudoku{}
}

// Compares two Sudoku boards for equality by checking if all cells contain
// identical values at corresponding positions. It returns true if both boards
// are exactly the same, false otherwise. It doesn't check validity of either board.
func (org *Sudoku) equal(comp *Sudoku) bool {
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if org[row][col] != comp[row][col] {
				return false
			}
		}
	}
	return true
}

// Checks if the given row and column coordinates are within the valid bounds
// of the Sudoku board. Returns true if both coordinates are within the valid
// range, false otherwise.
func validBounds(row, col int) bool {
	if row < 0 || row >= boardSize || col < 0 || col >= boardSize {
		return false
	}
	return true
}

// Checks if a given integer value is within the valid range for Sudoku cell
// values. It returns true if the value is within the range, false otherwise.
func validVal(val int) bool {
	if val < minValue {
		return false
	}
	if val > maxValue {
		return false
	}
	return true
}

// Checks if a given value can be placed in the specified row without
// violating Sudoku row constraints. Returns true if the value doesn't
// already exist in that row, false otherwise.
func (s *Sudoku) validRow(row, val int) bool {
	for col := 0; col < boardSize; col++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

// Checks if a given value can be placed in the specified column without
// violating Sudoku column constraints. Returns true if the value doesn't
// already exist in that column, false otherwise.
func (s *Sudoku) validCol(col, val int) bool {
	for row := 0; row < boardSize; row++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

// Checks if a given value can be placed in the specified box without
// violating Sudoku box constraints. Returns true if the value doesn't
// already exist in that box, false otherwise.
func (s *Sudoku) validBox(row, col, val int) bool {
	boxRow := (row / boxSize) * boxSize
	boxCol := (col / boxSize) * boxSize

	for r := 0; r < boxSize; r++ {
		for c := 0; c < boxSize; c++ {
			if s[boxRow+r][boxCol+c] == val {
				return false
			}
		}
	}

	return true
}

// Validates whether a given value can be legally placed at the specified position
// on the Sudoku board. It checks bounds for position coordinates, value range and
// constraint verification for row, column, and box uniqueness. Returns nil  if the
// placement is valid, or an error describing the specific validation failure.
// Empty cell values are always considered valid placements.
func (s *Sudoku) Validate(row, col, val int) error {
	if !validBounds(row, col) {
		return errors.New("position out of bounds")
	}

	if val == emptyCell {
		return nil
	}

	if !validVal(val) {
		return fmt.Errorf("value must be between %d and %d", minValue, maxValue)
	}

	if !s.validRow(row, val) {
		return errors.New("value already exists in this row")
	}

	if !s.validCol(col, val) {
		return errors.New("value already exists in this column")
	}

	if !s.validBox(row, col, val) {
		return errors.New("value already exists in this box")
	}

	return nil
}

// Checks whether the Sudoku board is completely filled by verifying that all cells
// contain non-empty values. Returns true if no empty cells are found, false otherwise.
// It doesn't check validity of the board.
func (s *Sudoku) Complete() bool {
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if s[row][col] == emptyCell {
				return false
			}
		}
	}

	return true
}

func (s *Sudoku) fill() bool {
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if s[row][col] == emptyCell {
				nums := rand.Perm(maxValue)
				for _, n := range nums {
					val := n + 1
					if s.validRow(row, val) && s.validCol(col, val) && s.validBox(row, col, val) {
						s[row][col] = val
						if s.fill() {
							return true
						}
						s[row][col] = emptyCell // backtrack
					}
				}
				return false
			}
		}
	}
	return true
}

// Fills the Sudoku board with a complete and valid solution using the
// provided seed for random number generation.
func (s *Sudoku) Fill(seed int64) {
	rand.New(rand.NewSource(seed))
	s.fill()
}

// Copies all cell values from the current Sudoku board s to the
// provided destination board c.
func (s *Sudoku) Copy(c *Sudoku) {
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
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
			s[row][col] = emptyCell
		}
	}
}

// Returns the number of non-empty values on the Sudoku board.
func (s *Sudoku) clues() int {
	clues := 0
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if s[row][col] != emptyCell {
				clues++
			}
		}
	}
	return clues
}

// Returns true if the Sudoku board has exactly one valid solution and
// fills the board with that solution. Otherwise it returns false and
// leaves the board as is.
func (s *Sudoku) UniqueSolution() bool {
	if s.clues() < minClues {
		return false
	}

	var emptyCells [][2]int
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if s[row][col] == emptyCell {
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

// Transforms the current complete Sudoku board into a puzzle by removing
// cells while ensuring the puzzle maintains exactly one unique solution.
// Uses the provided seed for random number generation.
func (s *Sudoku) GeneratePuzzle(seed int64) {
	rand.New(rand.NewSource(seed))

	cells := make([][2]int, 0, boardSize*boardSize)
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			cells = append(cells, [2]int{row, col})
		}
	}
	rand.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })

	for _, cell := range cells {
		row, col := cell[0], cell[1]
		backup := s[row][col]
		s[row][col] = emptyCell
		c := New()
		s.Copy(c)
		if !c.UniqueSolution() {
			s[row][col] = backup
		}
	}
}
