package lobby

import "testing"

func TestTokenUniqueness(t *testing.T) {
	num := 10000
	check := make(map[string]struct{}, num)

	for range num {
		token := newToken()
		if _, exists := check[token]; exists {
			t.Fatalf("duplicate token found: %s", token)
		}
		check[token] = struct{}{}
	}
}

func TestValidName(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  error
	}{
		{name: "empty name", input: "", want: nil},
		{name: "just too long name", input: "username123456789", want: ErrNameTooLong},
		{name: "almost too long name", input: "username12345678", want: nil},
		{name: "illegal character", input: "user+name", want: ErrInvalidChar},
		{name: "legal special character", input: "user-name", want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validName(test.input)
			if test.want != got {
				t.Errorf("want: %v, got: %v", test.want, got)
			}
		})
	}
}

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
		token := newToken()
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
		l = New(l.id, l.game, l.players, l.strict, l.size)

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
		players := l.Leave(token)
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
