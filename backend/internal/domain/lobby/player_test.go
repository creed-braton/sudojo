package lobby

import (
	"fmt"
	"testing"
)

func TestValidName(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  error
	}{
		{name: "empty name", input: "", want: nil},
		{name: "just too long name", input: "username123456789", want: ErrNameTooLong},
		{name: "almost too long name", input: "username12345678", want: nil},
		{name: "illegal character", input: "user+name", want: ErrNameChar},
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
	t.Run("unique player tokens", func(t *testing.T) {
		maxPlayer := 8192
		m := newPlayerManager(maxPlayer)
		for range maxPlayer {
			_, err := m.create("player")
			if err != nil {
				t.Fatal(err)
			}
		}

		tokens := make(map[string]bool)
		for token := range m.Players {
			if tokens[token] {
				t.Errorf("duplicate token found: %s", token)
			}
			tokens[token] = true
		}
	})

	t.Run("too many players", func(t *testing.T) {
		maxPlayer := 16
		m := newPlayerManager(maxPlayer)
		for range maxPlayer {
			_, err := m.create("in-size")
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err := m.create("out-of-size")
		if err == nil {
			t.Error("wanted error, got nil")
			return
		}
		if err != ErrLobbyFull {
			t.Errorf("want: %v, got: %v", ErrLobbyFull, err)
		}
	})

	t.Run("player retrieval", func(t *testing.T) {
		maxPlayer := 16
		m := newPlayerManager(maxPlayer)

		players := make([]string, 0, maxPlayer)

		for i := range maxPlayer {
			p, err := m.create(fmt.Sprintf("player%d", i))
			if err != nil {
				t.Fatal(err)
			}
			players = append(players, p)
		}

		for _, token := range players {
			retrieved, err := m.player(token)
			if err != nil {
				t.Fatal(err)
			}
			if retrieved == nil {
				t.Errorf("want player %s, got nil", token)
				continue
			}
			if retrieved.Token != token {
				t.Errorf("want token %s, got %s", token, retrieved.Token)
			}
		}

		invalid, err := m.player("invalid-token")
		if err == nil {
			t.Error("wanted error, got nil")
		}
		if invalid != nil {
			t.Errorf("wanted nil for invalid token, got player %s", invalid.Token)
		}
	})
}

func TestJoin(t *testing.T) {
	t.Run("existing player", func(t *testing.T) {
		m := newPlayerManager(4)
		token, err := m.create("Jet Gibbity")
		if err != nil {
			t.Fatal(err)
		}

		p, _ := m.player(token)
		if p.Active {
			t.Error("wanted player to be inactive initially")
		}

		payload, err := m.join(token)
		if err != nil {
			t.Fatalf("error joining: %v", err)
		}
		if payload == nil {
			t.Fatal("wanted payload got nil after join")
		}

		p, _ = m.player(token)
		if !p.Active {
			t.Error("wanted player to be active after join")
		}

		found := false
		for _, p := range payload.Players {
			if p.Token == token {
				found = true
				if !p.Active {
					t.Error("wanted joined player to be active in payload")
				}
				break
			}
		}
		if !found {
			t.Errorf("player %s not found in payload", token)
		}
	})

	t.Run("non existing player", func(t *testing.T) {
		m := newPlayerManager(4)
		_, err := m.join("nonexistent")
		if err == nil {
			t.Fatal("wanted error for joining with invalid token, got nil")
		}
		if err != ErrPlayerMiss {
			t.Errorf("want: %v, got: %v", ErrPlayerMiss, err)
		}
	})
}

func TestLeave(t *testing.T) {
	t.Run("existing player", func(t *testing.T) {
		m := newPlayerManager(4)
		token, err := m.create("Charlie")
		if err != nil {
			t.Fatal(err)
		}

		_, _ = m.join(token)

		payload, err := m.leave(token)
		if err != nil {
			t.Fatalf("error leaving: %v", err)
		}
		if payload == nil {
			t.Fatal("wanted payload got nil after leave")
		}

		player, _ := m.player(token)
		if player.Active {
			t.Error("wanted player to be inactive after leave")
		}

		found := false
		for _, p := range payload.Players {
			if p.Token == token {
				found = true
				if p.Active {
					t.Error("wanted player to be inactive in payload")
				}
				break
			}
		}
		if !found {
			t.Errorf("player %s not found in payload", token)
		}
	})

	t.Run("leave without joining", func(t *testing.T) {
		m := newPlayerManager(4)
		token, _ := m.create("Nestor")

		payload, err := m.leave(token)
		if err != nil {
			t.Fatalf("wanted error leaving: %v", err)
		}

		player, _ := m.player(token)
		if player.Active {
			t.Error("wanted player to remain inactive")
		}

		if payload == nil || len(payload.Players) != 1 {
			t.Error("wanted payload to include the player")
		}
	})

	t.Run("leave with invalid token", func(t *testing.T) {
		m := newPlayerManager(4)
		_, err := m.leave("invalid-token")
		if err == nil {
			t.Fatal("wante error for invalid token, got nil")
		}
		if err != ErrPlayerMiss {
			t.Errorf("want: %v, got: %v", ErrPlayerMiss, err)
		}
	})
}
