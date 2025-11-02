package lobby

import (
	"sudojo/internal/domain/game"
	"sudojo/internal/domain/sudoku"
)

type GameManager struct {
	Strict bool
	Game   *game.Game
}

type gamePayload struct {
	Row      int            `json:"row"`
	Column   int            `json:"column"`
	Value    *int           `json:"value,omitempty"`
	Initial  *sudoku.Sudoku `json:"initial_state,omitempty"`
	Current  *sudoku.Sudoku `json:"current_state,omitempty"`
	Started  *int64         `json:"started_at,omitempty"`
	Finished *int64         `json:"finished_at,omitempty"`
	Error    string         `json:"error,omitempty"`
	Conflict string         `json:"conflict,omitempty"`
}

func newGameManager(seed int64, strict bool) *GameManager {
	g := game.New(seed)
	g.Start()
	return &GameManager{
		Strict: strict,
		Game:   g,
	}
}

func (m *GameManager) insert(row, col, val int) (*gamePayload, error) {
	var current *sudoku.Sudoku
	var err error

	if m.Strict {
		current, err = m.Game.Strict(row, col, val)
	} else {
		current, err = m.Game.Lax(row, col, val)
	}

	if current == nil && err == nil {
		return nil, nil
	}

	payload := &gamePayload{
		Row:      row,
		Column:   col,
		Value:    &val,
		Current:  current,
		Finished: m.Game.Finished,
	}
	if err == game.ErrRowConflict ||
		err == game.ErrColConflict ||
		err == game.ErrBoxConflict ||
		err == game.ErrIncorrect {
		payload.Conflict = err.Error()
	} else if err != nil {
		return nil, err
	}

	return payload, nil
}

func (m *GameManager) state() *gamePayload {
	return &gamePayload{
		Initial: m.Game.Initial,
		Current: m.Game.State(),
	}
}

func (m *GameManager) ping(row, col int) (*gamePayload, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, game.ErrOutOfBounds
	}
	return &gamePayload{
		Row:    row,
		Column: col,
	}, nil
}
