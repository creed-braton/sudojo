package manager

import (
	"encoding/json"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
)

type gamePayload struct {
	Current  sudoku.Sudoku `json:"current,omitempty"`
	Initial  sudoku.Sudoku `json:"initial,omitempty"`
	Row      *int          `json:"row,omitempty"`
	Column   *int          `json:"column,omitempty"`
	Value    *int          `json:"value,omitempty"`
	Conflict string        `json:"conflict,omitempty"`
}

func (p *gamePayload) Marshal() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("error serializing game payload: %v", err)
	}
	return b, nil
}

type Game interface {
	Game() game.Game
	Strict() bool
	State(p player.Player) error
	Insert(row, col, val int, p player.Player, trace string) error
	Ping(row, col int, p player.Player, trace string) error
}

type gameManager struct {
	game   game.Game
	bus    event.EventBus
	strict bool
}

var _ Game = &gameManager{}

func NewGameManager(game game.Game, bus event.EventBus, strict bool) *gameManager {
	return &gameManager{game: game, bus: bus, strict: strict}
}

func (m *gameManager) Game() game.Game {
	return m.game
}

func (m *gameManager) Strict() bool {
	return m.strict
}

func (m *gameManager) State(p player.Player) error {
	payload := &gamePayload{Current: m.game.Current(), Initial: m.game.Initial()}
	e := event.New(event.StateEvent, p.Token(), "", "", payload)
	return p.Send(e)
}

func (m *gameManager) Insert(row, col, val int, p player.Player, trace string) error {
	return nil
}

func (m *gameManager) Ping(row, col int, p player.Player, trace string) error {
	return nil
}
