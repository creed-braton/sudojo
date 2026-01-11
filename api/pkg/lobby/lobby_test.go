package lobby

import (
	"math/rand"
	"runtime"
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sync"
	"testing"
)

func TestPlayer(t *testing.T) {
	t.Run("existing player", func(t *testing.T) {
		l := NewMock(true, false, true, 4)

		token, err := l.Create("alice")
		if err != nil {
			t.Fatalf("unexpected create error: '%v'", err)
		}

		p := l.Player(token)
		if p == nil {
			t.Fatal("expected player, got nil")
		}
		if p.Name() != "alice" {
			t.Errorf("expected name 'alice', got '%s'", p.Name())
		}
	})

	t.Run("non-existing player", func(t *testing.T) {
		l := NewMock(true, false, true, 4)

		token := player.NewToken()
		p := l.Player(token)
		if p != nil {
			t.Errorf("expected nil, got player '%v'", p)
		}
	})
}

func TestCreate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		l := NewMock(true, false, true, 4)

		_, err := l.Create("username123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("full lobby", func(t *testing.T) {
		size := 4
		l := NewMock(true, false, true, size)

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
		l := NewMock(true, false, true, 4)

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
		l := NewMock(true, false, true, 4)

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
		l := NewMock(true, false, true, 4)

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
		l := NewMock(true, false, true, 4)

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
		l := NewMock(true, true, true, 4)

		_, err := l.Insert(8, 8, 10, "", int64(42)) // invalid value range
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != game.ErrStrictValRange {
			t.Errorf("expected error: '%v', got: '%v'", game.ErrStrictValRange, err)
		}
	})

	t.Run("lax insert call", func(t *testing.T) {
		l := NewMock(true, false, true, 4)

		_, err := l.Insert(8, 8, 10, "", int64(42)) // invalid value range
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
			l := NewMock(true, tt.strict, true, 4)

			l.Insert(tt.input[0], tt.input[1], tt.input[2], "", int64(42))
			got := len(l.history.Artifacts()) > 0
			if got != tt.want {
				t.Errorf("expected history empty: '%t', got: '%t'", tt.want, got)
			}
		})
	}

	t.Run("history on not started game", func(t *testing.T) {
		want := false
		l := NewMock(false, true, true, 4)

		l.Insert(0, 0, 7, "", int64(42))
		got := len(l.history.Artifacts()) > 0
		if got != want {
			t.Errorf("expected history empty: '%t', got: '%t'", want, got)
		}
	})

	t.Run("history on finished game", func(t *testing.T) {
		want := false
		l := NewMock(true, true, true, 4)
		l.game.Finish(int64(42))

		l.Insert(0, 0, 7, "", int64(42))
		got := len(l.history.Artifacts()) > 0
		if got != want {
			t.Errorf("expected history empty: '%t', got: '%t'", want, got)
		}
	})
}

func TestUnderLoad(t *testing.T) {
	var wg sync.WaitGroup
	iterations := 1000

	for i := 0; i < 100; i++ {
		l := NewMock(true, true, true, 8)

		tokens := make([]string, 8)
		for j := 0; j < 8; j++ {
			token, err := l.Create("")
			if err != nil {
				t.Fatalf("unexpected error creating player %d: '%v'", j, err)
			}
			tokens[j] = token
		}

		for _, token := range tokens {
			wg.Add(1)
			go func(token string) {
				defer wg.Done()

				for k := 0; k < iterations; k++ {
					if rand.Intn(10) == 0 {
						runtime.Gosched()
					}

					switch rand.Intn(2) {
					case 0:
						l.Join(token)
					case 1:
						l.Leave(token)
					}
				}
			}(token)
		}
	}

	wg.Wait()
}
