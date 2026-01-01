package game

import (
	"math/rand"
	"runtime"
	"sudojo/pkg/sudoku"
	"sync"
	"testing"
)

const now = int64(42)

func setUp() *game {
	initial := sudoku.NewFromInts(
		[][]int{
			{0, 0, 0, 0, 0, 0, 0, 1, 0},
			{0, 0, 0, 0, 0, 2, 0, 0, 3},
			{0, 0, 0, 4, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 5, 0, 0},
			{4, 0, 1, 6, 0, 0, 0, 0, 0},
			{0, 0, 7, 1, 0, 0, 0, 0, 0},
			{0, 5, 0, 0, 0, 0, 2, 0, 0},
			{0, 0, 0, 0, 8, 0, 0, 4, 0},
			{0, 3, 0, 9, 1, 0, 0, 0, 0},
		},
	)
	current := sudoku.New()
	initial.Copy(current)
	solution := sudoku.NewFromInts(
		[][]int{
			{7, 4, 5, 3, 6, 8, 9, 1, 2},
			{8, 1, 9, 5, 7, 2, 4, 6, 3},
			{3, 6, 2, 4, 9, 1, 8, 5, 7},
			{6, 9, 3, 8, 2, 4, 5, 7, 1},
			{4, 2, 1, 6, 5, 7, 3, 9, 8},
			{5, 8, 7, 1, 3, 9, 6, 2, 4},
			{1, 5, 8, 7, 4, 6, 2, 3, 9},
			{9, 7, 6, 2, 8, 3, 1, 4, 5},
			{2, 3, 4, 9, 1, 5, 7, 8, 6},
		},
	)
	// insert one value so it differs from initial board
	current[8][8] = solution[8][8]

	return &game{
		initial:  initial,
		current:  current,
		solution: solution,
	}
}

func TestStart(t *testing.T) {
	t.Run("initialized game", func(t *testing.T) {
		if setUp().StartedAt() != nil {
			t.Fatal("expected game to be not started initially")
		}
	})

	t.Run("starting game", func(t *testing.T) {
		g := setUp()
		g.Start(now)

		got := g.StartedAt()
		if got == nil {
			t.Fatal("expected start timestamp to be not nil after start")
		}
		if *got != now {
			t.Error("expected start timestamp to be inserted timestamp")
		}
	})

	t.Run("starting game again", func(t *testing.T) {
		g := setUp()
		g.Start(int64(0))
		want := *g.StartedAt()
		g.Start(int64(1))

		if want != *g.StartedAt() {
			t.Error("expected start timestamp once set to not be overwritten")
		}
	})

	t.Run("strict insert on non-started game", func(t *testing.T) {
		row, col, val := 0, 0, 5
		s, err := setUp().Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrNotStarted {
			t.Errorf("expected error: '%v', got: '%v'", ErrNotStarted, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("lax insert on non-started game", func(t *testing.T) {
		row, col, val := 0, 0, 5
		s, err := setUp().Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrNotStarted {
			t.Errorf("expected error: '%v', got: '%v'", ErrNotStarted, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("get current board on non-started game", func(t *testing.T) {
		if setUp().Current() != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("get initial board on non-started game", func(t *testing.T) {
		if setUp().Initial() != nil {
			t.Error("expected nil, got initial board")
		}
	})
}

func TestFinish(t *testing.T) {
	t.Run("initialized game", func(t *testing.T) {
		if setUp().FinishedAt() != nil {
			t.Fatal("expected game to be not finished initially")
		}
	})

	t.Run("setting game finish", func(t *testing.T) {
		g := setUp()
		g.Start(now)

		err := g.Finish(now)
		if err != nil {
			t.Fatalf("expected nil, got: '%v'", err)
		}

		got := g.FinishedAt()
		if got == nil {
			t.Fatal("expected finish timestamp to be not nil after finish")
		}
		if *got != now {
			t.Error("expected finish timestamp to be inserted timestamp")
		}
	})

	t.Run("setting non-started game finish", func(t *testing.T) {
		g := setUp()

		err := g.Finish(now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrNotStarted {
			t.Errorf("expected: '%v', got: '%v'", ErrNotStarted, err)
		}

		got := g.FinishedAt()
		if got != nil {
			t.Errorf("expected nil, got finish timestamp: %d", *got)
		}
	})

	t.Run("setting game finish again", func(t *testing.T) {
		g := setUp()
		g.Start(now)
		g.Finish(now)
		want := *g.FinishedAt()
		g.Finish(int64(0))

		if want != *g.FinishedAt() {
			t.Error("expected finish timestamp once set to not be overwritten")
		}
	})

	t.Run("strict insert finishing game", func(t *testing.T) {
		g := setUp()
		g.Start(now)

		for row := range sudoku.BoardSize {
			for col := range sudoku.BoardSize {
				val := g.Solution().Cell(row, col)
				g.Strict(row, col, val, now)
			}
		}

		if g.FinishedAt() == nil {
			t.Fatal("expected finish timestamp to be not nil after game completion")
		}
		if *g.FinishedAt() != now {
			t.Error("expected finish timestamp to be strict inserted timestamp")
		}
	})

	t.Run("strict insert on already finished game", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start(now)
		g.Finish(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != ErrFinished {
			t.Errorf("expected error: '%v', got: '%v'", ErrFinished, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("lax insert finishing game", func(t *testing.T) {
		g := setUp()
		g.Start(now)

		for row := range sudoku.BoardSize {
			for col := range sudoku.BoardSize {
				val := g.Solution().Cell(row, col)
				g.Lax(row, col, val, now)
			}
		}

		if g.FinishedAt() == nil {
			t.Fatal("expected finish timestamp to be not nil after game completion")
		}
		if *g.FinishedAt() != now {
			t.Error("expected finish timestamp to be lax inserted timestamp")
		}
	})

	t.Run("lax insert on already finished game", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start(now)
		g.Finish(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != ErrFinished {
			t.Errorf("expected error: '%v', got: '%v'", ErrFinished, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})
}

func TestStrict(t *testing.T) {
	t.Run("incorrect value", func(t *testing.T) {
		row, col, val := 0, 0, 8
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrIncorrect {
			t.Errorf("expected error: '%v', got: '%v'", ErrIncorrect, err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 4
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrIncorrect {
			t.Errorf("expected error: '%v', got: '%v'", ErrIncorrect, err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("correct value", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error '%v'", err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		row, col, val := 9, 0, 2
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrOutOfBounds {
			t.Errorf("expected error: '%v', got: '%v'", ErrOutOfBounds, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("invalid value range", func(t *testing.T) {
		row, col, val := 0, 0, 10
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrStrictValRange {
			t.Errorf("expected error: '%v', got: '%v'", ErrStrictValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("empty cell value", func(t *testing.T) {
		row, col, val := 0, 0, sudoku.EmptyCell
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrStrictValRange {
			t.Errorf("expected error: '%v', got: '%v'", ErrStrictValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("initial clue position", func(t *testing.T) {
		row, col, val := 4, 0, 4
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrInitialClue {
			t.Errorf("expected error: '%v', got: '%v'", ErrInitialClue, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("correct value already exist", func(t *testing.T) {
		row, col, val := 8, 8, 6
		g := setUp()
		g.Start(now)
		s, err := g.Strict(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error '%v'", err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})
}

func TestLax(t *testing.T) {
	t.Run("incorrect value", func(t *testing.T) {
		row, col, val := 0, 0, 8
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error: '%v'", err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != val {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("returned copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("row constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 1
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrRowConflict {
			t.Errorf("expected error: '%v', got: '%v'", ErrRowConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("column constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 4
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrColConflict {
			t.Errorf("expected error: '%v', got: '%v'", ErrColConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("box constraint conflict", func(t *testing.T) {
		row, col, val := 3, 0, 1
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrBoxConflict {
			t.Errorf("expected error: '%v', got: '%v'", ErrBoxConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("correct value", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error '%v'", err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != val {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		row, col, val := 9, 0, 2
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrOutOfBounds {
			t.Errorf("expected error: '%v', got: '%v'", ErrOutOfBounds, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("invalid value range", func(t *testing.T) {
		row, col, val := 0, 0, 10
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrLaxValRange {
			t.Errorf("expected error: '%v', got: '%v'", ErrLaxValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("empty cell value", func(t *testing.T) {
		row, col, val := 8, 8, sudoku.EmptyCell
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error '%v'", err)
		}
		if s == nil {
			t.Fatal("expected current board, got nil")
		}
		if s.Cell(row, col) != val {
			t.Error("current board not updated")
		}
		if !s.Equal(g.current) {
			t.Fatal("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("initial clue position", func(t *testing.T) {
		row, col, val := 4, 0, 4
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrInitialClue {
			t.Errorf("expected error: '%v', got: '%v'", ErrInitialClue, err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})

	t.Run("correct value already exist", func(t *testing.T) {
		row, col, val := 8, 8, 6
		g := setUp()
		g.Start(now)
		s, err := g.Lax(row, col, val, now)
		if err != nil {
			t.Errorf("expected nil, got error '%v'", err)
		}
		if s != nil {
			t.Error("expected nil, got current board")
		}
	})
}

func TestUnderLoad(t *testing.T) {
	var wg sync.WaitGroup
	num := 1000

	for i := 0; i < 100; i++ {
		g := Generate(int64(i))
		g.Start(now)

		for j := 0; j < num; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				if rand.Intn(10) == 0 {
					runtime.Gosched()
				}

				switch rand.Intn(2) {
				case 0:
					row := id % 9
					col := (id / 9) % 9
					val := (id % 9) + 1
					g.Lax(row, col, val, now)
				case 1:
					g.Current()
				}
			}(j)
		}
	}

	wg.Wait()
}
