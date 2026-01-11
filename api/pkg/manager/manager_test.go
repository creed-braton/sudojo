package manager

import (
	"math/rand"
	"sudojo/pkg/event"
	"sudojo/pkg/message"
	"sudojo/pkg/sudoku"
	"sync"
	"testing"
)

func ptr(v int) *int {
	return &v
}

func TestTakeover(t *testing.T) {
	t.Run("old buffer closes with takeover reason and new buffer receives events", func(t *testing.T) {
		m := NewMock(true, true, true, 2)

		// create player
		token, err := m.Lobby().Create("alice")
		if err != nil {
			t.Fatalf("unexpected create error: '%v'", err)
		}

		// first join
		oldBuffer, err := m.Join(token)
		if err != nil {
			t.Fatalf("unexpected join error: '%v'", err)
		}

		// consume join event from old buffer
		evt, err := oldBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.JoinEvent {
			t.Errorf("expected event type '%s', got '%s'", event.JoinEvent, evt.Type())
		}

		// second join with same token (takeover)
		newBuffer, err := m.Join(token)
		if err != nil {
			t.Fatalf("unexpected takeover join error: '%v'", err)
		}

		// verify old buffer is closed with TakeoverReason
		_, err = oldBuffer.Receive()
		if err == nil {
			t.Fatal("expected error from old buffer, got nil")
		}
		closedErr, ok := err.(*event.BufferClosedError)
		if !ok {
			t.Fatalf("expected BufferClosedError, got %T", err)
		}
		if closedErr.Reason() != event.TakeoverReason {
			t.Errorf("expected TakeoverReason (%d), got %d", event.TakeoverReason, closedErr.Reason())
		}

		// verify old buffer cannot send
		sendErr := oldBuffer.Send(event.New(event.PingEvent, 0, ""))
		if sendErr == nil {
			t.Error("expected error when sending to old buffer, got nil")
		}

		// consume join event from new buffer
		evt, err = newBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error from new buffer: '%v'", err)
		}
		if evt.Type() != event.JoinEvent {
			t.Errorf("expected event type '%s', got '%s'", event.JoinEvent, evt.Type())
		}

		// send ping and verify only new buffer receives it
		pingMsg := message.New(message.PingMsg, ptr(4), ptr(4), nil, "ping-trace")
		pingMsg.SetSender(token)
		m.Process(pingMsg)

		// new buffer should receive ping
		evt, err = newBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error from new buffer: '%v'", err)
		}
		if evt.Type() != event.PingEvent {
			t.Errorf("expected event type '%s', got '%s'", event.PingEvent, evt.Type())
		}
		if evt.Trace() != "ping-trace" {
			t.Errorf("expected trace 'ping-trace', got '%s'", evt.Trace())
		}
	})

	t.Run("broadcast only reaches new buffer after takeover", func(t *testing.T) {
		m := NewMock(true, true, true, 2)

		// create two players
		token1, err := m.Lobby().Create("alice")
		if err != nil {
			t.Fatalf("unexpected create error: '%v'", err)
		}
		token2, err := m.Lobby().Create("bob")
		if err != nil {
			t.Fatalf("unexpected create error: '%v'", err)
		}

		// first player joins
		oldBuffer, err := m.Join(token1)
		if err != nil {
			t.Fatalf("unexpected join error: '%v'", err)
		}
		// consume join event
		_, _ = oldBuffer.Receive()

		// second player joins
		buffer2, err := m.Join(token2)
		if err != nil {
			t.Fatalf("unexpected join error: '%v'", err)
		}
		// consume join event from both buffers
		_, _ = oldBuffer.Receive()
		_, _ = buffer2.Receive()

		// first player rejoins (takeover)
		newBuffer, err := m.Join(token1)
		if err != nil {
			t.Fatalf("unexpected takeover join error: '%v'", err)
		}
		// consume rejoin event from second player and new buffer
		_, _ = buffer2.Receive()
		_, _ = newBuffer.Receive()

		// second player sends an insert which broadcasts to all
		insertMsg := message.New(message.InsertMsg, ptr(0), ptr(1), ptr(m.Lobby().Game().Solution().Cell(0, 1)), "insert-trace")
		insertMsg.SetSender(token2)
		m.Process(insertMsg)

		// new buffer should receive broadcast
		evt, err := newBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error from new buffer: '%v'", err)
		}
		if evt.Type() != event.InsertEvent {
			t.Errorf("expected event type '%s', got '%s'", event.InsertEvent, evt.Type())
		}

		// second player should also receive broadcast
		evt, err = buffer2.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error from second buffer: '%v'", err)
		}
		if evt.Type() != event.InsertEvent {
			t.Errorf("expected event type '%s', got '%s'", event.InsertEvent, evt.Type())
		}

		// old buffer should still be closed (no events)
		_, err = oldBuffer.Receive()
		if err == nil {
			t.Error("expected error from old buffer, got nil")
		}
	})
}

func TestConcurrentManager(t *testing.T) {
	t.Run("multiple players perform random actions concurrently", func(t *testing.T) {
		const (
			playerCount      = 8
			actionsPerPlayer = 50
		)

		m := NewMock(true, false, true, playerCount)

		// create all players
		tokens := make([]string, playerCount)
		for i := range playerCount {
			token, err := m.Lobby().Create("player" + string(rune('A'+i)))
			if err != nil {
				t.Fatalf("unexpected create error for player %d: '%v'", i, err)
			}
			tokens[i] = token
		}

		var wg sync.WaitGroup
		wg.Add(playerCount)

		for i := range playerCount {
			go func(playerIdx int) {
				defer wg.Done()

				token := tokens[playerIdx]
				joined := false

				for range actionsPerPlayer {
					action := rand.Intn(5)

					switch action {
					case 0: // join
						if !joined {
							_, err := m.Join(token)
							if err == nil {
								joined = true
							}
						}
					case 1: // leave
						if joined {
							m.Leave(token)
							joined = false
						}
					case 2: // ping
						if joined {
							row, col := rand.Intn(9), rand.Intn(9)
							msg := message.New(message.PingMsg, ptr(row), ptr(col), nil, "")
							msg.SetSender(token)
							m.Process(msg)
						}
					case 3: // insert
						if joined {
							row, col, val := rand.Intn(9), rand.Intn(9), rand.Intn(9)+1
							msg := message.New(message.InsertMsg, ptr(row), ptr(col), ptr(val), "")
							msg.SetSender(token)
							m.Process(msg)
						}
					case 4: // state
						if joined {
							msg := message.New(message.StateMsg, nil, nil, nil, "")
							msg.SetSender(token)
							m.Process(msg)
						}
					}
				}

				// cleanup: leave if still joined
				if joined {
					m.Leave(token)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent joins and takeovers", func(t *testing.T) {
		const (
			playerCount        = 8
			takeoversPerPlayer = 20
		)

		m := NewMock(true, false, true, playerCount)

		// create all players
		tokens := make([]string, playerCount)
		for i := range playerCount {
			token, err := m.Lobby().Create("player" + string(rune('A'+i)))
			if err != nil {
				t.Fatalf("unexpected create error for player %d: '%v'", i, err)
			}
			tokens[i] = token
		}

		var wg sync.WaitGroup
		wg.Add(playerCount)

		for i := range playerCount {
			go func(playerIdx int) {
				defer wg.Done()

				token := tokens[playerIdx]

				for range takeoversPerPlayer {
					// join (potentially taking over previous buffer)
					_, err := m.Join(token)
					if err != nil {
						continue
					}

					// perform some actions
					for range rand.Intn(5) + 1 {
						action := rand.Intn(3)
						switch action {
						case 0: // ping
							row, col := rand.Intn(9), rand.Intn(9)
							msg := message.New(message.PingMsg, ptr(row), ptr(col), nil, "")
							msg.SetSender(token)
							m.Process(msg)
						case 1: // insert
							row, col, val := rand.Intn(9), rand.Intn(9), rand.Intn(9)+1
							msg := message.New(message.InsertMsg, ptr(row), ptr(col), ptr(val), "")
							msg.SetSender(token)
							m.Process(msg)
						case 2: // state
							msg := message.New(message.StateMsg, nil, nil, nil, "")
							msg.SetSender(token)
							m.Process(msg)
						}
					}
				}

				// final leave
				m.Leave(token)
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent broadcasts during rapid joins", func(t *testing.T) {
		const playerCount = 8

		m := NewMock(true, false, true, playerCount)

		// create all players
		tokens := make([]string, playerCount)
		for i := range playerCount {
			token, err := m.Lobby().Create("player" + string(rune('A'+i)))
			if err != nil {
				t.Fatalf("unexpected create error for player %d: '%v'", i, err)
			}
			tokens[i] = token
		}

		var wg sync.WaitGroup
		wg.Add(playerCount)

		// half the players join and send broadcasts
		for i := range playerCount / 2 {
			go func(playerIdx int) {
				defer wg.Done()

				token := tokens[playerIdx]
				_, err := m.Join(token)
				if err != nil {
					return
				}

				// send multiple inserts (broadcasts)
				for range 10 {
					row, col, val := rand.Intn(9), rand.Intn(9), rand.Intn(9)+1
					msg := message.New(message.InsertMsg, ptr(row), ptr(col), ptr(val), "")
					msg.SetSender(token)
					m.Process(msg)
				}

				m.Leave(token)
			}(i)
		}

		// other half rapidly joins, performs action, and leaves
		for i := playerCount / 2; i < playerCount; i++ {
			go func(playerIdx int) {
				defer wg.Done()

				token := tokens[playerIdx]

				for range 20 {
					_, err := m.Join(token)
					if err != nil {
						continue
					}

					// send state request
					msg := message.New(message.StateMsg, nil, nil, nil, "")
					msg.SetSender(token)
					m.Process(msg)

					m.Leave(token)
				}
			}(i)
		}

		wg.Wait()
	})
}

func TestManager(t *testing.T) {
	t.Run("single player completes strict game", func(t *testing.T) {
		m := NewMock(true, true, true, 1)

		// create player
		token, err := m.Lobby().Create("alice")
		if err != nil {
			t.Fatalf("unexpected create error: '%v'", err)
		}

		// join player
		buffer, err := m.Join(token)
		if err != nil {
			t.Fatalf("unexpected join error: '%v'", err)
		}
		if buffer == nil {
			t.Fatal("expected buffer, got nil")
		}

		// receive join event
		evt, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.JoinEvent {
			t.Errorf("expected event type '%s', got '%s'", event.JoinEvent, evt.Type())
		}
		if len(evt.Players()) != 1 {
			t.Errorf("expected 1 player, got %d", len(evt.Players()))
		}
		if evt.Players()[0].Name() != "alice" {
			t.Errorf("expected player name 'alice', got '%s'", evt.Players()[0].Name())
		}
		if !evt.Players()[0].Active() {
			t.Error("expected player to be active")
		}

		// request state
		stateMsg := message.New(message.StateMsg, nil, nil, nil, "state-trace")
		stateMsg.SetSender(token)
		m.Process(stateMsg)

		evt, err = buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.StateEvent {
			t.Errorf("expected event type '%s', got '%s'", event.StateEvent, evt.Type())
		}
		if evt.Trace() != "state-trace" {
			t.Errorf("expected trace 'state-trace', got '%s'", evt.Trace())
		}
		if evt.Current() == nil {
			t.Error("expected current board, got nil")
		}
		if evt.Initial() == nil {
			t.Error("expected initial board, got nil")
		}
		if evt.Config() == nil {
			t.Error("expected config, got nil")
		}
		if len(evt.Players()) != 1 {
			t.Errorf("expected 1 player, got %d", len(evt.Players()))
		}

		// send ping
		pingMsg := message.New(message.PingMsg, ptr(4), ptr(4), nil, "ping-trace")
		pingMsg.SetSender(token)
		m.Process(pingMsg)

		evt, err = buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.PingEvent {
			t.Errorf("expected event type '%s', got '%s'", event.PingEvent, evt.Type())
		}
		if evt.Trace() != "ping-trace" {
			t.Errorf("expected trace 'ping-trace', got '%s'", evt.Trace())
		}
		if evt.Row() == nil || *evt.Row() != 4 {
			t.Errorf("expected row 4, got %v", evt.Row())
		}
		if evt.Column() == nil || *evt.Column() != 4 {
			t.Errorf("expected column 4, got %v", evt.Column())
		}
		if evt.Error() != "" {
			t.Errorf("expected no error, got '%s'", evt.Error())
		}

		// send ping with out of bounds coordinates
		pingMsg = message.New(message.PingMsg, ptr(9), ptr(0), nil, "ping-oob-trace")
		pingMsg.SetSender(token)
		m.Process(pingMsg)

		evt, err = buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.PingEvent {
			t.Errorf("expected event type '%s', got '%s'", event.PingEvent, evt.Type())
		}
		if evt.Error() == "" {
			t.Error("expected error for out of bounds ping, got none")
		}

		// insert values until game is complete
		// the mock game has initial board with clues and solution
		// we need to insert correct values for all empty non-initial cells
		solution := m.Lobby().Game().Solution()
		initial := m.Lobby().Game().Initial()

		insertCount := 0
		for row := range sudoku.BoardSize {
			for col := range sudoku.BoardSize {
				// skip initial clues
				if initial.Cell(row, col) != sudoku.EmptyCell {
					continue
				}
				// skip already filled cell (mock has [8][8] pre-filled)
				if m.Lobby().Game().Current().Cell(row, col) != sudoku.EmptyCell {
					continue
				}

				val := solution.Cell(row, col)
				insertMsg := message.New(message.InsertMsg, ptr(row), ptr(col), ptr(val), "insert-trace")
				insertMsg.SetSender(token)
				m.Process(insertMsg)

				evt, err = buffer.Receive()
				if err != nil {
					t.Fatalf("unexpected receive error on insert (%d,%d): '%v'", row, col, err)
				}
				if evt.Type() != event.InsertEvent {
					t.Errorf("expected event type '%s', got '%s'", event.InsertEvent, evt.Type())
				}
				if evt.Error() != "" {
					t.Errorf("expected no error on insert (%d,%d), got '%s'", row, col, evt.Error())
				}
				if evt.Conflict() != "" {
					t.Errorf("expected no conflict on insert (%d,%d), got '%s'", row, col, evt.Conflict())
				}
				if evt.Current() == nil {
					t.Errorf("expected current board on insert (%d,%d), got nil", row, col)
				}
				insertCount++
			}
		}

		if insertCount == 0 {
			t.Error("expected at least one insert operation")
		}

		// verify game is finished
		if m.Lobby().Game().FinishedAt() == nil {
			t.Error("expected game to be finished")
		}

		// verify current board equals solution
		if !m.Lobby().Game().Current().Equal(solution) {
			t.Error("expected current board to equal solution")
		}

		// attempt insert after game is finished
		insertMsg := message.New(message.InsertMsg, ptr(0), ptr(0), ptr(7), "post-finish-trace")
		insertMsg.SetSender(token)
		m.Process(insertMsg)

		evt, err = buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if evt.Type() != event.InsertEvent {
			t.Errorf("expected event type '%s', got '%s'", event.InsertEvent, evt.Type())
		}
		if evt.Error() == "" {
			t.Error("expected error for insert on finished game, got none")
		}

		// leave player
		err = m.Leave(token)
		if err != nil {
			t.Errorf("unexpected leave error: '%v'", err)
		}

		// verify player is inactive
		p := m.Lobby().Player(token)
		if p == nil {
			t.Fatal("expected player, got nil")
		}
		if p.Active() {
			t.Error("expected player to be inactive after leave")
		}
	})
}
