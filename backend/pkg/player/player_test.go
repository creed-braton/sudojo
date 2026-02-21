package player

import "testing"

func TestPlayer(t *testing.T) {
	token, name := "test-token", "test-name"
	p := New(token, name)

	if p == nil {
		t.Fatal("expected player, got nil")
	}
	if p.Token() != token {
		t.Errorf("expected token '%s', got '%s'", token, p.Token())
	}
	if p.Name() != name {
		t.Errorf("expected name '%s', got '%s'", name, p.Name())
	}
	if p.Active() {
		t.Error("expected active to be false, got true")
	}
}

func TestTokenUniqueness(t *testing.T) {
	num := 10000
	check := make(map[string]struct{}, num)

	for range num {
		token := NewToken()
		if _, exist := check[token]; exist {
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
			got := ValidName(test.input)
			if test.want != got {
				t.Errorf("expected error: '%v', got: '%v'", test.want, got)
			}
		})
	}
}

func TestSort(t *testing.T) {
	players := map[string]Player{
		"charlie": New("charlie", ""),
		"alice":   New("alice", ""),
		"bob":     New("bob", ""),
		"bard":    New("bard", ""),
	}

	sorted := Sort(players)

	if len(sorted) != 4 {
		t.Fatalf("expected 4 players, got %d", len(sorted))
	}
	if sorted[0].Token() != "alice" {
		t.Errorf("expected first player token 'alice', got '%s'", sorted[0].Token())
	}
	if sorted[1].Token() != "bard" {
		t.Errorf("expected second player token 'bard', got '%s'", sorted[1].Token())
	}
	if sorted[2].Token() != "bob" {
		t.Errorf("expected third player token 'bob', got '%s'", sorted[2].Token())
	}
	if sorted[3].Token() != "charlie" {
		t.Errorf("expected fourth player token 'charlie', got '%s'", sorted[3].Token())
	}
}
