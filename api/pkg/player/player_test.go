package player

import "testing"

func TestPlayer(t *testing.T) {
	token, name := "test-token", "test-name"
	p := New(token, name)

	if p == nil {
		t.Fatal("unexpected nil player")
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
				t.Errorf("want: %v, got: %v", test.want, got)
			}
		})
	}
}
