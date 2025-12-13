package lobby

import (
	"sudojo/pkg/history"
	"sudojo/pkg/player"
	"testing"
)

func TestCreate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		l := Open(false, 8)
		_, err := l.Create("username123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("full lobby", func(t *testing.T) {
		size := 8
		p := Open(false, size)
		for range size {
			_, err := p.Create("")
			if err != nil {
				t.Errorf("unexpected error '%v' creating player", err)
			}
		}
		_, err := p.Create("")
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
		l := Open(false, 8)
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
		l := Open(false, 8)
		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected error '%v' creating player", err)
		}
		players, err := l.Join(token)
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})

	t.Run("join initial player", func(t *testing.T) {
		l := Open(false, 8)
		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected error '%v' creating player", err)
		}
		l = New(l.id, l.game, l.players, history.New(nil), l.strict, l.size)

		players, err := l.Join(token)
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
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
		l := Open(false, 8)
		token, err := l.Create("")
		if err != nil {
			t.Errorf("unexpected error '%v' creating player", err)
		}
		_, err = l.Join(token)
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
		}
		players, err := l.Leave(token)
		if err != nil {
			t.Errorf("unexpected error '%v' leaving player", err)
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if players[0].Active() {
			t.Error("expected player to be inactive")
		}
	})
}
