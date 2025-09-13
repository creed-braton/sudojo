package domain

import (
	"crypto/rand"
	"encoding/base64"
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
	
	emptyCells := findEmptyCells(solution)
	solutions := 0
	solve(solution, emptyCells, 0, &solutions)
	
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
