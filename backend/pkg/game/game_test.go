package game

import (
	"math/rand"
	"sudojo/pkg/sudoku"
	"sync"
	"testing"
	"time"
)

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
	current[8][8] = solution[8][8]

	return &game{
		initial:  initial,
		current:  current,
		solution: solution,
	}
}

func TestStart(t *testing.T) {
	t.Run("starting game", func(t *testing.T) {
		g := setUp()

		if g.Started() != nil {
			t.Fatal("expected game to be not started initially")
		}

		g.Start()

		started := g.Started()
		if started == nil {
			t.Fatal("expected Started() to be not nil after Start()")
		}
		if *started == 0 {
			t.Error("expected non zero timestamp from Start()")
		}
	})

	t.Run("starting game again", func(t *testing.T) {
		g := setUp()
		g.Start()
		first := *g.Started()

		time.Sleep(time.Second)
		g.Start()
		second := *g.Started()

		if first != second {
			t.Errorf("expected Start() to be idempotent, timestamps differ: first=%d second=%d", first, second)
		}
	})
}

func TestInsert(t *testing.T) {
	for i := 0; i < 100; i++ {
		g := Generate(int64(i))

		if g.Finished() != nil {
			t.Errorf("expected Finished() to be nil initially for game %d", i)
		}

		initialCopy := sudoku.New()
		g.Initial().Copy(initialCopy)
		solutionCopy := sudoku.New()
		g.Solution().Copy(solutionCopy)
		g.Start()

		// Fill all empty cells with solution values
		for row := 0; row < 9; row++ {
			for col := 0; col < 9; col++ {
				if g.Current().Cell(row, col) == sudoku.EmptyCell {
					_, err := g.Strict(row, col, g.Solution().Cell(row, col))
					if err != nil {
						t.Fatalf("unexpected error inserting solution at Cell(%d,%d) in game %d: %v", row, col, i, err)
					}

					if row != 8 || col != 8 { // not last cell
						if g.Finished() != nil && g.Current().Complete() == false {
							t.Errorf("expected Finished() to be nil before final cell in game %d", i)
						}
					}
				}
			}
		}

		if g.Finished() == nil {
			t.Errorf("expected Finished() to be not nil after completing game %d", i)
		}

		if !initialCopy.Equal(g.Initial()) {
			t.Errorf("expected initial board unchanged for game %d", i)
		}
		if !solutionCopy.Equal(g.Solution()) {
			t.Errorf("expected solution board unchanged for game %d", i)
		}
	}
}

func TestStrict(t *testing.T) {
	t.Run("not started", func(t *testing.T) {
		row, col, val := 0, 0, 5
		g := setUp()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errNotStarted {
			t.Errorf("want: %v, got: %v", errNotStarted, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("already finished", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start()
		now := time.Now().UTC().UnixNano()
		g.finished = &now
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrAlreadyFinished {
			t.Errorf("want: %v, got: %v", ErrAlreadyFinished, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("incorrect value", func(t *testing.T) {
		row, col, val := 0, 0, 8
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrIncorrect {
			t.Errorf("want: %v, got: %v", ErrIncorrect, err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 4
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrIncorrect {
			t.Errorf("want: %v, got: %v", ErrIncorrect, err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("correct value", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != g.solution.Cell(row, col) {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		row, col, val := 9, 0, 2
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrOutOfBounds {
			t.Errorf("want: %v, got: %v", ErrOutOfBounds, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("invalid value range", func(t *testing.T) {
		row, col, val := 0, 0, 10
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errStrictValRange {
			t.Errorf("want: %v, got: %v", errStrictValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("empty cell value", func(t *testing.T) {
		row, col, val := 0, 0, sudoku.EmptyCell
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errStrictValRange {
			t.Errorf("want: %v, got: %v", errStrictValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("initial clue position", func(t *testing.T) {
		row, col, val := 4, 0, 4
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errInitialClue {
			t.Errorf("want: %v, got: %v", errInitialClue, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("correct value already exist", func(t *testing.T) {
		row, col, val := 8, 8, 6
		g := setUp()
		g.Start()
		s, err := g.Strict(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})
}

func TestLax(t *testing.T) {
	t.Run("not started", func(t *testing.T) {
		row, col, val := 0, 0, 5
		g := setUp()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errNotStarted {
			t.Errorf("want: %v, got: %v", errNotStarted, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("already finished", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start()
		now := time.Now().UTC().UnixNano()
		g.finished = &now
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrAlreadyFinished {
			t.Errorf("want: %v, got: %v", ErrAlreadyFinished, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("incorrect value", func(t *testing.T) {
		row, col, val := 0, 0, 8
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != val {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("row constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 1
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrRowConflict {
			t.Errorf("want: %v, got: %v", ErrRowConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("column constraint conflict", func(t *testing.T) {
		row, col, val := 0, 0, 4
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrColConflict {
			t.Errorf("want: %v, got: %v", ErrColConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("box constraint conflict", func(t *testing.T) {
		row, col, val := 3, 0, 1
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrBoxConflict {
			t.Errorf("want: %v, got: %v", ErrBoxConflict, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("correct value", func(t *testing.T) {
		row, col, val := 0, 0, 7
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != val {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		row, col, val := 9, 0, 2
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != ErrOutOfBounds {
			t.Errorf("want: %v, got: %v", ErrOutOfBounds, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("invalid value range", func(t *testing.T) {
		row, col, val := 0, 0, 10
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errLaxValRange {
			t.Errorf("want: %v, got: %v", errLaxValRange, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("empty cell value", func(t *testing.T) {
		row, col, val := 8, 8, sudoku.EmptyCell
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s == nil {
			t.Error("expected current board copy, got nil")
			return
		}
		if s.Cell(row, col) != val {
			t.Error("current board copy not updated")
		}
		if !s.Equal(g.current) {
			t.Error("copy and current board not equal")
		}
		g.solution.Copy(s)
		if s.Equal(g.current) {
			t.Error("copy not pointer independent to current")
		}
	})

	t.Run("initial clue position", func(t *testing.T) {
		row, col, val := 4, 0, 4
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err == nil {
			t.Error("expected error, got nil")
			return
		}
		if err != errInitialClue {
			t.Errorf("want: %v, got: %v", errInitialClue, err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})

	t.Run("correct value already exist", func(t *testing.T) {
		row, col, val := 8, 8, 6
		g := setUp()
		g.Start()
		s, err := g.Lax(row, col, val)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if s != nil {
			t.Error("expected nil, got current board copy")
		}
	})
}

func TestUnderLoad(t *testing.T) {
	g := setUp()
	g.Start()

	var wg sync.WaitGroup
	num := 1000

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			row := id % 9
			col := (id / 9) % 9
			val := (id % 9) + 1
			_, _ = g.Lax(row, col, val)
		}(i)
	}

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = g.Current()
		}()
	}

	wg.Wait()
}

func BenchmarkGame(b *testing.B) {
	g := setUp()
	g.Start()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			row, col, val := rand.Intn(9), rand.Intn(9), rand.Intn(10)
			_, _ = g.Lax(row, col, val)
			_ = g.Current()
		}
	})
}
