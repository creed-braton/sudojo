package lobby

import (
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"testing"

	"github.com/google/uuid"
)

const now = int64(42)

func setupLobby(strict bool, maxSize int) *lobby {
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
	current.SetCell(8, 8, solution.Cell(8, 8))
	game, _ := game.New(current, initial, solution, nil, nil, "joker")

	config, _ := NewConfig(strict, false, false, maxSize)

	return New(
		uuid.NewString(),
		config,
		game,
		history.New(nil),
		make(map[string]player.Player),
	)
}

func TestCreate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		_, err := l.Create("username123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("full lobby", func(t *testing.T) {
		size := 4
		l := setupLobby(false, size)
		l.game.Start(now)

		for range size {
			_, err := l.Create("")
			if err != nil {
				t.Errorf("unexpected create error '%v'", err)
			}
		}
		_, err := l.Create("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrLobbyFull {
			t.Errorf("expected '%v', got '%v'", ErrLobbyFull, err)
		}
	})
}

func TestJoin(t *testing.T) {
	t.Run("non-existing player", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		token := player.NewToken()
		_, err := l.Join(token)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrPlayerNotFound {
			t.Errorf("expected '%v', got '%v'", ErrPlayerNotFound, err)
		}
	})

	t.Run("join player", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected create error '%v'", err)
		}
		players, err := l.Join(token)
		if err != nil {
			t.Errorf("unexpected join error '%v'", err)
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})

	t.Run("join initial player", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected create error: '%v'", err)
		}
		l = New(l.id, l.config, l.game, l.history, l.players)

		players, err := l.Join(token)
		if err != nil {
			t.Errorf("unexpected join error: '%v'", err)
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})
}

func TestLeave(t *testing.T) {
	t.Run("leave player", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected create error: '%v'", err)
		}
		_, err = l.Join(token)
		if err != nil {
			t.Errorf("unexpected join error: '%v'", err)
		}
		players, err := l.Leave(token)
		if err != nil {
			t.Errorf("unexpected leave error: '%v'", err)
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if players[0].Active() {
			t.Error("expected player to be inactive")
		}
	})
}

func TestInsert(t *testing.T) {
	t.Run("strict insert call", func(t *testing.T) {
		l := setupLobby(true, 4)
		l.game.Start(now)

		_, err := l.Insert(8, 8, 10, "", now) // invalid value range
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != game.ErrStrictValRange {
			t.Errorf("expected error: '%v', got: '%v'", game.ErrStrictValRange, err)
		}
	})

	t.Run("lax insert call", func(t *testing.T) {
		l := setupLobby(false, 4)
		l.game.Start(now)

		_, err := l.Insert(8, 8, 10, "", now) // invalid value range
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != game.ErrLaxValRange {
			t.Errorf("expected error: '%v', got: '%v'", game.ErrLaxValRange, err)
		}
	})

	tests := []struct {
		name   string
		strict bool
		input  [3]int
		want   bool
	}{
		{name: "history on initial clue", strict: true, input: [3]int{0, 7, 1}, want: false},
		{name: "history on already inserted", strict: true, input: [3]int{8, 8, 6}, want: false},
		{name: "history on invalid strict value range", strict: true, input: [3]int{8, 8, 0}, want: false},
		{name: "history on invalid lax value range", strict: false, input: [3]int{8, 8, 10}, want: false},
		{name: "history on invalid bounds", strict: true, input: [3]int{9, 9, 1}, want: false},
		{name: "history on strict incorrect value", strict: true, input: [3]int{0, 0, 1}, want: true},
		{name: "history on strict correct value", strict: true, input: [3]int{0, 0, 7}, want: true},
		{name: "history on lax incorrect value", strict: false, input: [3]int{0, 0, 1}, want: true},
		{name: "history on lax correct value", strict: false, input: [3]int{0, 0, 7}, want: true},
		{name: "history on lax row conflict", strict: false, input: [3]int{0, 0, 1}, want: true},
		{name: "history on lax column conflict", strict: false, input: [3]int{0, 0, 4}, want: true},
		{name: "history on lax box conflict", strict: false, input: [3]int{0, 4, 2}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := setupLobby(tt.strict, 4)
			l.game.Start(now)

			l.Insert(tt.input[0], tt.input[1], tt.input[2], "", now)
			got := len(l.history.Artifacts()) > 0
			if got != tt.want {
				t.Errorf("expected history empty: '%t', got: '%t'", tt.want, got)
			}
		})
	}

	t.Run("history on not started game", func(t *testing.T) {
		want := false
		l := setupLobby(true, 4)

		l.Insert(0, 0, 7, "", now)
		got := len(l.history.Artifacts()) > 0
		if got != want {
			t.Errorf("expected history empty: '%t', got: '%t'", want, got)
		}
	})

	t.Run("history on finished game", func(t *testing.T) {
		want := false
		l := setupLobby(true, 4)
		l.game.Start(now)
		l.game.Finish(now)

		l.Insert(0, 0, 7, "", now)
		got := len(l.history.Artifacts()) > 0
		if got != want {
			t.Errorf("expected history empty: '%t', got: '%t'", want, got)
		}
	})
}

func TestUnderLoad(t *testing.T) {
	l := setupLobby(true, 8)
	l.game.Start(now)
	// TODO
}
