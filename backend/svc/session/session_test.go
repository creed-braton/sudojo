package session

import (
	"log/slog"
	"os"
	"sudojo/adp/database"
	"sudojo/adp/metrics"
	"sudojo/adp/socket"
	"sudojo/pkg/lobby"
	"testing"
	"time"
)

// Helper to read messages from a channel with timeout
func readWithTimeout(t *testing.T, ch <-chan *socket.Message, timeout time.Duration) *socket.Message {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(timeout):
		t.Fatal("timeout waiting for message")
		return nil
	}
}

func TestTakeover(t *testing.T) {
	// Setup: create mock dependencies
	l := lobby.NewMock(true, false, true, 4)
	db := database.NewMock()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	m := metrics.NewMock()

	// Create session service
	svc := New(l, db, logger, m)

	// Create two players in the lobby
	witnessToken, err := svc.CreatePlayer("Witness")
	if err != nil {
		t.Fatalf("failed to create witness player: %v", err)
	}

	playerToken, err := svc.CreatePlayer("Player")
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	// Create mock sockets with their channels
	witnessOut := make(chan *socket.Message, 32)
	witnessSocket := socket.NewMock(nil, witnessOut)

	playerFirstOut := make(chan *socket.Message, 32)
	playerFirstSocket := socket.NewMock(nil, playerFirstOut)

	playerSecondOut := make(chan *socket.Message, 32)
	playerSecondSocket := socket.NewMock(nil, playerSecondOut)

	// Connect the witness first
	if err := svc.JoinPlayer(witnessToken, witnessSocket); err != nil {
		t.Fatalf("failed to join witness: %v", err)
	}

	// Witness should receive a "join" message with both players listed
	// (Witness active, Player inactive since they haven't joined yet)
	msg := readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "join" {
		t.Errorf("expected message type 'join', got '%s'", msg.Type)
	}
	if len(msg.Players) != 2 {
		t.Fatalf("expected 2 players in join message, got %d", len(msg.Players))
	}

	// Verify Witness is active, Player is inactive
	for _, p := range msg.Players {
		switch p.Name {
		case "Witness":
			if !p.Active {
				t.Errorf("expected Witness to be active")
			}
		case "Player":
			if p.Active {
				t.Errorf("expected Player to be inactive (not joined yet)")
			}
		default:
			t.Errorf("unexpected player name: %s", p.Name)
		}
	}

	// Connect the player's first connection
	if err := svc.JoinPlayer(playerToken, playerFirstSocket); err != nil {
		t.Fatalf("failed to join player (first connection): %v", err)
	}

	// Witness should receive a "join" message with both players active
	msg = readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "join" {
		t.Errorf("expected message type 'join', got '%s'", msg.Type)
	}
	if len(msg.Players) != 2 {
		t.Fatalf("expected 2 players in join message, got %d", len(msg.Players))
	}

	// Verify both players are active (sorted by token, so order may vary)
	for _, p := range msg.Players {
		if !p.Active {
			t.Errorf("expected player '%s' to be active", p.Name)
		}
	}

	// Player's first socket should also receive the join message
	firstSocketMsg := readWithTimeout(t, playerFirstOut, time.Second)
	if firstSocketMsg.Type != "join" {
		t.Errorf("expected message type 'join' for first socket, got '%s'", firstSocketMsg.Type)
	}

	// Now perform takeover: connect player's second connection with same token
	if err := svc.JoinPlayer(playerToken, playerSecondSocket); err != nil {
		t.Fatalf("failed to join player (second connection / takeover): %v", err)
	}

	// The first socket should have been closed with takeover code
	select {
	case <-playerFirstSocket.Closed():
		if playerFirstSocket.CloseCode != socket.CloseTakeover {
			t.Errorf("expected close code %d (takeover), got %d",
				socket.CloseTakeover, playerFirstSocket.CloseCode)
		}
	case <-time.After(time.Second):
		t.Error("first socket was not closed after takeover")
	}

	// Witness should receive another "join" message with both players still active
	msg = readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "join" {
		t.Errorf("expected message type 'join', got '%s'", msg.Type)
	}
	if len(msg.Players) != 2 {
		t.Fatalf("expected 2 players in join message after takeover, got %d", len(msg.Players))
	}

	// Critical assertion: both players should still be active after takeover
	activeCount := 0
	for _, p := range msg.Players {
		if p.Active {
			activeCount++
		}
	}
	if activeCount != 2 {
		t.Errorf("expected 2 active players after takeover, got %d", activeCount)
	}

	// Verify the player names are correct
	names := make(map[string]bool)
	for _, p := range msg.Players {
		names[p.Name] = p.Active
	}

	if active, ok := names["Witness"]; !ok || !active {
		t.Errorf("expected Witness to be present and active")
	}
	if active, ok := names["Player"]; !ok || !active {
		t.Errorf("expected Player to be present and active")
	}

	// Cleanup
	svc.Close(socket.CloseFinished, "test complete")
}

func TestLeave(t *testing.T) {
	// Setup: create mock dependencies
	l := lobby.NewMock(true, false, true, 4)
	db := database.NewMock()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	m := metrics.NewMock()

	// Create session service
	svc := New(l, db, logger, m)

	// Create two players in the lobby
	witnessToken, err := svc.CreatePlayer("Witness")
	if err != nil {
		t.Fatalf("failed to create witness player: %v", err)
	}

	playerToken, err := svc.CreatePlayer("Player")
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	// Create mock sockets with their channels
	witnessOut := make(chan *socket.Message, 32)
	witnessSocket := socket.NewMock(nil, witnessOut)

	playerIn := make(chan *socket.Message, 32)
	playerOut := make(chan *socket.Message, 32)
	playerSocket := socket.NewMock(playerIn, playerOut)

	// Connect the witness first
	if err := svc.JoinPlayer(witnessToken, witnessSocket); err != nil {
		t.Fatalf("failed to join witness: %v", err)
	}

	// Witness receives join message
	msg := readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "join" {
		t.Errorf("expected message type 'join', got '%s'", msg.Type)
	}

	// Connect the player
	if err := svc.JoinPlayer(playerToken, playerSocket); err != nil {
		t.Fatalf("failed to join player: %v", err)
	}

	// Witness receives join message with both players active
	msg = readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "join" {
		t.Errorf("expected message type 'join', got '%s'", msg.Type)
	}

	// Verify both players are active
	activeCount := 0
	for _, p := range msg.Players {
		if p.Active {
			activeCount++
		}
	}
	if activeCount != 2 {
		t.Errorf("expected 2 active players before disconnect, got %d", activeCount)
	}

	// Player also receives the join message
	readWithTimeout(t, playerOut, time.Second)

	// Simulate client disconnect by closing the in channel
	close(playerIn)

	// Give time for the goroutines to process the disconnect
	time.Sleep(100 * time.Millisecond)

	// Witness should receive a "leave" message with updated player list
	msg = readWithTimeout(t, witnessOut, time.Second)
	if msg.Type != "leave" {
		t.Errorf("expected message type 'leave', got '%s'", msg.Type)
	}
	if len(msg.Players) != 2 {
		t.Fatalf("expected 2 players in leave message, got %d", len(msg.Players))
	}

	// Verify Witness is still active, Player is now inactive
	for _, p := range msg.Players {
		switch p.Name {
		case "Witness":
			if !p.Active {
				t.Errorf("expected Witness to still be active")
			}
		case "Player":
			if p.Active {
				t.Errorf("expected Player to be inactive after disconnect")
			}
		default:
			t.Errorf("unexpected player name: %s", p.Name)
		}
	}

	// Verify the player socket was closed
	select {
	case <-playerSocket.Closed():
		// Expected - socket should be closed
	case <-time.After(time.Second):
		t.Error("player socket was not closed after client disconnect")
	}

	// Cleanup
	svc.Close(socket.CloseFinished, "test complete")
}

func TestBroadcast(t *testing.T) {
	// Setup: create mock dependencies
	l := lobby.NewMock(true, false, true, 4)
	db := database.NewMock()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	m := metrics.NewMock()

	// Create session service
	svc := New(l, db, logger, m)

	// Create three players
	player1Token, err := svc.CreatePlayer("Player1")
	if err != nil {
		t.Fatalf("failed to create player1: %v", err)
	}

	player2Token, err := svc.CreatePlayer("Player2")
	if err != nil {
		t.Fatalf("failed to create player2: %v", err)
	}

	player3Token, err := svc.CreatePlayer("Player3")
	if err != nil {
		t.Fatalf("failed to create player3: %v", err)
	}

	// Create mock sockets for all players
	player1In := make(chan *socket.Message, 32)
	player1Out := make(chan *socket.Message, 32)
	player1Socket := socket.NewMock(player1In, player1Out)

	player2In := make(chan *socket.Message, 32)
	player2Out := make(chan *socket.Message, 32)
	player2Socket := socket.NewMock(player2In, player2Out)

	player3In := make(chan *socket.Message, 32)
	player3Out := make(chan *socket.Message, 32)
	player3Socket := socket.NewMock(player3In, player3Out)

	// Connect all players
	if err := svc.JoinPlayer(player1Token, player1Socket); err != nil {
		t.Fatalf("failed to join player1: %v", err)
	}
	readWithTimeout(t, player1Out, time.Second) // Consume join message

	if err := svc.JoinPlayer(player2Token, player2Socket); err != nil {
		t.Fatalf("failed to join player2: %v", err)
	}
	readWithTimeout(t, player1Out, time.Second) // Consume join message
	readWithTimeout(t, player2Out, time.Second) // Consume join message

	if err := svc.JoinPlayer(player3Token, player3Socket); err != nil {
		t.Fatalf("failed to join player3: %v", err)
	}
	readWithTimeout(t, player1Out, time.Second) // Consume join message
	readWithTimeout(t, player2Out, time.Second) // Consume join message
	readWithTimeout(t, player3Out, time.Second) // Consume join message

	// Player1 sends an insert message
	row := 0
	col := 0
	val := 7
	trace := "test-trace-123"
	insertMsg := &socket.Message{
		Type:   "insert",
		Trace:  trace,
		Row:    &row,
		Column: &col,
		Value:  &val,
	}

	// Send the insert message from player1
	player1In <- insertMsg

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// All three players should receive the broadcast insert message
	msg1 := readWithTimeout(t, player1Out, time.Second)
	msg2 := readWithTimeout(t, player2Out, time.Second)
	msg3 := readWithTimeout(t, player3Out, time.Second)

	// Verify all messages are insert type with correct trace
	for i, msg := range []*socket.Message{msg1, msg2, msg3} {
		if msg.Type != "insert" {
			t.Errorf("player %d: expected message type 'insert', got '%s'", i+1, msg.Type)
		}
		if msg.Trace != trace {
			t.Errorf("player %d: expected trace '%s', got '%s'", i+1, trace, msg.Trace)
		}
		if msg.Row == nil || *msg.Row != row {
			t.Errorf("player %d: expected row %d, got %v", i+1, row, msg.Row)
		}
		if msg.Column == nil || *msg.Column != col {
			t.Errorf("player %d: expected column %d, got %v", i+1, col, msg.Column)
		}
		if msg.Value == nil || *msg.Value != val {
			t.Errorf("player %d: expected value %d, got %v", i+1, val, msg.Value)
		}
		if msg.Current == nil {
			t.Errorf("player %d: expected current board to be set", i+1)
		}
	}

	// Cleanup
	svc.Close(socket.CloseFinished, "test complete")
}
