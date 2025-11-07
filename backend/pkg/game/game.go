package game

import (
	"errors"
	"fmt"
	"sudojo/pkg/sudoku"
	"sync"
	"time"
)

var (
	ErrIncorrect       = errors.New("input incorrect")
	ErrRowConflict     = errors.New("value already exist in row")
	ErrColConflict     = errors.New("value already exist in column")
	ErrBoxConflict     = errors.New("value already exist in box")
	ErrOutOfBounds     = errors.New("cell position out of bounds")
	errLaxValRange     = fmt.Errorf("input must be between %d and %d", sudoku.EmptyCell, sudoku.MaxValue)
	errStrictValRange  = fmt.Errorf("input must be between %d and %d", sudoku.MinValue, sudoku.MaxValue)
	errInitialClue     = errors.New("cannot overwrite initial clue")
	ErrAlreadyFinished = errors.New("game is already finish")
	errNotStarted      = errors.New("game has not started yet")
)

// Represents a Sudoku game that encapsulates the current puzzle state, original
// puzzle setup, complete solution, and additional metadata. Provides thread-safe
// methods for interacting with the shared mutable current board state.
type Game interface {
	// Puts the game in started state by setting the start timestamp if not already set.
	Start()
	// Thread-safely inserts a value into the current board if within bounds, not
	// an initial clue, and causes no row, column, or box conflicts. Returns the
	// updated board if inserted, otherwise nil, and an error for invalid or
	// conflicting inputs.
	Lax(row, col, val int) (sudoku.Sudoku, error)
	// Thread-safely inserts a value into the current board if within bounds and
	// not an initial clue. The correct solution value is placed regardless of the
	// input, and an error is returned if the provided value does not match the
	// solution.
	Strict(row, col, val int) (sudoku.Sudoku, error)
	// Returns a thread-safe copy of the current board state.
	Current() sudoku.Sudoku
	// Returns the initial board state.
	Initial() sudoku.Sudoku
	// Returns the solution of the puzzle.
	Solution() sudoku.Sudoku
	// Returns nano second start timestamp, nil if not started.
	Started() *int64
	// Returns nano second finish timestamp, nil if not finished.
	Finished() *int64
}

type game struct {
	current  sudoku.Sudoku // mutable
	initial  sudoku.Sudoku // immutable
	solution sudoku.Sudoku // immutable
	started  *int64
	finished *int64
	lock     sync.RWMutex // mutex to synchronize operations on current state
}

var _ Game = &game{}

func New(current, initial, solution sudoku.Sudoku, started, finished *int64) *game {
	return &game{
		current:  current,
		initial:  initial,
		solution: solution,
		started:  started,
		finished: finished,
	}
}

// Generates a full valid Sudoku solution, derives a uniquely solvable puzzle,
// and initializes the current game state.
func Generate(seed int64) *game {
	s := sudoku.New()
	s.Fill(seed)

	solution := sudoku.New()
	s.Copy(solution)

	initial := sudoku.New()
	s.Copy(initial)
	initial.GeneratePuzzle(seed)

	current := sudoku.New()
	initial.Copy(current)

	return &game{
		current:  current,
		initial:  initial,
		solution: solution,
	}
}

func (g *game) Start() {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started == nil {
		start := time.Now().UTC().UnixNano()
		g.started = &start
	}
}

// Updates the current board with the given value and marks the game as finished
// if the board becomes complete. Returns a deep copy of the updated state.
func (g *game) insert(row, col, val int) sudoku.Sudoku {
	g.current.SetCell(row, col, val)
	if g.current.Complete() {
		finished := time.Now().UTC().UnixNano()
		g.finished = &finished
	}

	current := sudoku.New()
	g.current.Copy(current)
	return current
}

func (g *game) Lax(row, col, val int) (sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if val != sudoku.EmptyCell && !sudoku.ValidVal(val) {
		return nil, errLaxValRange
	}
	if g.initial.Cell(row, col) != sudoku.EmptyCell {
		return nil, errInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started == nil {
		return nil, errNotStarted
	}
	if g.finished != nil {
		return nil, ErrAlreadyFinished
	}
	if g.current.Cell(row, col) == val {
		return nil, nil
	}

	if val != sudoku.EmptyCell {
		if !g.current.ValidRow(row, val) {
			return nil, ErrRowConflict
		}
		if !g.current.ValidCol(col, val) {
			return nil, ErrColConflict
		}
		if !g.current.ValidBox(row, col, val) {
			return nil, ErrBoxConflict
		}
	}

	return g.insert(row, col, val), nil
}

func (g *game) Strict(row, col, val int) (sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if !sudoku.ValidVal(val) {
		return nil, errStrictValRange
	}
	if g.initial.Cell(row, col) != sudoku.EmptyCell {
		return nil, errInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started == nil {
		return nil, errNotStarted
	}
	if g.finished != nil {
		return nil, ErrAlreadyFinished
	}
	if g.current.Cell(row, col) == val {
		return nil, nil
	}

	var err error
	if g.solution.Cell(row, col) != val {
		err = ErrIncorrect
	}

	return g.insert(row, col, g.solution.Cell(row, col)), err
}

func (g *game) Current() sudoku.Sudoku {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if g.started == nil {
		return nil
	}

	current := sudoku.New()
	g.current.Copy(current)
	return current
}

func (g *game) Initial() sudoku.Sudoku {
	return g.initial
}

func (g *game) Solution() sudoku.Sudoku {
	return g.solution
}

func (g *game) Started() *int64 {
	return g.started
}

func (g *game) Finished() *int64 {
	return g.finished
}
