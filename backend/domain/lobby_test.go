package domain

import (
	"regexp"
	"testing"
)

func TestNewLobby(t *testing.T) {
	lobby, err := NewLobby()
	
	if err != nil {
		t.Fatalf("Failed to create new lobby: %v", err)
	}
	
	if len(lobby.ID) != 32 {
		t.Errorf("Expected ID length of 32, got %d", len(lobby.ID))
	}
	
	urlSafePattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !urlSafePattern.MatchString(lobby.ID) {
		t.Errorf("ID contains non-URL-safe characters: %s", lobby.ID)
	}
	
	if lobby.Puzzle == nil {
		t.Fatal("Puzzle is nil")
	}
	
	if lobby.Solution == nil {
		t.Fatal("Solution is nil")
	}
	
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if lobby.Solution.Board[row][col] == EmptyCell {
				t.Errorf("Solution has empty cell at position [%d][%d]", row, col)
			}
		}
	}
	
	if !lobby.Solution.IsValid() {
		t.Error("Solution is not valid")
	}
	
	hasEmptyCell := false
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if lobby.Puzzle.Board[row][col] == EmptyCell {
				hasEmptyCell = true
				break
			}
		}
		if hasEmptyCell {
			break
		}
	}
	
	if !hasEmptyCell {
		t.Error("Puzzle has no empty cells, should be different from solution")
	}
	
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			puzzleValue := lobby.Puzzle.Board[row][col]
			if puzzleValue != EmptyCell {
				solutionValue := lobby.Solution.Board[row][col]
				if puzzleValue != solutionValue {
					t.Errorf("Puzzle value %d at [%d][%d] does not match solution value %d", 
						puzzleValue, row, col, solutionValue)
				}
			}
		}
	}
}

func TestGenerateRandomID(t *testing.T) {
	lengths := []int{8, 16, 32, 64}
	
	for _, length := range lengths {
		id, err := generateRandomID(length)
		
		if err != nil {
			t.Fatalf("Failed to generate random ID of length %d: %v", length, err)
		}
		
		if len(id) != length {
			t.Errorf("Expected ID length of %d, got %d", length, len(id))
		}
		
		urlSafePattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
		if !urlSafePattern.MatchString(id) {
			t.Errorf("ID contains non-URL-safe characters: %s", id)
		}
	}
	
	idMap := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateRandomID(32)
		if err != nil {
			t.Fatalf("Failed to generate random ID: %v", err)
		}
		
		if idMap[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		
		idMap[id] = true
	}
}