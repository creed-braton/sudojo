package manager

import (
	"encoding/json"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
)

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
	payload := &gameState{Current: m.game.Current(), Initial: m.game.Initial()}
	e := event.New("state", p.Token(), "", "", payload)
	return p.Send(e)
}

func (m *gameManager) Insert(row, col, val int, p player.Player, trace string) error {
	return nil
}

func (m *gameManager) Ping(row, col int, p player.Player, trace string) error {
	return nil
}
