package game

import (
	"errors"
	"fmt"
	"sudojo/pkg/sudoku"
	"sync"
	"sync/atomic"
)

var (
	ErrIncorrect      = errors.New("input value incorrect")
	ErrRowConflict    = errors.New("value already exist in row")
	ErrColConflict    = errors.New("value already exist in column")
	ErrBoxConflict    = errors.New("value already exist in box")
	ErrOutOfBounds    = errors.New("cell position out of bounds")
	ErrFinished       = errors.New("game is already finished")
	ErrNotFinished    = errors.New("game is not finished")
	ErrNotStarted     = errors.New("game has not started yet")
	ErrDifficulty     = errors.New("invalid game difficulty")
	ErrLaxValRange    = fmt.Errorf("input value must be between %d and %d", sudoku.EmptyCell, sudoku.MaxValue)
	ErrStrictValRange = fmt.Errorf("input value must be between %d and %d", sudoku.MinValue, sudoku.MaxValue)
	ErrInitialClue    = errors.New("cannot overwrite initial clue")
)

const (
	Easy    = "easy"
	Medium  = "medium"
	Hard    = "hard"
	Expert  = "expert"
	Extreme = "extreme"
	Joker   = "joker"
)

// Represents a Sudoku game that encapsulates the current puzzle state, original
// puzzle setup, complete solution, and additional metadata. Safe for concurrent use.
type Game interface {
	// Hash of initial board state to identify same games.
	Hash() string
	// Inserts a value into the current board if within bounds, not an initial clue,
	// and causes no row, column, or box conflicts. Returns the updated board if
	// inserted, otherwise nil, and an error for invalid or conflicting inputs.
	// The now parameter is used as the finish timestamp if the move completes the board.
	// Returns ErrOutOfBounds, ErrLaxValRange, ErrInitialClue, ErrNotStarted,
	// ErrFinished, or ErrRowConflict, ErrColConflict, ErrBoxConflict on conflict.
	// Safe for concurrent use.
	Lax(row, col, val int, now int64) (sudoku.Sudoku, error)
	// Inserts a value into the current board if within bounds and not an initial
	// clue. The correct solution value is placed regardless of the input, and an
	// error is returned if the provided value does not match the solution.
	// The now parameter is used as the finish timestamp if the move completes the board.
	// Returns ErrOutOfBounds, ErrStrictValRange, ErrInitialClue, ErrNotStarted,
	// ErrFinished, or ErrIncorrect if the value doesn't match the solution.
	// Safe for concurrent use.
	Strict(row, col, val int, now int64) (sudoku.Sudoku, error)
	// Returns a copy of the current board state. Safe for concurrent use.
	Current() sudoku.Sudoku
	// Returns the initial board state. The returned value is immutable and must
	// not be altered; safe for concurrent reads.
	Initial() sudoku.Sudoku
	// Returns the solution of the puzzle. The returned value is immutable and must
	// not be altered; safe for concurrent reads.
	Solution() sudoku.Sudoku
	// Puts the game in started state by setting the start timestamp if not already set.
	Start(now int64)
	// Returns nanosecond start timestamp, nil if not started.
	StartedAt() *int64
	// Puts the game in finished state by setting the finish timestamp if not already set.
	// Game must be started otherwise will return ErrNotStarted.
	Finish(now int64) error
	// Returns nanosecond finish timestamp, nil if not finished.
	FinishedAt() *int64
	// Returns the difficulty of the game (easy, medium, hard, expert, extreme, joker).
	Difficulty() string
}

type game struct {
	hash       string
	current    sudoku.Sudoku // mutable
	initial    sudoku.Sudoku // immutable
	solution   sudoku.Sudoku // immutable
	started    atomic.Pointer[int64]
	finished   atomic.Pointer[int64]
	difficulty string
	lock       sync.RWMutex // mutex to synchronize operations on current state
}

var _ Game = &game{}

// Creates a game with the provided board states, timestamps, and difficulty.
// Returns ErrDifficulty if the difficulty is not valid.
func New(
	current, initial, solution sudoku.Sudoku,
	started, finished *int64, difficulty string,
) (*game, error) {
	if difficulty != Easy && difficulty != Medium &&
		difficulty != Hard && difficulty != Expert &&
		difficulty != Extreme && difficulty != Joker {
		return nil, ErrDifficulty
	}

	g := &game{
		current:    current,
		initial:    initial,
		solution:   solution,
		difficulty: difficulty,
	}
	if started != nil {
		g.started.Store(started)
	}
	if finished != nil {
		g.finished.Store(finished)
	}
	return g, nil
}

// Generates a full valid Sudoku solution, derives a uniquely solvable puzzle,
// and initializes the current game state. The generated game is of joker
// difficulty.
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
		current:    current,
		initial:    initial,
		solution:   solution,
		difficulty: Joker,
	}
}

func (g *game) Hash() string {
	return g.initial.Hash()
}

// Updates the current board with the given value and marks the game as finished
// using the provided now timestamp if the board becomes complete. Returns a deep
// copy of the updated state. Isn't concurrency-safe so write lock must be held
// by the caller.
func (g *game) insert(row, col, val int, now int64) sudoku.Sudoku {
	g.current.SetCell(row, col, val)
	if g.current.Complete() {
		g.finished.Store(&now)
	}

	current := sudoku.New()
	g.current.Copy(current)
	return current
}

func (g *game) Lax(row, col, val int, now int64) (sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if val != sudoku.EmptyCell && !sudoku.ValidVal(val) {
		return nil, ErrLaxValRange
	}
	if g.initial.Cell(row, col) != sudoku.EmptyCell {
		return nil, ErrInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started.Load() == nil {
		return nil, ErrNotStarted
	}
	if g.finished.Load() != nil {
		return nil, ErrFinished
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

	return g.insert(row, col, val, now), nil
}

func (g *game) Strict(row, col, val int, now int64) (sudoku.Sudoku, error) {
	if !sudoku.ValidBounds(row, col) {
		return nil, ErrOutOfBounds
	}
	if !sudoku.ValidVal(val) {
		return nil, ErrStrictValRange
	}
	if g.initial.Cell(row, col) != sudoku.EmptyCell {
		return nil, ErrInitialClue
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started.Load() == nil {
		return nil, ErrNotStarted
	}
	if g.finished.Load() != nil {
		return nil, ErrFinished
	}
	if g.current.Cell(row, col) == val {
		return nil, nil
	}

	var err error
	if g.solution.Cell(row, col) != val {
		err = ErrIncorrect
	}

	return g.insert(row, col, g.solution.Cell(row, col), now), err
}

func (g *game) Current() sudoku.Sudoku {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if g.started.Load() == nil {
		return nil
	}

	current := sudoku.New()
	g.current.Copy(current)
	return current
}

func (g *game) Initial() sudoku.Sudoku {
	if g.started.Load() == nil {
		return nil
	}
	return g.initial
}

func (g *game) Solution() sudoku.Sudoku {
	return g.solution
}

func (g *game) Start(now int64) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started.Load() == nil {
		g.started.Store(&now)
	}
}

func (g *game) StartedAt() *int64 {
	return g.started.Load()
}

func (g *game) Finish(now int64) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.started.Load() == nil {
		return ErrNotStarted
	}

	if g.finished.Load() == nil {
		g.finished.Store(&now)
	}
	return nil
}

func (g *game) FinishedAt() *int64 {
	return g.finished.Load()
}

func (g *game) Difficulty() string {
	return g.difficulty
}
