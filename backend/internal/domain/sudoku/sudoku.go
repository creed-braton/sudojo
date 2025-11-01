package sudoku

import (
	"math/rand"
)

const (
	BoardSize = 9  // Number of rows and columns in a Sudoku board
	boxSize   = 3  // Dimension of each sub-box in the Sudoku grid
	EmptyCell = 0  // Represents an unfilled cell
	MinValue  = 1  // Smallest valid Sudoku cell value
	MaxValue  = 9  // Largest valid Sudoku cell value
	minClues  = 17 // Minimum number of clues required for an unique solution Sudoku puzzle: https://arxiv.org/abs/1201.0749
)

// Represents a 9x9 Sudoku board. Provides methods for validating cell
// placements, checking board completeness, generating randomly complete
// boards, deriving uniquely solvable puzzles, and more utility functions.
//
// Each cell stores an integer value between 0 and 9, where 0 indicates an
// empty cell and values 1–9 represent placed digits.
type Sudoku [BoardSize][BoardSize]int

// Returns a pointer to a new, empty Sudoku board.
func New() *Sudoku {
	return &Sudoku{}
}

// Compares two Sudoku boards for equality by checking if all cells contain
// identical values at corresponding positions. It returns true if both boards
// are exactly the same, false otherwise. It doesn't check validity of either board.
func (org *Sudoku) Equal(comp *Sudoku) bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if org[row][col] != comp[row][col] {
				return false
			}
		}
	}
	return true
}

// Exports the Sudoku structs values into a 9x9 integer matrix.
func (s *Sudoku) Int() [][]int {
	c := make([][]int, 9)
	for i := range s {
		c[i] = s[i][:]
	}
	return c
}

// Expects a 9x9 integer matrix and turns it into the Sudoku struct.
func Load(matrix [][]int) *Sudoku {
	s := New()
	for i := 0; i < len(matrix) && i < BoardSize; i++ {
		for j := 0; j < len(matrix[i]) && j < BoardSize; j++ {
			s[i][j] = matrix[i][j]
		}
	}
	return s
}

// Checks if the given row and column coordinates are within the valid bounds
// of the Sudoku board. Returns true if both coordinates are within the valid
// range, false otherwise.
func ValidBounds(row, col int) bool {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return false
	}
	return true
}

// Checks if a given integer value is within the valid range for Sudoku cell
// values. It returns true if the value is within the range, false otherwise.
func ValidVal(val int) bool {
	if val < MinValue {
		return false
	}
	if val > MaxValue {
		return false
	}
	return true
}

// Checks if a given value can be placed in the specified row without
// violating Sudoku row constraints. Returns true if the value doesn't
// already exist in that row, false otherwise.
func (s *Sudoku) ValidRow(row, val int) bool {
	for col := 0; col < BoardSize; col++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

// Checks if a given value can be placed in the specified column without
// violating Sudoku column constraints. Returns true if the value doesn't
// already exist in that column, false otherwise.
func (s *Sudoku) ValidCol(col, val int) bool {
	for row := 0; row < BoardSize; row++ {
		if s[row][col] == val {
			return false
		}
	}

	return true
}

// Checks if a given value can be placed in the specified box without
// violating Sudoku box constraints. Returns true if the value doesn't
// already exist in that box, false otherwise.
func (s *Sudoku) ValidBox(row, col, val int) bool {
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

// Checks whether the Sudoku board is completely filled by verifying that all cells
// contain non-empty values. Returns true if no empty cells are found, false otherwise.
// It doesn't check validity of the board.
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

// Recursively fills the board with random valid values until a full valid Sudoku
// solution is reached. Uses backtracking to explore possibilities.
func (s *Sudoku) fill() bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s[row][col] == EmptyCell {
				nums := rand.Perm(MaxValue)
				for _, n := range nums {
					val := n + 1
					if s.ValidRow(row, val) && s.ValidCol(col, val) && s.ValidBox(row, col, val) {
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

// Fills the Sudoku board with a complete and valid solution using the
// provided seed for random number generation.
func (s *Sudoku) Fill(seed int64) {
	rand.New(rand.NewSource(seed))
	s.fill()
}

// Copies all cell values from the current Sudoku board s to the
// provided destination board c.
func (s *Sudoku) Copy(c *Sudoku) {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			c[row][col] = s[row][col]
		}
	}
}

// Recursively fills empty cells to explore all valid Sudoku solutions.
// Search terminates early if more than one valid solution is found.
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
		if s.ValidRow(row, val) && s.ValidCol(col, val) && s.ValidBox(row, col, val) {
			s[row][col] = val
			s.solve(emptyCells, index+1, count, solution)
			s[row][col] = EmptyCell
		}
	}
}

// Returns the number of non-empty values on the Sudoku board.
func (s *Sudoku) clues() int {
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

// Returns true if the Sudoku board has exactly one valid solution and
// fills the board with that solution. Otherwise it returns false and
// leaves the board as is.
func (s *Sudoku) UniqueSolution() bool {
	if s.clues() < minClues {
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

// Transforms the current complete Sudoku board into a puzzle by removing
// cells while ensuring the puzzle maintains exactly one unique solution.
// Uses the provided seed for random number generation.
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
