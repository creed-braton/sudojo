package lobby

import "testing"

func TestTokenUniqueness(t *testing.T) {
	num := 10000
	check := make(map[string]struct{}, num)

	for range num {
		token := newToken()
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
			got := validName(test.input)
			if test.want != got {
				t.Errorf("want: %v, got: %v", test.want, got)
			}
		})
	}
}
