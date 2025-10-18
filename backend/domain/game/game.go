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
	lock     sync.RWMutex
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

	return &Game{Id: uuid.New(),
		Current:  current,
		Initial:  initial,
		Solution: solution,
		Created:  time.Now().UTC().UnixNano(),
		Finished: nil,
	}
}

func (g *Game) Insert(row, col, val int) (*sudoku.Sudoku, error) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.Finished != nil {
		return nil, errors.New("game is already finished")
	}

	if g.Current[row][col] == val {
		return nil, nil
	}

	err := g.Current.Validate(row, col, val)
	if err != nil {
		return nil, err
	}
	g.Current[row][col] = val

	if g.Current.Complete() {
		solved := time.Now().UTC().UnixNano()
		g.Finished = &solved
	}

	current := sudoku.New()
	g.Current.Copy(current)

	return current, nil
}

func (g *Game) State() (*sudoku.Sudoku, *sudoku.Sudoku) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	initial := sudoku.New()
	g.Initial.Copy(initial)
	current := sudoku.New()
	g.Current.Copy(current)

	return initial, current
}
