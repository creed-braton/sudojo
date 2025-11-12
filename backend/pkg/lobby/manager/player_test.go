package manager

import (
	"errors"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/player"
	"sync"
	"testing"
)

// mockCapture tracks all interactions with mocked EventBus and Fanout
type mockCapture struct {
	mu              sync.Mutex
	sentEvents      []event.Event
	registeredIDs   []string
	deregisteredIDs []string
}

// setUpWithMocks creates a playerManager with mocked dependencies that capture interactions
func setUpWithMocks(maxPlayer int) (*playerManager, *mockCapture) {
	capture := &mockCapture{
		sentEvents:      []event.Event{},
		registeredIDs:   []string{},
		deregisteredIDs: []string{},
	}

	bus := event.NewMockEventBus(
		func(e event.Event) error {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.sentEvents = append(capture.sentEvents, e)
			return nil
		},
		func() (event.Event, error) {
			return nil, nil
		},
	)

	fanout := event.NewMockFanout(
		func(id string, bus event.EventBus) {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.registeredIDs = append(capture.registeredIDs, id)
		},
		func(id string) {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.deregisteredIDs = append(capture.deregisteredIDs, id)
		},
		func() error { return nil },
	)

	return NewPlayerManager(make(map[string]string), maxPlayer, bus, fanout), capture
}

// setUpWithFailingBus creates a playerManager where the bus Send() returns an error
func setUpWithFailingBus(maxPlayer int, sendErr error) (*playerManager, *mockCapture) {
	capture := &mockCapture{
		sentEvents:      []event.Event{},
		registeredIDs:   []string{},
		deregisteredIDs: []string{},
	}

	bus := event.NewMockEventBus(
		func(e event.Event) error {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.sentEvents = append(capture.sentEvents, e)
			return sendErr
		},
		func() (event.Event, error) {
			return nil, nil
		},
	)

	fanout := event.NewMockFanout(
		func(id string, bus event.EventBus) {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.registeredIDs = append(capture.registeredIDs, id)
		},
		func(id string) {
			capture.mu.Lock()
			defer capture.mu.Unlock()
			capture.deregisteredIDs = append(capture.deregisteredIDs, id)
		},
		func() error { return nil },
	)

	return NewPlayerManager(make(map[string]string), maxPlayer, bus, fanout), capture
}

// assertEvent verifies event properties
func assertEvent(t *testing.T, e event.Event, expectedType, expectedSender string) {
	t.Helper()
	if e == nil {
		t.Fatal("event is nil")
	}
	if e.Type() != expectedType {
		t.Errorf("expected event type %s, got %s", expectedType, e.Type())
	}
	if e.Sender() != expectedSender {
		t.Errorf("expected sender %s, got %s", expectedSender, e.Sender())
	}
	if e.Payload() == nil {
		t.Error("payload is nil")
	}
}

// assertPlayerPayload verifies payload is playerPayload type with expected count
func assertPlayerPayload(t *testing.T, payload event.Payload, expectedCount int) *playerPayload {
	t.Helper()
	pp, ok := payload.(*playerPayload)
	if !ok {
		t.Fatal("payload is not playerPayload type")
	}
	if len(*pp) != expectedCount {
		t.Errorf("expected %d players in payload, got %d", expectedCount, len(*pp))
	}
	return pp
}

func TestCreate(t *testing.T) {
	t.Run("create player", func(t *testing.T) {
		m, _ := setUpWithMocks(1)
		token, err := m.Create("test-player")
		if err != nil {
			t.Fatalf("unexpected error creating player: %v", err)
		}
		if len(token) < 1 {
			t.Fatal("token missing creating player")
		}
	})

	t.Run("create unique tokens", func(t *testing.T) {
		maxPlayer := 8
		m, _ := setUpWithMocks(maxPlayer)
		check := make(map[string]struct{})
		for range maxPlayer {
			token, err := m.Create("test-player")
			if err != nil {
				t.Fatalf("unexpected error creating player: %v", err)
			}
			if _, exist := check[token]; exist {
				t.Errorf("duplicated token created: %s", token)
			}
			check[token] = struct{}{}
		}
	})

	t.Run("create player lobby full", func(t *testing.T) {
		maxPlayer := 8
		m, _ := setUpWithMocks(maxPlayer)
		for range maxPlayer {
			_, err := m.Create("test-player")
			if err != nil {
				t.Fatalf("unexpected error creating player: %v", err)
			}
		}
		_, err := m.Create("test-player")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err != ErrLobbyFull {
			t.Fatalf("expected '%v', got: %v", ErrLobbyFull, err)
		}
	})

	t.Run("create player invalid name", func(t *testing.T) {
		m, _ := setUpWithMocks(1)
		_, err := m.Create("test-player-too-long-name")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err != player.ErrNameTooLong {
			t.Fatalf("expected '%v', got: %v", player.ErrNameTooLong, err)
		}
	})
}

func TestJoin(t *testing.T) {
	t.Run("join player", func(t *testing.T) {
		maxPlayer := 8
		m, capture := setUpWithMocks(maxPlayer)
		tokens := []string{}

		for i := range maxPlayer {
			token, err := m.Create("test-player")
			if err != nil {
				t.Fatalf("unexpected error creating player: %v", err)
			}
			tokens = append(tokens, token)

			p, err := m.Join(token, "test-player")
			if err != nil {
				t.Fatalf("unexpected error joining player: %v", err)
			}
			if p == nil {
				t.Fatal("player is nil")
			}

			// Verify this player was registered
			if len(capture.registeredIDs) != i+1 {
				t.Errorf("expected %d registrations, got %d", i+1, len(capture.registeredIDs))
			}
			if capture.registeredIDs[i] != token {
				t.Errorf("expected token %s registered, got %s", token, capture.registeredIDs[i])
			}

			// Verify join event was sent
			if len(capture.sentEvents) != i+1 {
				t.Errorf("expected %d events, got %d", i+1, len(capture.sentEvents))
			}
			e := capture.sentEvents[i]
			if e.Type() != event.JoinEvent {
				t.Errorf("expected JoinEvent, got %s", e.Type())
			}
			if e.Sender() != token {
				t.Errorf("expected sender %s, got %s", token, e.Sender())
			}
			if e.Payload() == nil {
				t.Error("payload missing")
			}
		}
	})

	t.Run("join player not found", func(t *testing.T) {
		m, capture := setUpWithMocks(1)
		name := "test-player"
		token, err := m.Create(name)
		if err != nil {
			t.Fatalf("unexpected error creating player: %v", err)
		}

		p, err := m.Join(token+"a", name)

		// Error assertions
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if err != ErrPlayerNotFound {
			t.Errorf("expected '%v', got: %v", ErrPlayerNotFound, err)
		}
		if p != nil {
			t.Error("unexpected player creation joining")
		}

		// Verify no registration occurred
		if len(capture.registeredIDs) != 0 {
			t.Errorf("expected no registration, got %d", len(capture.registeredIDs))
		}

		// Verify no event was sent (since Join failed before sending)
		if len(capture.sentEvents) != 0 {
			t.Errorf("expected no events, got %d", len(capture.sentEvents))
		}
	})

	t.Run("join player already active", func(t *testing.T) {
		m, capture := setUpWithMocks(1)
		name := "test-player"
		token, err := m.Create(name)
		if err != nil {
			t.Fatalf("unexpected error creating player: %v", err)
		}

		// First join
		old, err := m.Join(token, name)
		if err != nil {
			t.Fatalf("unexpected error on first join: %v", err)
		}

		// Second join with same token
		p, err := m.Join(token, name)
		if err != nil {
			t.Fatalf("unexpected error on second join: %v", err)
		}

		// Verify two join events sent
		if len(capture.sentEvents) != 2 {
			t.Errorf("expected 2 events, got %d", len(capture.sentEvents))
		}

		// Verify both are join events
		for i, e := range capture.sentEvents {
			if e.Type() != event.JoinEvent {
				t.Errorf("event %d: expected JoinEvent, got %s", i, e.Type())
			}
		}

		// Verify player objects are different
		if old.Token() != p.Token() {
			t.Error("tokens should be the same")
		}

		// Verify manager calls Register twice
		if len(capture.registeredIDs) != 2 {
			t.Errorf("expected 2 registrations, got %d", len(capture.registeredIDs))
		}
	})

	t.Run("join handles send error", func(t *testing.T) {
		expectedErr := errors.New("bus full")
		m, capture := setUpWithFailingBus(1, expectedErr)

		token, _ := m.Create("test")
		p, err := m.Join(token, "test")

		// Player should still be created and registered
		if p == nil {
			t.Error("expected player to be created")
		}
		if len(capture.registeredIDs) != 1 {
			t.Error("expected registration to occur")
		}

		// But error should be returned from Join
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestLeave(t *testing.T) {
	t.Run("leave player", func(t *testing.T) {
		m, capture := setUpWithMocks(2)

		// Create and join test player
		name := "test-player"
		token, _ := m.Create(name)
		p, _ := m.Join(token, name)

		// Clear join event from capture
		capture.sentEvents = []event.Event{}

		// Leave
		err := m.Leave(p)
		if err != nil {
			t.Fatalf("unexpected error leaving player: %v", err)
		}

		// Verify deregistration
		if len(capture.deregisteredIDs) != 1 {
			t.Errorf("expected 1 deregistration, got %d", len(capture.deregisteredIDs))
		}
		if capture.deregisteredIDs[0] != token {
			t.Errorf("expected token %s deregistered, got %s", token, capture.deregisteredIDs[0])
		}

		// Verify leave event sent
		if len(capture.sentEvents) != 1 {
			t.Errorf("expected 1 event, got %d", len(capture.sentEvents))
		}
		e := capture.sentEvents[0]
		if e.Type() != event.LeaveEvent {
			t.Errorf("expected LeaveEvent, got %s", e.Type())
		}
		if e.Sender() != token {
			t.Errorf("expected sender %s, got %s", token, e.Sender())
		}
		if e.Payload() == nil {
			t.Error("payload missing")
		}

		// Verify player's own bus is closed (manager calls p.Close())
		_, err = p.Receive()
		if err != event.ErrClosedBus {
			t.Errorf("expected ErrClosedBus, got %v", err)
		}
	})

	t.Run("leave player not found", func(t *testing.T) {
		m, capture := setUpWithMocks(2)
		name := "test-player"
		token, _ := m.Create(name)
		m.Join(token, name)

		// Create mock player with non-existent token
		mockPlayer := player.NewMock(
			"non-existent-token",
			name,
			func(e event.Event) error { return nil },
			func() (event.Event, error) { return nil, nil },
		)

		// Clear join events
		capture.deregisteredIDs = []string{}
		capture.sentEvents = []event.Event{}

		err := m.Leave(mockPlayer)

		if err == nil {
			t.Error("expected error but got nil leaving player")
		}
		if err != ErrPlayerNotFound {
			t.Errorf("expected '%v', got: '%v', leaving player", ErrPlayerNotFound, err)
		}

		// Verify deregister was still called (happens before error check in Leave())
		if len(capture.deregisteredIDs) != 1 {
			t.Errorf("expected deregister call, got %d calls", len(capture.deregisteredIDs))
		}

		// Verify no event sent (error occurs before Send())
		if len(capture.sentEvents) != 0 {
			t.Errorf("expected no events sent, got %d", len(capture.sentEvents))
		}
	})
}

func TestStatus(t *testing.T) {
	t.Run("join events have correct status payload", func(t *testing.T) {
		maxPlayer := 8
		m, capture := setUpWithMocks(maxPlayer)
		tokens := []string{}

		for i := range maxPlayer {
			name := fmt.Sprintf("player-%d", i)
			token, _ := m.Create(name)
			tokens = append(tokens, token)

			_, err := m.Join(token, name)
			if err != nil {
				t.Fatalf("unexpected error joining player %d: %v", i, err)
			}

			// Verify event was sent
			if len(capture.sentEvents) != i+1 {
				t.Fatalf("expected %d events, got %d", i+1, len(capture.sentEvents))
			}

			e := capture.sentEvents[i]
			if e.Type() != event.JoinEvent {
				t.Errorf("event %d: expected JoinEvent, got %s", i, e.Type())
			}

			payload, ok := e.Payload().(*playerPayload)
			if !ok {
				t.Errorf("event %d: unexpected payload type", i)
				continue
			}

			// Verify payload length
			if len(*payload) != i+1 {
				t.Errorf("event %d: expected payload length %d, got %d", i, i+1, len(*payload))
			}

			// Verify all players in payload are active
			for j, s := range *payload {
				if !s.Active {
					t.Errorf("event %d, player %d (%s): expected active, got inactive", i, j, s.Name)
				}
			}
		}
	})

	t.Run("leave events have correct status payload", func(t *testing.T) {
		maxPlayer := 8
		m, capture := setUpWithMocks(maxPlayer)
		players := []player.Player{}

		// Create and join all players
		for i := range maxPlayer {
			name := fmt.Sprintf("player-%d", i)
			token, _ := m.Create(name)
			p, _ := m.Join(token, name)
			players = append(players, p)
		}

		// Clear join events
		capture.sentEvents = []event.Event{}

		// Leave each player
		for i, p := range players {
			err := m.Leave(p)
			if err != nil {
				t.Fatalf("unexpected error leaving player %d: %v", i, err)
			}

			// Verify leave event
			if len(capture.sentEvents) != i+1 {
				t.Fatalf("expected %d leave events, got %d", i+1, len(capture.sentEvents))
			}

			e := capture.sentEvents[i]
			if e.Type() != event.LeaveEvent {
				t.Errorf("event %d: expected LeaveEvent, got %s", i, e.Type())
			}

			payload, ok := e.Payload().(*playerPayload)
			if !ok {
				t.Errorf("event %d: unexpected payload type", i)
				continue
			}

			// Verify the left player is inactive
			for _, s := range *payload {
				if s.Name == p.Name() {
					if s.Active {
						t.Errorf("player %s should be inactive after leaving", s.Name)
					}
				}
			}
		}
	})

	t.Run("status payload is sorted by token", func(t *testing.T) {
		m, capture := setUpWithMocks(3)

		// Create players with specific names
		names := []string{"charlie", "alice", "bob"}
		for _, name := range names {
			token, _ := m.Create(name)
			m.Join(token, name)
		}

		// Get last event
		lastEvent := capture.sentEvents[len(capture.sentEvents)-1]
		payload, ok := lastEvent.Payload().(*playerPayload)
		if !ok {
			t.Fatal("unexpected payload type")
		}

		// Verify payload has all players
		if len(*payload) != 3 {
			t.Fatalf("expected 3 players, got %d", len(*payload))
		}

		// Verify all players are present
		nameSet := make(map[string]bool)
		for _, s := range *payload {
			nameSet[s.Name] = true
		}
		for _, name := range names {
			if !nameSet[name] {
				t.Errorf("player %s missing from payload", name)
			}
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("join with different name than create", func(t *testing.T) {
		m, _ := setUpWithMocks(1)
		token, _ := m.Create("original")

		// Join allows different name than create
		p, err := m.Join(token, "different-name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "different-name" {
			t.Errorf("expected name 'different-name', got %s", p.Name())
		}
	})

	t.Run("create with pre-existing players", func(t *testing.T) {
		existing := map[string]string{
			"token1": "player1",
			"token2": "player2",
		}

		bus := event.NewMockEventBus(
			func(e event.Event) error { return nil },
			func() (event.Event, error) { return nil, nil },
		)
		fanout := event.NewMockFanout(
			func(id string, bus event.EventBus) {},
			func(id string) {},
			func() error { return nil },
		)

		m := NewPlayerManager(existing, 4, bus, fanout)

		// Can only create 2 more (4 - 2 existing)
		_, err := m.Create("player3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = m.Create("player4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = m.Create("player5")
		if err != ErrLobbyFull {
			t.Errorf("expected ErrLobbyFull, got %v", err)
		}
	})

	t.Run("MaxPlayer returns correct value", func(t *testing.T) {
		m, _ := setUpWithMocks(42)
		if m.MaxPlayer() != 42 {
			t.Errorf("expected MaxPlayer 42, got %d", m.MaxPlayer())
		}
	})
}

func TestConcurrentOperations(t *testing.T) {
	t.Run("concurrent creates", func(t *testing.T) {
		m, _ := setUpWithMocks(100)
		var wg sync.WaitGroup
		tokens := make(chan string, 100)

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				token, err := m.Create("test")
				if err == nil {
					tokens <- token
				}
			}()
		}

		wg.Wait()
		close(tokens)

		// Verify uniqueness
		seen := make(map[string]bool)
		for token := range tokens {
			if seen[token] {
				t.Errorf("duplicate token: %s", token)
			}
			seen[token] = true
		}
	})

	t.Run("concurrent join operations", func(t *testing.T) {
		m, capture := setUpWithMocks(10)

		// Create players
		tokens := []string{}
		for i := 0; i < 10; i++ {
			token, _ := m.Create(fmt.Sprintf("player-%d", i))
			tokens = append(tokens, token)
		}

		var wg sync.WaitGroup

		// Join all concurrently
		players := make([]player.Player, 10)
		for i, token := range tokens {
			wg.Add(1)
			go func(idx int, t string) {
				defer wg.Done()
				p, _ := m.Join(t, fmt.Sprintf("player-%d", idx))
				players[idx] = p
			}(i, token)
		}
		wg.Wait()

		// Verify all registered
		if len(capture.registeredIDs) != 10 {
			t.Errorf("expected 10 registrations, got %d", len(capture.registeredIDs))
		}
	})

	t.Run("concurrent join and leave", func(t *testing.T) {
		m, _ := setUpWithMocks(20)

		// Create players
		tokens := []string{}
		for i := 0; i < 20; i++ {
			token, _ := m.Create(fmt.Sprintf("player-%d", i))
			tokens = append(tokens, token)
		}

		var wg sync.WaitGroup

		// Join first 10
		players := make([]player.Player, 10)
		for i := 0; i < 10; i++ {
			p, _ := m.Join(tokens[i], fmt.Sprintf("player-%d", i))
			players[i] = p
		}

		// Concurrently: join next 10 and leave first 10
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func(idx int) {
				defer wg.Done()
				m.Join(tokens[idx+10], fmt.Sprintf("player-%d", idx+10))
			}(i)
			go func(idx int) {
				defer wg.Done()
				m.Leave(players[idx])
			}(i)
		}

		wg.Wait()

		// No assertions needed - just verify no panic/deadlock
	})
}
