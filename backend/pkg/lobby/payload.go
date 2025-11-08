package lobby

import (
	"encoding/json"
	"fmt"
	"sudojo/pkg/sudoku"
)

type status struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type playerState []*status

func (p *playerState) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing player state: %v", err)
	}
	return b, nil
}

type gameState struct {
	Current sudoku.Sudoku `json:"current"`
	Initial sudoku.Sudoku `json:"initial,omitempty"`
}

func (p *gameState) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing game state: %v", err)
	}
	return b, nil
}
