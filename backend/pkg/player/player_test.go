package player

import (
	"sudojo/pkg/event"
	"testing"
)

func TestPlayer(t *testing.T) {
	token, name := newToken(), "test-player"
	p := New(token, name)
	if p.Token() != token {
		t.Errorf("want: %s, got: %s", token, p.Token())
	}
	if p.Name() != name {
		t.Errorf("want: %s, got: %s", name, p.Name())
	}
	if p.Active() != false {
		t.Errorf("want: %t, got: %t", false, p.Active())
	}
	p.SetActive(true)
	if p.Active() != true {
		t.Errorf("want: %t, got: %t", true, p.Active())
	}
}

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

func TestPoolCreate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		p := NewPlayerPool(make(map[string]string), 8, event.NewEventBus())
		_, err := p.Create("username123456789")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("full pool", func(t *testing.T) {
		maxSize := 8
		p := NewPlayerPool(make(map[string]string), maxSize, event.NewEventBus())
		for range maxSize {
			_, err := p.Create("")
			if err != nil {
				t.Errorf("unexpected error '%v' creating player", err)
			}
		}
		_, err := p.Create("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrPoolFull {
			t.Errorf("expected '%v', got '%v'", ErrPoolFull, err)
		}
	})
}

func TestPoolJoin(t *testing.T) {
	t.Run("non-existing player", func(t *testing.T) {
		maxSize := 8
		p := NewPlayerPool(make(map[string]string), maxSize, event.NewEventBus())
		token := newToken()
		_, err := p.Join(token, event.NewEventChan(), event.NewEventChan())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrPlayerNotFound {
			t.Errorf("expected '%v', got '%v'", ErrPlayerNotFound, err)
		}
	})

	t.Run("join player", func(t *testing.T) {
		called := false
		register := func(id string, pub, sub event.EventChan) {
			called = true
		}
		bus := event.NewMockEventBus(register, func(id string) {}, func() {}, func() {})
		p := NewPlayerPool(make(map[string]string), 8, bus)
		token, err := p.Create("")
		if err != nil {
			t.Errorf("unexpected error '%v' creating player", err)
		}
		players, err := p.Join(token, event.NewEventChan(), event.NewEventChan())
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
		}
		if !called {
			t.Error("event bus register function not called")
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})

	t.Run("join initial player", func(t *testing.T) {
		called := false
		register := func(id string, pub, sub event.EventChan) {
			called = true
		}
		bus := event.NewMockEventBus(register, func(id string) {}, func() {}, func() {})
		token := newToken()
		init := map[string]string{token: ""}
		p := NewPlayerPool(init, 8, bus)
		players, err := p.Join(token, event.NewEventChan(), event.NewEventChan())
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
		}
		if !called {
			t.Error("event bus register function not called")
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if !players[0].Active() {
			t.Error("expected player to be active")
		}
	})
}

func TestPoolLeave(t *testing.T) {
	t.Run("non-existing player", func(t *testing.T) {
		maxSize := 8
		p := NewPlayerPool(make(map[string]string), maxSize, event.NewEventBus())
		token := newToken()
		_, err := p.Leave(token)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrPlayerNotFound {
			t.Errorf("expected '%v', got '%v'", ErrPlayerNotFound, err)
		}
	})

	t.Run("leave player", func(t *testing.T) {
		called := false
		deregister := func(id string) {
			called = true
		}
		bus := event.NewMockEventBus(func(id string, pub, sub event.EventChan) {}, deregister, func() {}, func() {})
		p := NewPlayerPool(make(map[string]string), 8, bus)
		token, err := p.Create("")
		if err != nil {
			t.Errorf("unexpected error '%v' creating player", err)
		}
		_, err = p.Join(token, event.NewEventChan(), event.NewEventChan())
		if err != nil {
			t.Errorf("unexpected error '%v' joining player", err)
		}
		players, err := p.Leave(token)
		if err != nil {
			t.Errorf("unexpected error '%v' leaving player", err)
		}
		if !called {
			t.Error("event bus deregister function not called")
		}
		if len(players) != 1 {
			t.Fatalf("expected player list length %d, got %d", 1, len(players))
		}
		if players[0].Active() {
			t.Error("expected player to be inactive")
		}
	})
}
