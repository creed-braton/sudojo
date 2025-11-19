package lobby

import (
	"sudojo/pkg/game"
	"sudojo/pkg/player"
	"sudojo/pkg/sudoku"
	"testing"

	"github.com/google/uuid"
)

func setupGame() game.Game {
	current := sudoku.New()
	initial := sudoku.New()
	solution := sudoku.New()

	return game.NewMockGame(
		func() {},
		func(row, col, val int) (sudoku.Sudoku, error) {
			return current, nil
		},
		func(row, col, val int) (sudoku.Sudoku, error) {
			return current, nil
		},
		func() sudoku.Sudoku { return current },
		func() sudoku.Sudoku { return initial },
		func() sudoku.Sudoku { return solution },
		func() *int64 { var t int64 = 123; return &t },
		func() *int64 { return nil },
	)
}

func setupPool(size int) player.Pool {
	return player.NewMockPool(
		func() int { return size },
		func(name string) (string, error) {
			return "test-token", nil
		},
		func(token string) ([]player.Player, error) {
			p := player.New(token, "test-player")
			p.SetActive(true)
			return []player.Player{p}, nil
		},
		func(token string) ([]player.Player, error) {
			p := player.New(token, "test-player")
			p.SetActive(false)
			return []player.Player{p}, nil
		},
		func() []player.Player {
			return []player.Player{}
		},
	)
}

func TestNew(t *testing.T) {
	id, strict, maxPlayer := uuid.NewString(), false, 8
	lobby := New(id, strict, setupGame(), setupPool(maxPlayer))

	if id != lobby.Id() {
		t.Errorf("expected id '%s' got '%s'", id, lobby.Id())
	}
	if strict != lobby.Strict() {
		t.Errorf("expected strict '%t' got '%t'", strict, lobby.Strict())
	}
	if maxPlayer != lobby.MaxPlayer() {
		t.Errorf("expected max player '%d' got '%d'", maxPlayer, lobby.MaxPlayer())
	}
}

func TestOpen(t *testing.T) {
	lobby := Open(8, false)

	if len(lobby.Id()) < 1 {
		t.Error("empty lobby id")
	}
	if lobby.game.Started() == nil {
		t.Error("lobby is not started")
	}
}

func TestCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		token, err := lobby.Create("test-player")

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if token == "" {
			t.Error("token is empty")
		}
	})

	t.Run("full lobby", func(t *testing.T) {
		fullPool := player.NewMockPool(
			func() int { return 0 },
			func(name string) (string, error) {
				return "", player.ErrPoolFull
			},
			func(token string) ([]player.Player, error) { return nil, nil },
			func(token string) ([]player.Player, error) { return nil, nil },
			func() []player.Player { return nil },
		)
		lobby := New(uuid.NewString(), false, setupGame(), fullPool)
		_, err := lobby.Create("player")

		if err != ErrLobbyFull {
			t.Errorf("expected '%v', got '%v'", ErrLobbyFull, err)
		}
	})

	t.Run("invalid player name", func(t *testing.T) {
		invalidPool := player.NewMockPool(
			func() int { return 8 },
			func(name string) (string, error) {
				return "", player.ErrNameTooLong
			},
			func(token string) ([]player.Player, error) { return nil, nil },
			func(token string) ([]player.Player, error) { return nil, nil },
			func() []player.Player { return nil },
		)
		lobby := New(uuid.NewString(), false, setupGame(), invalidPool)
		_, err := lobby.Create("")

		if err != player.ErrNameTooLong {
			t.Errorf("expected '%v', got '%v'", player.ErrNameTooLong, err)
		}
	})
}

func TestJoin(t *testing.T) {
	t.Run("successful join", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		payload, err := lobby.Join("token")

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		players := payload.Players()
		if len(players) == 0 {
			t.Error("expected at least one player in payload")
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})

	t.Run("player not found", func(t *testing.T) {
		mockPool := player.NewMockPool(
			func() int { return 8 },
			func(name string) (string, error) { return "", nil },
			func(token string) ([]player.Player, error) {
				return nil, player.ErrPlayerNotFound
			},
			func(token string) ([]player.Player, error) { return nil, nil },
			func() []player.Player { return nil },
		)
		lobby := New(uuid.NewString(), false, setupGame(), mockPool)
		_, err := lobby.Join("invalid-token")

		if err != player.ErrPlayerNotFound {
			t.Errorf("expected '%v', got '%v'", player.ErrPlayerNotFound, err)
		}
	})
}

func TestLeave(t *testing.T) {
	t.Run("successful leave", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		payload, err := lobby.Leave("token")

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		players := payload.Players()
		if len(players) == 0 {
			t.Error("expected at least one player in payload")
		}
		if players[0].Active() {
			t.Error("expected player to be inactive")
		}
	})

	t.Run("player not found", func(t *testing.T) {
		notFoundPool := player.NewMockPool(
			func() int { return 8 },
			func(name string) (string, error) { return "", nil },
			func(token string) ([]player.Player, error) { return nil, nil },
			func(token string) ([]player.Player, error) {
				return nil, player.ErrPlayerNotFound
			},
			func() []player.Player { return nil },
		)
		lobby := New(uuid.NewString(), false, setupGame(), notFoundPool)
		_, err := lobby.Leave("invalid-token")

		if err != player.ErrPlayerNotFound {
			t.Errorf("expected '%v', got '%v'", player.ErrPlayerNotFound, err)
		}
	})
}

func TestInsertStrict(t *testing.T) {
	t.Run("correct value", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, nil
			},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 0; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), true, mockGame, setupPool(8))
		payload, err := lobby.Insert(0, 0, 5)

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Conflict() != "" {
			t.Errorf("expected no conflict, got %s", payload.Conflict())
		}
		if payload.Current() == nil {
			t.Error("expected current board in payload")
		}
	})

	t.Run("incorrect value conflict", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, game.ErrIncorrect
			},
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 0; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), true, mockGame, setupPool(8))
		payload, err := lobby.Insert(0, 0, 5)

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Conflict() == "" {
			t.Error("expected conflict message")
		}
		if payload.Current() == nil {
			t.Error("expected current board in payload")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, game.ErrOutOfBounds
			},
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 0; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), true, mockGame, setupPool(8))
		_, err := lobby.Insert(0, 0, 5)

		if err != game.ErrOutOfBounds {
			t.Errorf("expected '%v', got '%v'", game.ErrOutOfBounds, err)
		}
	})
}

func TestInsertLax(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, nil
			},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 0; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), false, mockGame, setupPool(8))
		payload, err := lobby.Insert(0, 0, 5)

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Conflict() != "" {
			t.Errorf("expected no conflict, got '%s'", payload.Conflict())
		}
	})

	t.Run("sudoku conflict", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, game.ErrRowConflict
			},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 0; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), false, mockGame, setupPool(8))
		payload, err := lobby.Insert(0, 0, 5)

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Conflict() == "" {
			t.Error("expected conflict message")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		current := sudoku.New()
		mockGame := game.NewMockGame(
			func() {},
			func(row, col, val int) (sudoku.Sudoku, error) {
				return current, game.ErrOutOfBounds
			},
			func(row, col, val int) (sudoku.Sudoku, error) { return current, nil },
			func() sudoku.Sudoku { return current },
			func() sudoku.Sudoku { return sudoku.New() },
			func() sudoku.Sudoku { return sudoku.New() },
			func() *int64 { var t int64 = 123; return &t },
			func() *int64 { return nil },
		)
		lobby := New(uuid.NewString(), false, mockGame, setupPool(8))
		_, err := lobby.Insert(0, 0, 5)

		if err != game.ErrOutOfBounds {
			t.Errorf("expected '%v', got '%v'", game.ErrOutOfBounds, err)
		}
	})
}

func TestPing(t *testing.T) {
	t.Run("valid coordinates", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		row, col := 5, 7
		payload, err := lobby.Ping(row, col)

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if row != *payload.Row() {
			t.Errorf("expected row '%d', got '%d'", row, payload.Row())
		}
		if col != *payload.Column() {
			t.Errorf("expected column '%d', got '%d'", col, payload.Column())
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		_, err := lobby.Ping(9, 0)

		if err != game.ErrOutOfBounds {
			t.Errorf("expected '%v', got '%v'", game.ErrOutOfBounds, err)
		}
	})
}

func TestState(t *testing.T) {
	t.Run("returns current and initial board", func(t *testing.T) {
		lobby := New(uuid.NewString(), false, setupGame(), setupPool(8))
		payload := lobby.State()

		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Current() == nil {
			t.Error("expected current board state in payload")
		}
		if payload.Initial() == nil {
			t.Error("expected initial board state in payload")
		}
	})
}
