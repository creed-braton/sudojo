package player

import (
	"sudojo/pkg/event"
	"testing"
)

func TestPlayer(t *testing.T) {
	token, name, bus := NewToken(), "test-player", event.NewEventBus()
	p := New(token, name, bus)
	if err := p.Send(event.New("", "", "", "", nil)); err != nil {
		t.Errorf("unexpected error sending event: %v", err)
	}

	if p.Token() != token {
		t.Errorf("want: %s, got: %s", token, p.Token())
	}
	if p.Name() != name {
		t.Errorf("want: %s, got: %s", name, p.Name())
	}
	e, err := p.Receive()
	if err != nil {
		t.Errorf("unexpected error polling event: %v", err)
	}
	if e == nil {
		t.Error("received nil instead of event")
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
