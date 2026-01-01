package manager

import (
	"sudojo/pkg/event"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"testing"
)

func setup(strict bool) ([]string, []Player, func(reason int)) {
	size := 8
	lobby := lobby.Open(strict, size)
	mng := New(lobby)
	tokens := []string{}
	players := []Player{}
	go func() {
		for {
			if err := mng.Pump(); err != nil {
				return
			}
		}
	}()
	for range size {
		t, _ := mng.Create("")
		p, _ := mng.Join(t)
		tokens = append(tokens, t)
		players = append(players, p)
		for _, p := range players {
			p.Receive() // drop join events
		}
	}
	return tokens, players, mng.Close
}

func TestPlayerLeave(t *testing.T) {
	want := "test-token"
	lobby := lobby.Open(false, 8)
	broadcast := func(e event.Event) error {
		return nil
	}
	called := false
	leave := func(token string) {
		called = true
		if want != token {
			t.Errorf("expected leave call with token '%s' got '%s'", want, token)
		}
	}
	player := NewPlayer(want, lobby, event.NewBuffer(), broadcast, leave, func() {})
	player.Leave()
	if !called {
		t.Fatal("expected controller leave function to be called")
	}
	if _, err := player.Receive(); err == nil {
		t.Errorf("expected error receiving event got nil")
	}
}

func TestPlayerPing(t *testing.T) {
	t.Run("ping out of bounds", func(t *testing.T) {
		_, players, teardown := setup(false)
		defer teardown(event.BufferReason)
		if err := players[0].Ping(9, 9, ""); err != nil {
			t.Fatalf("unexpected error pinging '%v'", err)
		}
		e, err := players[0].Receive()
		if err != nil {
			t.Fatalf("unexpected error receiving event '%v'", err)
		}
		if e.Error() != game.ErrOutOfBounds.Error() {
			t.Errorf("expeected error '%s' got '%s'", game.ErrOutOfBounds.Error(), e.Error())
		}
	})

	t.Run("ping distribution", func(t *testing.T) {
		tokens, players, teardown := setup(false)
		defer teardown(event.BufferReason)
		for i, p := range players {
			if err := p.Ping(0, 0, tokens[i]); err != nil {
				t.Fatalf("unexpected error pinging player '%d': '%v'", i, err)
			}
			for j, p := range players {
				e, err := p.Receive()
				if err != nil {
					t.Errorf("unexpected error receiving event player '%d': '%v'", j, err)
					continue
				}
				if e.Type() != event.PingEvent {
					t.Errorf("expected event type '%s' got '%s'", event.PingEvent, e.Type())
				}
				if e.Sender() != tokens[i] {
					t.Errorf("expected sender '%s' got '%s'", tokens[i], e.Sender())
				}
				if e.Trace() != tokens[i] {
					t.Errorf("expected trace '%s' got '%s'", tokens[i], e.Trace())
				}
				if e.Payload() == nil {
					t.Error("payload missing")
					continue
				}
				if e.Payload().Row() == nil {
					t.Error("row in payload missing")
					continue
				}
				if *e.Payload().Row() != 0 {
					t.Errorf("expected row '%d' got '%d'", *e.Payload().Row(), 0)
				}
				if e.Payload().Column() == nil {
					t.Error("column in payload missing")
					continue
				}
				if *e.Payload().Column() != 0 {
					t.Errorf("expected column '%d' got '%d'", *e.Payload().Column(), 0)
				}
			}
		}
	})
}

func TestPlayerInsert(t *testing.T) {
	t.Run("strict insert", func(t *testing.T) {
	})

	t.Run("incorrect strict insert", func(t *testing.T) {
	})

	t.Run("invalid strict insert", func(t *testing.T) {
	})

	t.Run("lax insert", func(t *testing.T) {
	})

	t.Run("conflicting lax insert", func(t *testing.T) {
	})

	t.Run("invalid lax insert", func(t *testing.T) {
	})
}

func TestPlayerState(t *testing.T) {

}
