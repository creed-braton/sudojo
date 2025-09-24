package game

import (
	"errors"
	"sudojo/domain/sudoku"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	Id       uuid.UUID
	Current  *sudoku.Sudoku
	Initial  *sudoku.Sudoku
	Solution *sudoku.Sudoku
	Lock     sync.RWMutex
	Created  int64
	Finished *int64
}

func New() *Game {
	seed := time.Now().UTC().UnixNano()

	s := sudoku.New()
	s.Fill(seed)

	solution := sudoku.New()
	s.Copy(solution)

	initial := sudoku.New()
	s.Copy(initial)
	initial.GeneratePuzzle(seed)

	current := sudoku.New()
	initial.Copy(current)

	return &Game{
		Id:       uuid.New(),
		Current:  current,
		Initial:  initial,
		Solution: solution,
		Created:  time.Now().UTC().UnixNano(),
		Finished: nil,
	}
}

func (g *Game) Move(row, col, val int) (bool, error) {
	if g.Finished != nil {
		return false, errors.New("game is already finished")
	}

	g.Lock.Lock()
	defer g.Lock.Unlock()

	update, err := g.Current.Insert(row, col, val)
	if err != nil {
		return update, err
	}

	if update && g.Current.Complete() {
		solved := time.Now().UTC().UnixNano()
		g.Finished = &solved
	}

	return update, nil
}
