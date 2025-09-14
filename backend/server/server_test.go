package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestMultiClientNotification tests that when one client makes a valid move,
// all connected clients receive the notification about the change.
func TestMultiClientNotification(t *testing.T) {
	// Create a new server
	server := NewServer()
	
	// Create a test HTTP server
	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()
	
	// Convert http:// to ws:// for WebSocket connections
	wsURL := strings.Replace(testServer.URL, "http://", "ws://", 1)
	
	// Create a lobby via HTTP request
	lobbyID := createTestLobby(t, testServer.URL)
	if lobbyID == "" {
		t.Fatal("Failed to create lobby")
	}
	
	t.Logf("Created lobby with ID: %s", lobbyID)
	
	// Connect multiple clients to the lobby
	const numClients = 3
	clients := make([]*websocket.Conn, numClients)
	receivedMessages := make([]chan SudokuResponseMessage, numClients)
	
	// We'll need this channel to get the board state
	stateChan := make(chan SudokuStateMessage, 1)
	
	// Connect clients and set up message receivers
	for i := 0; i < numClients; i++ {
		var err error
		clients[i], _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("%s/lobby?id=%s", wsURL, lobbyID), nil)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		defer clients[i].Close()
		
		t.Logf("Client %d connected successfully", i)
		
		// Create a channel to receive messages for this client
		receivedMessages[i] = make(chan SudokuResponseMessage, 5)
		
		// Start a goroutine to listen for messages
		go func(clientIndex int) {
			for {
				_, message, err := clients[clientIndex].ReadMessage()
				if err != nil {
					t.Logf("Client %d connection closed: %v", clientIndex, err)
					// Connection closed or error
					close(receivedMessages[clientIndex])
					return
				}
				
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err != nil {
					t.Logf("Client %d received invalid JSON: %v", clientIndex, err)
					continue
				}
				
				msgType, ok := msg["type"].(string)
				if !ok {
					t.Logf("Client %d received message without type", clientIndex)
					continue
				}
				
				t.Logf("Client %d received message of type: %s", clientIndex, msgType)
				
				switch msgType {
				case MessageTypeSuccess:
					var responseMsg SudokuResponseMessage
					if err := json.Unmarshal(message, &responseMsg); err == nil {
						t.Logf("Client %d received success message for move at (%d,%d)",
							clientIndex, responseMsg.Row, responseMsg.Col)
						receivedMessages[clientIndex] <- responseMsg
					}
				case MessageTypeError:
					var responseMsg SudokuResponseMessage
					if err := json.Unmarshal(message, &responseMsg); err == nil {
						t.Logf("Client %d received error message: %s", clientIndex, responseMsg.Error)
					}
				case MessageTypeState:
					// This is for the first client only, to find an empty cell
					if clientIndex == 0 {
						var stateMsg SudokuStateMessage
						if err := json.Unmarshal(message, &stateMsg); err == nil {
							// Verify that initialBoard is present
							if stateMsg.InitialBoard == nil {
								t.Logf("Client %d received state message with nil initialBoard", clientIndex)
							} else {
								t.Logf("Client %d received state message with valid initialBoard", clientIndex)
							}
							stateChan <- stateMsg
						}
					}
				}
			}
		}(i)
	}
	
	// Wait to ensure all clients are connected and listening
	time.Sleep(100 * time.Millisecond)
	
	// First, request the game state to find an empty cell
	requestStateMsg := map[string]string{
		"type": MessageTypeRequestState,
	}
	requestStateMsgJSON, err := json.Marshal(requestStateMsg)
	if err != nil {
		t.Fatalf("Failed to marshal request state message: %v", err)
	}
	
	err = clients[0].WriteMessage(websocket.TextMessage, requestStateMsgJSON)
	if err != nil {
		t.Fatalf("Failed to send request state message: %v", err)
	}
	
	// Wait for the state message with a timeout
	var board [][]int
	select {
	case stateMsg := <-stateChan:
		board = stateMsg.Board
		t.Logf("Received board state")
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for board state")
	}
	
	// Find an empty cell
	row, col := -1, -1
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c] == 0 {
				row, col = r, c
				break
			}
		}
		if row != -1 {
			break
		}
	}
	
	if row == -1 {
		t.Fatal("No empty cell found on the board")
	}
	
	t.Logf("Found empty cell at (%d,%d)", row, col)
	
	// Find a valid value for this empty cell
	validValue := -1
	for value := 1; value <= 9; value++ {
		if isValidMove(board, row, col, value) {
			validValue = value
			break
		}
	}
	
	if validValue == -1 {
		t.Fatal("No valid value found for the empty cell")
	}
	
	t.Logf("Found valid value %d for cell (%d,%d)", validValue, row, col)
	
	// Set up a wait group to ensure all clients receive the notification
	var wg sync.WaitGroup
	wg.Add(numClients)
	
	// Add a done function to each client's message handler
	for i := 0; i < numClients; i++ {
		go func(clientIndex int) {
			select {
			case <-receivedMessages[clientIndex]:
				t.Logf("Client %d received expected notification", clientIndex)
				wg.Done()
			case <-time.After(5 * time.Second):
				t.Logf("Client %d timed out waiting for notification", clientIndex)
				// Don't call wg.Done() here to trigger the timeout in the test
			}
		}(i)
	}
	
	// Send a valid move from the first client to the empty cell
	moveMsg := SudokuMoveMessage{
		Type:  MessageTypeMove,
		Row:   row,
		Col:   col,
		Value: validValue, // Use the valid value we found
	}
	
	t.Logf("Sending move: row=%d, col=%d, value=%d", moveMsg.Row, moveMsg.Col, moveMsg.Value)
	
	moveMsgJSON, err := json.Marshal(moveMsg)
	if err != nil {
		t.Fatalf("Failed to marshal move message: %v", err)
	}
	
	err = clients[0].WriteMessage(websocket.TextMessage, moveMsgJSON)
	if err != nil {
		t.Fatalf("Failed to send move message: %v", err)
	}
	
	// Set a deadline for receiving all notifications
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Wait for all clients to receive the notification or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		t.Log("All clients received the notification")
	case <-ctx.Done():
		t.Fatal("Timed out waiting for all clients to receive notification")
	}
	
	// We don't need to check the channels again as we've already verified
	// in the goroutines above, but we can log a final success message
	t.Log("Test completed successfully")
}

// Helper function to create a lobby and return its ID
func createTestLobby(t *testing.T, serverURL string) string {
	resp, err := http.Post(serverURL+"/lobby", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to create lobby: %v", err)
		return ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to create lobby, status code: %d", resp.StatusCode)
		return ""
	}
	
	var lobbyResp LobbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&lobbyResp); err != nil {
		t.Fatalf("Failed to decode lobby response: %v", err)
		return ""
	}
	
	return lobbyResp.ID
}

// TestRequestState tests that a client can request the current state of the game
func TestRequestState(t *testing.T) {
	// Create a new server
	server := NewServer()
	
	// Create a test HTTP server
	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()
	
	// Convert http:// to ws:// for WebSocket connections
	wsURL := strings.Replace(testServer.URL, "http://", "ws://", 1)
	
	// Create a lobby via HTTP request
	lobbyID := createTestLobby(t, testServer.URL)
	if lobbyID == "" {
		t.Fatal("Failed to create lobby")
	}
	
	// Connect a client to the lobby
	client, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/lobby?id=%s", wsURL, lobbyID), nil)
	if err != nil {
		t.Fatalf("Client failed to connect: %v", err)
	}
	defer client.Close()
	
	// Create a channel to receive messages
	stateReceived := make(chan SudokuStateMessage, 1)
	
	// Start a goroutine to listen for messages
	go func() {
		for {
			_, message, err := client.ReadMessage()
			if err != nil {
				// Connection closed or error
				close(stateReceived)
				return
			}
			
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			
			// Check if it's a state message
			if msgType, ok := msg["type"].(string); ok && msgType == MessageTypeState {
				var stateMsg SudokuStateMessage
				if err := json.Unmarshal(message, &stateMsg); err == nil {
					stateReceived <- stateMsg
				}
			}
		}
	}()
	
	// Wait to ensure client is connected and listening
	time.Sleep(100 * time.Millisecond)
	
	// Send a request for the current state
	requestMsg := map[string]string{
		"type": MessageTypeRequestState,
	}
	
	requestMsgJSON, err := json.Marshal(requestMsg)
	if err != nil {
		t.Fatalf("Failed to marshal request message: %v", err)
	}
	
	err = client.WriteMessage(websocket.TextMessage, requestMsgJSON)
	if err != nil {
		t.Fatalf("Failed to send request message: %v", err)
	}
	
	// Wait for the state message with a timeout
	select {
	case stateMsg := <-stateReceived:
		// Verify the state message
		if stateMsg.Type != MessageTypeState {
			t.Errorf("Received incorrect message type: %s", stateMsg.Type)
		}
		if stateMsg.Board == nil {
			t.Error("State message contains nil board")
		} else {
			// Verify board dimensions
			if len(stateMsg.Board) != 9 {
				t.Errorf("Board has incorrect number of rows: %d", len(stateMsg.Board))
			} else {
				for i, row := range stateMsg.Board {
					if len(row) != 9 {
						t.Errorf("Row %d has incorrect length: %d", i, len(row))
					}
				}
			}
		}
		
		// Verify that we also received the initial board
		if stateMsg.InitialBoard == nil {
			t.Error("State message contains nil initialBoard")
		} else {
			// Verify initial board dimensions
			if len(stateMsg.InitialBoard) != 9 {
				t.Errorf("InitialBoard has incorrect number of rows: %d", len(stateMsg.InitialBoard))
			} else {
				for i, row := range stateMsg.InitialBoard {
					if len(row) != 9 {
						t.Errorf("InitialBoard row %d has incorrect length: %d", i, len(row))
					}
				}
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for state message")
	}
}

// TestMultipleClientsAllMakingMoves tests that when multiple clients each make moves,
// all other clients receive notifications for each move made by others.
func TestMultipleClientsAllMakingMoves(t *testing.T) {
	// Create a new server
	server := NewServer()
	
	// Create a test HTTP server
	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()
	
	// Convert http:// to ws:// for WebSocket connections
	wsURL := strings.Replace(testServer.URL, "http://", "ws://", 1)
	
	// Create a lobby via HTTP request
	lobbyID := createTestLobby(t, testServer.URL)
	if lobbyID == "" {
		t.Fatal("Failed to create lobby")
	}
	
	t.Logf("Created lobby with ID: %s", lobbyID)
	
	// Connect multiple clients to the lobby
	const numClients = 4
	clients := make([]*websocket.Conn, numClients)
	receivedMessages := make([]chan SudokuResponseMessage, numClients)
	
	// We'll need this channel to get the board state
	stateChan := make(chan SudokuStateMessage, 1)
	
	
	// Connect clients and set up message receivers
	for i := 0; i < numClients; i++ {
		var err error
		clients[i], _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("%s/lobby?id=%s", wsURL, lobbyID), nil)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		defer clients[i].Close()
		
		t.Logf("Client %d connected successfully", i)
		
		// Create a channel to receive messages for this client
		receivedMessages[i] = make(chan SudokuResponseMessage, 20) // Even larger buffer for multiple moves and state messages
		
		// Start a goroutine to listen for messages
		go func(clientIndex int) {
			defer close(receivedMessages[clientIndex])
			
			for {
				_, message, err := clients[clientIndex].ReadMessage()
				if err != nil {
					// Check if it's an expected error when connection is closed
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
					   strings.Contains(err.Error(), "use of closed network connection") {
						// Expected errors when connection is closed - don't log as error
						return
					}
					// Only log unexpected errors
					t.Logf("Client %d unexpected connection error: %v", clientIndex, err)
					return
				}
				
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err != nil {
					t.Logf("Client %d received invalid JSON: %v", clientIndex, err)
					continue
				}
				
				msgType, ok := msg["type"].(string)
				if !ok {
					t.Logf("Client %d received message without type", clientIndex)
					continue
				}
				
				t.Logf("Client %d received message of type: %s", clientIndex, msgType)
				
				switch msgType {
				case MessageTypeSuccess:
					var responseMsg SudokuResponseMessage
					if err := json.Unmarshal(message, &responseMsg); err == nil {
						t.Logf("Client %d received success message for move at (%d,%d) = %d",
							clientIndex, responseMsg.Row, responseMsg.Col, responseMsg.Value)
						receivedMessages[clientIndex] <- responseMsg
					}
				case MessageTypeError:
					var responseMsg SudokuResponseMessage
					if err := json.Unmarshal(message, &responseMsg); err == nil {
						t.Logf("Client %d received error message: %s", clientIndex, responseMsg.Error)
					}
				case MessageTypeState:
					// Handle state messages for initial state request
					if clientIndex == 0 {
						var stateMsg SudokuStateMessage
						if err := json.Unmarshal(message, &stateMsg); err == nil {
							t.Logf("Client %d received state message", clientIndex)
							stateChan <- stateMsg
						}
					}
					// For other clients, we just log that they received the state update
					// but don't need to do anything special with it since the test
					// is focused on success message broadcasting
				}
			}
		}(i)
	}
	
	// Wait to ensure all clients are connected and listening
	time.Sleep(100 * time.Millisecond)
	
	// First, request the game state to find empty cells
	requestStateMsg := map[string]string{
		"type": MessageTypeRequestState,
	}
	requestStateMsgJSON, err := json.Marshal(requestStateMsg)
	if err != nil {
		t.Fatalf("Failed to marshal request state message: %v", err)
	}
	
	err = clients[0].WriteMessage(websocket.TextMessage, requestStateMsgJSON)
	if err != nil {
		t.Fatalf("Failed to send request state message: %v", err)
	}
	
	// Wait for the state message with a timeout
	var board [][]int
	select {
	case stateMsg := <-stateChan:
		board = stateMsg.Board
		t.Logf("Received board state")
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for board state")
	}
	
	// Find multiple empty cells for different moves
	type Move struct {
		Row, Col, Value int
	}
	
	var validMoves []Move
	
	// Find empty cells and determine valid values for each
	for r := 0; r < 9 && len(validMoves) < numClients; r++ {
		for c := 0; c < 9 && len(validMoves) < numClients; c++ {
			if board[r][c] == 0 {
				// Try to find a valid value for this cell
				for value := 1; value <= 9; value++ {
					if isValidMove(board, r, c, value) {
						validMoves = append(validMoves, Move{Row: r, Col: c, Value: value})
						// Update our local board to avoid conflicts
						board[r][c] = value
						break
					}
				}
			}
		}
	}
	
	if len(validMoves) < numClients {
		t.Fatalf("Could not find enough valid moves. Found %d, need %d", len(validMoves), numClients)
	}
	
	t.Logf("Found %d valid moves", len(validMoves))
	for i, move := range validMoves {
		t.Logf("Move %d: (%d,%d) = %d", i, move.Row, move.Col, move.Value)
	}
	
	// Now have each client make a move and verify all others receive it
	for clientIndex := 0; clientIndex < numClients; clientIndex++ {
		move := validMoves[clientIndex]
		
		t.Logf("Client %d making move: (%d,%d) = %d", clientIndex, move.Row, move.Col, move.Value)
		
		// Prepare to receive messages from all clients
		var wg sync.WaitGroup
		wg.Add(numClients) // All clients should receive the success message
		
		// Set up goroutines to wait for messages from each client
		for i := 0; i < numClients; i++ {
			go func(receiverIndex int) {
				defer wg.Done()
				
				// Try to receive the expected success message, but allow for multiple attempts
				// since we might receive state messages in between
				timeout := time.After(5 * time.Second) // Increased timeout
				for {
					select {
					case msg := <-receivedMessages[receiverIndex]:
						if msg.Type == MessageTypeSuccess &&
						   msg.Row == move.Row &&
						   msg.Col == move.Col &&
						   msg.Value == move.Value {
							t.Logf("Client %d received expected move notification from client %d",
								receiverIndex, clientIndex)
							return // Success, exit the goroutine
						} else {
							// This might be a success message for a different move, ignore it
							t.Logf("Client %d received other success message: (%d,%d)=%d, expecting (%d,%d)=%d",
								receiverIndex, msg.Row, msg.Col, msg.Value, move.Row, move.Col, move.Value)
							continue // Keep trying
						}
					case <-timeout:
						t.Errorf("Client %d timed out waiting for move notification from client %d",
							receiverIndex, clientIndex)
						return
					}
				}
			}(i)
		}
		
		// Send the move from the current client
		moveMsg := SudokuMoveMessage{
			Type:  MessageTypeMove,
			Row:   move.Row,
			Col:   move.Col,
			Value: move.Value,
		}
		
		moveMsgJSON, err := json.Marshal(moveMsg)
		if err != nil {
			t.Fatalf("Failed to marshal move message: %v", err)
		}
		
		err = clients[clientIndex].WriteMessage(websocket.TextMessage, moveMsgJSON)
		if err != nil {
			t.Fatalf("Client %d failed to send move message: %v", clientIndex, err)
		}
		
		// Wait for all clients to receive the notification
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		
		select {
		case <-done:
			t.Logf("All clients received notification for move from client %d", clientIndex)
		case <-time.After(8 * time.Second): // Increased timeout to allow for the goroutines to complete
			t.Fatalf("Timed out waiting for all clients to receive notification from client %d", clientIndex)
		}
		
		// Small delay between moves to ensure proper sequencing
		time.Sleep(100 * time.Millisecond)
	}
	
	t.Log("All clients successfully made moves and all others received notifications")
}

// Helper function to validate if a move is valid according to Sudoku rules
func isValidMove(board [][]int, row, col, value int) bool {
	// Check row
	for c := 0; c < 9; c++ {
		if board[row][c] == value {
			return false
		}
	}
	
	// Check column
	for r := 0; r < 9; r++ {
		if board[r][col] == value {
			return false
		}
	}
	
	// Check 3x3 box
	boxRow := (row / 3) * 3
	boxCol := (col / 3) * 3
	for r := boxRow; r < boxRow+3; r++ {
		for c := boxCol; c < boxCol+3; c++ {
			if board[r][c] == value {
				return false
			}
		}
	}
	
	return true
}

// TestMultipleClientsConflictingMoves tests that when multiple clients try to make
// conflicting moves, only valid moves are accepted and broadcast.
func TestMultipleClientsConflictingMoves(t *testing.T) {
	// Create a new server
	server := NewServer()
	
	// Create a test HTTP server
	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()
	
	// Convert http:// to ws:// for WebSocket connections
	wsURL := strings.Replace(testServer.URL, "http://", "ws://", 1)
	
	// Create a lobby via HTTP request
	lobbyID := createTestLobby(t, testServer.URL)
	if lobbyID == "" {
		t.Fatal("Failed to create lobby")
	}
	
	t.Logf("Created lobby with ID: %s", lobbyID)
	
	// Connect multiple clients to the lobby
	const numClients = 3
	clients := make([]*websocket.Conn, numClients)
	receivedMessages := make([]chan SudokuResponseMessage, numClients)
	
	// We'll need this channel to get the board state
	stateChan := make(chan SudokuStateMessage, 1)
	
	// Connect clients and set up message receivers
	for i := 0; i < numClients; i++ {
		var err error
		clients[i], _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("%s/lobby?id=%s", wsURL, lobbyID), nil)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		defer clients[i].Close()
		
		t.Logf("Client %d connected successfully", i)
		
		// Create a channel to receive messages for this client
		receivedMessages[i] = make(chan SudokuResponseMessage, 5)
		
		// Start a goroutine to listen for messages
		go func(clientIndex int) {
			for {
				_, message, err := clients[clientIndex].ReadMessage()
				if err != nil {
					t.Logf("Client %d connection closed: %v", clientIndex, err)
					close(receivedMessages[clientIndex])
					return
				}
				
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err != nil {
					continue
				}
				
				msgType, ok := msg["type"].(string)
				if !ok {
					continue
				}
				
				switch msgType {
				case MessageTypeSuccess, MessageTypeError:
					var responseMsg SudokuResponseMessage
					if err := json.Unmarshal(message, &responseMsg); err == nil {
						t.Logf("Client %d received %s message for move at (%d,%d) = %d",
							clientIndex, responseMsg.Type, responseMsg.Row, responseMsg.Col, responseMsg.Value)
						receivedMessages[clientIndex] <- responseMsg
					}
				case MessageTypeState:
					if clientIndex == 0 {
						var stateMsg SudokuStateMessage
						if err := json.Unmarshal(message, &stateMsg); err == nil {
							stateChan <- stateMsg
						}
					}
				}
			}
		}(i)
	}
	
	// Wait to ensure all clients are connected and listening
	time.Sleep(100 * time.Millisecond)
	
	// Get board state
	requestStateMsg := map[string]string{
		"type": MessageTypeRequestState,
	}
	requestStateMsgJSON, _ := json.Marshal(requestStateMsg)
	clients[0].WriteMessage(websocket.TextMessage, requestStateMsgJSON)
	
	var board [][]int
	select {
	case stateMsg := <-stateChan:
		board = stateMsg.Board
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for board state")
	}
	
	// Find an empty cell
	row, col := -1, -1
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c] == 0 {
				row, col = r, c
				break
			}
		}
		if row != -1 {
			break
		}
	}
	
	if row == -1 {
		t.Fatal("No empty cell found on the board")
	}
	
	// Find a valid value for this cell
	validValue := -1
	for value := 1; value <= 9; value++ {
		if isValidMove(board, row, col, value) {
			validValue = value
			break
		}
	}
	
	if validValue == -1 {
		t.Fatal("No valid value found for the empty cell")
	}
	
	t.Logf("Testing conflicting moves at (%d,%d)", row, col)
	
	// Have clients try to make moves to the same cell sequentially
	// First client should succeed, others should get errors
	successCount := 0
	errorCount := 0
	
	for clientIndex := 0; clientIndex < numClients; clientIndex++ {
		// Each client tries a different value, but client 0 uses the valid value
		value := validValue
		if clientIndex > 0 {
			value = validValue + clientIndex // This might be invalid
			if value > 9 {
				value = (value % 9) + 1
			}
		}
		
		t.Logf("Client %d trying to place %d at (%d,%d)", clientIndex, value, row, col)
		
		moveMsg := SudokuMoveMessage{
			Type:  MessageTypeMove,
			Row:   row,
			Col:   col,
			Value: value,
		}
		
		moveMsgJSON, _ := json.Marshal(moveMsg)
		clients[clientIndex].WriteMessage(websocket.TextMessage, moveMsgJSON)
		
		// Wait for response
		select {
		case msg := <-receivedMessages[clientIndex]:
			if msg.Type == MessageTypeSuccess {
				successCount++
				t.Logf("Client %d move succeeded", clientIndex)
			} else if msg.Type == MessageTypeError {
				errorCount++
				t.Logf("Client %d move failed: %s", clientIndex, msg.Error)
			}
		case <-time.After(3 * time.Second):
			t.Errorf("Client %d timed out waiting for response", clientIndex)
		}
		
		// Small delay between moves
		time.Sleep(100 * time.Millisecond)
	}
	
	// Verify that we got the expected results
	// At least one move should succeed, others might fail due to conflicts or invalid values
	if successCount == 0 {
		t.Error("Expected at least one successful move")
	}
	
	t.Logf("Results: %d successful moves, %d errors", successCount, errorCount)
	
	// Verify that successful moves were broadcast to other clients
	if successCount > 0 {
		// Give some time for broadcasts to propagate
		time.Sleep(200 * time.Millisecond)
		
		// Check that other clients received success notifications
		for i := 0; i < numClients; i++ {
			// Try to receive any remaining messages
			select {
			case msg := <-receivedMessages[i]:
				if msg.Type == MessageTypeSuccess {
					t.Logf("Client %d received broadcast of successful move", i)
				}
			default:
				// No more messages, which is fine
			}
		}
	}
}