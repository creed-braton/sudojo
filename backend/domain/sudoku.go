package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const (
	BoardSize = 9
	BoxSize   = 3
	EmptyCell = 0
	MinValue  = 1
	MaxValue  = 9
)

type Sudoku struct {
	Board [BoardSize][BoardSize]int
}

func NewSudoku(initialBoard ...[BoardSize][BoardSize]int) *Sudoku {
	s := &Sudoku{}

	if len(initialBoard) > 0 {
		s.Board = initialBoard[0]
	}

	return s
}

func (s *Sudoku) ValidateMove(row, col, value int) error {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return errors.New("position out of bounds")
	}

	if value < MinValue || value > MaxValue {
		return fmt.Errorf("value must be between %d and %d", MinValue, MaxValue)
	}

	if s.Board[row][col] != EmptyCell {
		return errors.New("cell is already filled")
	}

	if !s.isValidInRow(row, value) {
		return errors.New("value already exists in this row")
	}

	if !s.isValidInColumn(col, value) {
		return errors.New("value already exists in this column")
	}

	if !s.isValidInBox(row, col, value) {
		return errors.New("value already exists in this 3x3 box")
	}

	return nil
}

func (s *Sudoku) isValidInRow(row, value int) bool {
	for col := 0; col < BoardSize; col++ {
		if s.Board[row][col] == value {
			return false
		}
	}
	return true
}

func (s *Sudoku) isValidInColumn(col, value int) bool {
	for row := 0; row < BoardSize; row++ {
		if s.Board[row][col] == value {
			return false
		}
	}
	return true
}

func (s *Sudoku) isValidInBox(row, col, value int) bool {
	boxRow := (row / BoxSize) * BoxSize
	boxCol := (col / BoxSize) * BoxSize

	for r := 0; r < BoxSize; r++ {
		for c := 0; c < BoxSize; c++ {
			if s.Board[boxRow+r][boxCol+c] == value {
				return false
			}
		}
	}
	return true
}

func (s *Sudoku) MakeMove(row, col, value int) error {
	if err := s.ValidateMove(row, col, value); err != nil {
		return err
	}

	s.Board[row][col] = value
	return nil
}

func (s *Sudoku) ClearCell(row, col int) error {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return errors.New("position out of bounds")
	}

	s.Board[row][col] = EmptyCell
	return nil
}

// ClearCellWithInitialCheck clears a cell only if it wasn't part of the initial puzzle
func (s *Sudoku) ClearCellWithInitialCheck(row, col int, initialBoard *Sudoku) error {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return errors.New("position out of bounds")
	}

	// Check if this cell was part of the initial puzzle
	if initialBoard.Board[row][col] != EmptyCell {
		return errors.New("cannot clear initial puzzle cells")
	}

	s.Board[row][col] = EmptyCell
	return nil
}

func (s *Sudoku) IsComplete() bool {
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s.Board[row][col] == EmptyCell {
				return false
			}
		}
	}

	return s.IsValid()
}

func (s *Sudoku) IsValid() bool {
	for row := 0; row < BoardSize; row++ {
		if !s.isValidRow(row) {
			return false
		}
	}

	for col := 0; col < BoardSize; col++ {
		if !s.isValidColumn(col) {
			return false
		}
	}

	for boxRow := 0; boxRow < BoardSize; boxRow += BoxSize {
		for boxCol := 0; boxCol < BoardSize; boxCol += BoxSize {
			if !s.isValidBox(boxRow, boxCol) {
				return false
			}
		}
	}

	return true
}

func (s *Sudoku) isValidRow(row int) bool {
	seen := make(map[int]bool)
	for col := 0; col < BoardSize; col++ {
		val := s.Board[row][col]
		if val != EmptyCell {
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}
	return true
}

func (s *Sudoku) isValidColumn(col int) bool {
	seen := make(map[int]bool)
	for row := 0; row < BoardSize; row++ {
		val := s.Board[row][col]
		if val != EmptyCell {
			if seen[val] {
				return false
			}
			seen[val] = true
		}
	}
	return true
}

func (s *Sudoku) isValidBox(boxRow, boxCol int) bool {
	seen := make(map[int]bool)
	for row := 0; row < BoxSize; row++ {
		for col := 0; col < BoxSize; col++ {
			val := s.Board[boxRow+row][boxCol+col]
			if val != EmptyCell {
				if seen[val] {
					return false
				}
				seen[val] = true
			}
		}
	}
	return true
}

func (s *Sudoku) GetValue(row, col int) (int, error) {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return 0, errors.New("position out of bounds")
	}

	return s.Board[row][col], nil
}

func (s *Sudoku) Copy() *Sudoku {
	newSudoku := NewSudoku()
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			newSudoku.Board[row][col] = s.Board[row][col]
		}
	}
	return newSudoku
}

func GeneratePuzzle(difficulty int) *Sudoku {
	rand.Seed(time.Now().UnixNano())

	s := generateCompletedBoard()

	cells := removeCellsKeepingUniqueSolution(s, difficulty)

	return cells
}

func generateCompletedBoard() *Sudoku {
	s := NewSudoku()
	fillBoard(s, 0, 0)
	return s
}

func fillBoard(s *Sudoku, row, col int) bool {
	if row == BoardSize {
		return true
	}

	nextRow, nextCol := getNextCell(row, col)

	numbers := generateRandomOrder()

	for _, num := range numbers {
		if s.isValidInRow(row, num) && s.isValidInColumn(col, num) && s.isValidInBox(row, col, num) {
			s.Board[row][col] = num

			if fillBoard(s, nextRow, nextCol) {
				return true
			}

			s.Board[row][col] = EmptyCell
		}
	}

	return false
}

func getNextCell(row, col int) (int, int) {
	col++
	if col == BoardSize {
		col = 0
		row++
	}
	return row, col
}

func generateRandomOrder() []int {
	numbers := make([]int, 9)
	for i := range numbers {
		numbers[i] = i + 1
	}

	for i := len(numbers) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}

	return numbers
}

func removeCellsKeepingUniqueSolution(s *Sudoku, difficulty int) *Sudoku {
	puzzle := s.Copy()

	cellsToRemove := calculateCellsToRemove(difficulty)

	cells := make([][2]int, 0, BoardSize*BoardSize)
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			cells = append(cells, [2]int{row, col})
		}
	}

	shuffleCells(cells)

	removed := 0
	for _, cell := range cells {
		row, col := cell[0], cell[1]

		if removed >= cellsToRemove {
			break
		}

		backup := puzzle.Board[row][col]
		puzzle.Board[row][col] = EmptyCell

		if !hasUniqueSolution(puzzle) {
			puzzle.Board[row][col] = backup
		} else {
			removed++
		}
	}

	return puzzle
}

func calculateCellsToRemove(difficulty int) int {
	minCells := 20
	maxCells := 60

	if difficulty < 1 {
		difficulty = 1
	} else if difficulty > 10 {
		difficulty = 10
	}

	difficultyFactor := float64(difficulty) / 10.0
	cellsToRemove := minCells + int(float64(maxCells-minCells)*difficultyFactor)

	return cellsToRemove
}

func shuffleCells(cells [][2]int) {
	for i := len(cells) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}

func hasUniqueSolution(s *Sudoku) bool {
	solutions := 0
	emptyCells := findEmptyCells(s)

	if len(emptyCells) == 0 {
		return true
	}

	solve(s.Copy(), emptyCells, 0, &solutions)

	return solutions == 1
}

func findEmptyCells(s *Sudoku) [][2]int {
	var emptyCells [][2]int

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s.Board[row][col] == EmptyCell {
				emptyCells = append(emptyCells, [2]int{row, col})
			}
		}
	}

	return emptyCells
}

func solve(s *Sudoku, emptyCells [][2]int, index int, solutions *int) {
	if *solutions > 1 {
		return
	}

	if index >= len(emptyCells) {
		(*solutions)++
		return
	}

	row, col := emptyCells[index][0], emptyCells[index][1]

	for num := 1; num <= 9; num++ {
		if s.isValidInRow(row, num) && s.isValidInColumn(col, num) && s.isValidInBox(row, col, num) {
			s.Board[row][col] = num
			solve(s, emptyCells, index+1, solutions)
			s.Board[row][col] = EmptyCell
		}
	}
}

// SolvePuzzle solves a Sudoku puzzle and returns true if a solution was found
func SolvePuzzle(s *Sudoku) bool {
	emptyCells := findEmptyCells(s)
	if len(emptyCells) == 0 {
		return true
	}
	
	return solveSingle(s, emptyCells, 0)
}

// solveSingle attempts to solve the puzzle and keeps the solution in the board
func solveSingle(s *Sudoku, emptyCells [][2]int, index int) bool {
	if index >= len(emptyCells) {
		return true
	}

	row, col := emptyCells[index][0], emptyCells[index][1]

	for num := 1; num <= 9; num++ {
		if s.isValidInRow(row, num) && s.isValidInColumn(col, num) && s.isValidInBox(row, col, num) {
			s.Board[row][col] = num
			
			if solveSingle(s, emptyCells, index+1) {
				return true
			}
			
			s.Board[row][col] = EmptyCell
		}
	}
	
	return false
}
