package game

import (
	"errors"
	"fmt"
	"sudojo/internal/domain/sudoku"
	"sync"
	"time"
)

// Represents a Sudoku game that encapsulates the current puzzle state, original
// puzzle setup, complete solution, and additional metadata. Provides thread-safe
// methods for interacting with the shared mutable current board state.
type Game struct {
	Current  *sudoku.Sudoku `json:"current"`     // current board state (mutable)
	Initial  *sudoku.Sudoku `json:"initial"`     // initial board state (immutable)
	Solution *sudoku.Sudoku `json:"solution"`    // solution of the board (immutable)
	Seed     *int64         `json:"-"`           // seed of puzzle generation, nil if not generated
	Started  *int64         `json:"started_at"`  // nano second start timestamp, nil if not started
	Finished *int64         `json:"finished_at"` // nano second finish timestamp, nil if not finished
	lock     sync.RWMutex   // mutex to synchronize operations on current state
}

// Creates a new Game using the specified seed and validation mode. It generates
// a full valid Sudoku solution, derives a uniquely solvable puzzle, and
// initializes the current game state. The strict flag sets which validation
// function is applied for inserts.
func New(seed int64) *Game {
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
		Current:  current,
		Initial:  initial,
		Solution: solution,
		Seed:     &seed,
	}
}

// Puts the game in started state by setting the start timestamp if not already set.
func (g *Game) Start() {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.Started == nil {
		start := time.Now().UTC().UnixNano()
		g.Started = &start
	}
}

// Updates the current board with the given value and marks the game as finished
// if the board becomes complete. Returns a deep copy of the updated state.
func (g *Game) insert(row, col, val int) *sudoku.Sudoku {
	g.Current[row][col] = val
	if g.Current.Complete() {
		finish := time.Now().UTC().UnixNano()
		g.Finished = &finish
	}

	current := sudoku.New()
	g.Current.Copy(current)
	return current
}

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

// Thread-safely inserts a value into the current board if within bounds, not
// an initial clue, and causes no row, column, or box conflicts. Returns the
// updated board if inserted, otherwise nil, and an error for invalid or
// conflicting inputs.
func (g *Game) Lax(row, col, val int) (*sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if val != sudoku.EmptyCell && !sudoku.ValidVal(val) {
		return nil, errLaxValRange
	}
	if g.Initial[row][col] != sudoku.EmptyCell {
		return nil, errInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.Started == nil {
		return nil, errNotStarted
	}
	if g.Finished != nil {
		return nil, ErrAlreadyFinished
	}
	if g.Current[row][col] == val {
		return nil, nil
	}

	if val != sudoku.EmptyCell {
		if !g.Current.ValidRow(row, val) {
			return nil, ErrRowConflict
		}
		if !g.Current.ValidCol(col, val) {
			return nil, ErrColConflict
		}
		if !g.Current.ValidBox(row, col, val) {
			return nil, ErrBoxConflict
		}
	}

	return g.insert(row, col, val), nil
}

// Thread-safely inserts a value into the current board if within bounds and
// not an initial clue. The correct solution value is placed regardless of the
// input, and an error is returned if the provided value does not match the
// solution.
func (g *Game) Strict(row, col, val int) (*sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if !sudoku.ValidVal(val) {
		return nil, errStrictValRange
	}
	if g.Initial[row][col] != sudoku.EmptyCell {
		return nil, errInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.Started == nil {
		return nil, errNotStarted
	}
	if g.Finished != nil {
		return nil, ErrAlreadyFinished
	}
	if g.Current[row][col] == val {
		return nil, nil
	}

	var err error
	if g.Solution[row][col] != val {
		err = ErrIncorrect
	}

	return g.insert(row, col, g.Solution[row][col]), err
}

// Returns a thread-safe copy of the current board state.
func (g *Game) State() *sudoku.Sudoku {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if g.Started == nil {
		return nil
	}

	current := sudoku.New()
	g.Current.Copy(current)
	return current
}
