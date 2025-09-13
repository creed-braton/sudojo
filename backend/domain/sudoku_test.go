package domain

import (
	"fmt"
	"testing"
)

func TestNewSudoku(t *testing.T) {
	s := NewSudoku()
	
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s.Board[row][col] != EmptyCell {
				t.Errorf("Expected empty cell at (%d,%d), got %d", row, col, s.Board[row][col])
			}
		}
	}
	
	initialBoard := [BoardSize][BoardSize]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
	
	s = NewSudoku(initialBoard)
	
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if s.Board[row][col] != initialBoard[row][col] {
				t.Errorf("Expected %d at (%d,%d), got %d", 
					initialBoard[row][col], row, col, s.Board[row][col])
			}
		}
	}
}

func TestValidateMove(t *testing.T) {
	initialBoard := [BoardSize][BoardSize]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
	
	s := NewSudoku(initialBoard)
	
	tests := []struct {
		name      string
		row       int
		col       int
		value     int
		expectErr bool
	}{
		{"Valid move", 0, 2, 4, false},
		{"Out of bounds row", -1, 0, 1, true},
		{"Out of bounds row", BoardSize, 0, 1, true},
		{"Out of bounds column", 0, -1, 1, true},
		{"Out of bounds column", 0, BoardSize, 1, true},
		{"Value too small", 0, 2, 0, true},
		{"Value too large", 0, 2, 10, true},
		{"Cell already filled", 0, 0, 1, true},
		{"Value exists in row", 0, 2, 5, true},
		{"Value exists in column", 2, 0, 5, true},
		{"Value exists in box", 1, 1, 9, true},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := s.ValidateMove(test.row, test.col, test.value)
			
			if test.expectErr && err == nil {
				t.Errorf("Expected error but got none for (%d,%d) with value %d", 
					test.row, test.col, test.value)
			}
			
			if !test.expectErr && err != nil {
				t.Errorf("Unexpected error: %v for (%d,%d) with value %d", 
					err, test.row, test.col, test.value)
			}
		})
	}
}

func TestMakeMove(t *testing.T) {
	s := NewSudoku()
	
	err := s.MakeMove(0, 0, 5)
	if err != nil {
		t.Errorf("Unexpected error making valid move: %v", err)
	}
	
	if s.Board[0][0] != 5 {
		t.Errorf("Expected cell (0,0) to be 5, got %d", s.Board[0][0])
	}
	
	err = s.MakeMove(0, 0, 6)
	if err == nil {
		t.Error("Expected error making move on filled cell, got none")
	}
	
	err = s.MakeMove(0, 1, 5)
	if err == nil {
		t.Error("Expected error making move with duplicate value in row, got none")
	}
}

func TestClearCell(t *testing.T) {
	s := NewSudoku()
	
	s.Board[0][0] = 5
	
	err := s.ClearCell(0, 0)
	if err != nil {
		t.Errorf("Unexpected error clearing cell: %v", err)
	}
	
	if s.Board[0][0] != EmptyCell {
		t.Errorf("Expected cell (0,0) to be empty, got %d", s.Board[0][0])
	}
	
	err = s.ClearCell(-1, 0)
	if err == nil {
		t.Error("Expected error clearing out of bounds cell, got none")
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		board    [BoardSize][BoardSize]int
		expected bool
	}{
		{
			"Empty board",
			[BoardSize][BoardSize]int{},
			true,
		},
		{
			"Valid incomplete board",
			[BoardSize][BoardSize]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			true,
		},
		{
			"Invalid row",
			[BoardSize][BoardSize]int{
				{5, 5, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			false,
		},
		{
			"Invalid column",
			[BoardSize][BoardSize]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{5, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			false,
		},
		{
			"Invalid box",
			[BoardSize][BoardSize]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 5, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			false,
		},
		{
			"Valid complete board",
			[BoardSize][BoardSize]int{
				{5, 3, 4, 6, 7, 8, 9, 1, 2},
				{6, 7, 2, 1, 9, 5, 3, 4, 8},
				{1, 9, 8, 3, 4, 2, 5, 6, 7},
				{8, 5, 9, 7, 6, 1, 4, 2, 3},
				{4, 2, 6, 8, 5, 3, 7, 9, 1},
				{7, 1, 3, 9, 2, 4, 8, 5, 6},
				{9, 6, 1, 5, 3, 7, 2, 8, 4},
				{2, 8, 7, 4, 1, 9, 6, 3, 5},
				{3, 4, 5, 2, 8, 6, 1, 7, 9},
			},
			true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSudoku(test.board)
			
			if s.IsValid() != test.expected {
				t.Errorf("Expected IsValid to return %v, got %v", test.expected, s.IsValid())
			}
		})
	}
}

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		board    [BoardSize][BoardSize]int
		expected bool
	}{
		{
			"Empty board",
			[BoardSize][BoardSize]int{},
			false,
		},
		{
			"Incomplete board",
			[BoardSize][BoardSize]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
			false,
		},
		{
			"Complete but invalid board",
			[BoardSize][BoardSize]int{
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
				{5, 5, 5, 5, 5, 5, 5, 5, 5},
			},
			false,
		},
		{
			"Complete and valid board",
			[BoardSize][BoardSize]int{
				{5, 3, 4, 6, 7, 8, 9, 1, 2},
				{6, 7, 2, 1, 9, 5, 3, 4, 8},
				{1, 9, 8, 3, 4, 2, 5, 6, 7},
				{8, 5, 9, 7, 6, 1, 4, 2, 3},
				{4, 2, 6, 8, 5, 3, 7, 9, 1},
				{7, 1, 3, 9, 2, 4, 8, 5, 6},
				{9, 6, 1, 5, 3, 7, 2, 8, 4},
				{2, 8, 7, 4, 1, 9, 6, 3, 5},
				{3, 4, 5, 2, 8, 6, 1, 7, 9},
			},
			true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSudoku(test.board)
			
			if s.IsComplete() != test.expected {
				t.Errorf("Expected IsComplete to return %v, got %v", test.expected, s.IsComplete())
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	initialBoard := [BoardSize][BoardSize]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
	
	s := NewSudoku(initialBoard)
	
	tests := []struct {
		row        int
		col        int
		expected   int
		expectErr  bool
	}{
		{0, 0, 5, false},
		{0, 2, 0, false},
		{-1, 0, 0, true},
		{0, -1, 0, true},
		{BoardSize, 0, 0, true},
		{0, BoardSize, 0, true},
	}
	
	for _, test := range tests {
		val, err := s.GetValue(test.row, test.col)
		
		if test.expectErr && err == nil {
			t.Errorf("Expected error getting value at (%d,%d), got none", test.row, test.col)
		}
		
		if !test.expectErr && err != nil {
			t.Errorf("Unexpected error getting value at (%d,%d): %v", test.row, test.col, err)
		}
		
		if !test.expectErr && val != test.expected {
			t.Errorf("Expected value %d at (%d,%d), got %d", 
				test.expected, test.row, test.col, val)
		}
	}
}

func TestCopy(t *testing.T) {
	initialBoard := [BoardSize][BoardSize]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
	
	original := NewSudoku(initialBoard)
	copy := original.Copy()
	
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if original.Board[row][col] != copy.Board[row][col] {
				t.Errorf("Copy differs at (%d,%d): original=%d, copy=%d", 
					row, col, original.Board[row][col], copy.Board[row][col])
			}
		}
	}
	
	copy.Board[0][0] = 9
	if original.Board[0][0] == copy.Board[0][0] {
		t.Errorf("Copy should be independent of original, but both have value %d at (0,0)",
			copy.Board[0][0])
	}
}

func TestGeneratePuzzle(t *testing.T) {
	// Test with different difficulty levels
	difficulties := []int{1, 5, 10}
	
	for _, difficulty := range difficulties {
		t.Run(fmt.Sprintf("Difficulty %d", difficulty), func(t *testing.T) {
			puzzle := GeneratePuzzle(difficulty)
			
			// Verify the puzzle is valid
			if !puzzle.IsValid() {
				t.Errorf("Generated puzzle is not valid")
			}
			
			// Count empty cells
			emptyCells := 0
			for row := 0; row < BoardSize; row++ {
				for col := 0; col < BoardSize; col++ {
					if puzzle.Board[row][col] == EmptyCell {
						emptyCells++
					}
				}
			}
			
			// Verify puzzle has the expected number of empty cells based on difficulty
			// Higher difficulty should have more empty cells
			switch difficulty {
			case 1:
				if emptyCells < 20 || emptyCells > 30 {
					t.Errorf("Expected 20-30 empty cells for difficulty 1, got %d", emptyCells)
				}
			case 5:
				if emptyCells < 30 || emptyCells > 45 {
					t.Errorf("Expected 30-45 empty cells for difficulty 5, got %d", emptyCells)
				}
			case 10:
				if emptyCells < 45 || emptyCells > 60 {
					t.Errorf("Expected 45-60 empty cells for difficulty 10, got %d", emptyCells)
				}
			}
			
			// Verify the puzzle has a unique solution
			solutions := 0
			emptyCellPositions := findEmptyCells(puzzle)
			puzzleCopy := puzzle.Copy() // Make a copy to not modify the original
			solve(puzzleCopy, emptyCellPositions, 0, &solutions)
			
			if solutions != 1 {
				t.Errorf("Expected exactly 1 solution, got %d", solutions)
			}
		})
	}
}