package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

type Lobby struct {
	ID       string
	Puzzle   *Sudoku
	Solution *Sudoku
}

func NewLobby() (*Lobby, error) {
	id, err := generateRandomID(32)
	if err != nil {
		return nil, err
	}
	
	puzzle := GeneratePuzzle(5)
	
	solution := puzzle.Copy()
	
	// Solve the puzzle completely to get the solution
	if !SolvePuzzle(solution) {
		// This should never happen with a valid puzzle, but just in case
		return nil, errors.New("failed to solve the puzzle")
	}
	
	return &Lobby{
		ID:       id,
		Puzzle:   puzzle,
		Solution: solution,
	}, nil
}

func generateRandomID(length int) (string, error) {
	byteLength := length * 3 / 4
	if length%4 != 0 {
		byteLength++
	}
	
	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	
	id := base64.URLEncoding.EncodeToString(randomBytes)
	
	if len(id) > length {
		id = id[:length]
	}
	
	return id, nil
}
