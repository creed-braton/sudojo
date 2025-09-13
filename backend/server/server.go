package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"sudojo/domain"
)

// Message types
const (
	MessageTypeMove         = "move"
	MessageTypeSuccess      = "success"
	MessageTypeError        = "error"
	MessageTypeRequestState = "request_state"
	MessageTypeState        = "state"
)

// CORS middleware wraps handlers to add appropriate headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for the preflight request
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		// Set CORS headers for the main request
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		
		next.ServeHTTP(w, r)
	})
}

type Server struct {
	lobbies     map[string]*Lobby
	lobbiesMutex sync.RWMutex
	upgrader     websocket.Upgrader
}

type Lobby struct {
	domainLobby *domain.Lobby
	clients     map[*websocket.Conn]bool
	clientsMutex sync.RWMutex
}

type LobbyResponse struct {
	ID string `json:"id"`
}

func NewServer() *Server {
	return &Server{
		lobbies: make(map[string]*Lobby),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) createLobby() (*Lobby, error) {
	domainLobby, err := domain.NewLobby()
	if err != nil {
		return nil, err
	}

	lobby := &Lobby{
		domainLobby: domainLobby,
		clients:     make(map[*websocket.Conn]bool),
	}

	s.lobbiesMutex.Lock()
	s.lobbies[domainLobby.ID] = lobby
	s.lobbiesMutex.Unlock()

	return lobby, nil
}

func (s *Server) HandleLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleCreateLobby(w, r)
		return
	}

	if r.Method == http.MethodGet {
		lobbyID := r.URL.Query().Get("id")
		if lobbyID != "" {
			s.handleJoinLobby(w, r, lobbyID)
			return
		}
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleCreateLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := s.createLobby()
	if err != nil {
		http.Error(w, "Failed to create lobby", http.StatusInternalServerError)
		return
	}

	resp := LobbyResponse{
		ID: lobby.domainLobby.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleJoinLobby(w http.ResponseWriter, r *http.Request, lobbyID string) {
	s.lobbiesMutex.RLock()
	lobby, exists := s.lobbies[lobbyID]
	s.lobbiesMutex.RUnlock()

	if !exists {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	lobby.clientsMutex.Lock()
	lobby.clients[conn] = true
	lobby.clientsMutex.Unlock()

	go s.handleWebSocketConnection(conn, lobby)
}

func (s *Server) handleWebSocketConnection(conn *websocket.Conn, lobby *Lobby) {
	defer func() {
		conn.Close()
		lobby.clientsMutex.Lock()
		delete(lobby.clients, conn)
		lobby.clientsMutex.Unlock()
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			s.handleLobbyMessage(lobby, conn, message)
		}
	}
}

// SudokuMoveMessage represents a move request from a client
type SudokuMoveMessage struct {
	Type  string `json:"type"`
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Value int    `json:"value"`
}

// SudokuResponseMessage represents a response to a move
type SudokuResponseMessage struct {
	Type    string `json:"type"`
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	Value   int    `json:"value"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SudokuStateMessage represents the current state of the Sudoku board
type SudokuStateMessage struct {
	Type  string     `json:"type"`
	Board [][]int    `json:"board"`
}

func (s *Server) handleLobbyMessage(lobby *Lobby, sender *websocket.Conn, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		sendErrorResponse(sender, "Invalid JSON format")
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		sendErrorResponse(sender, "Message type is required")
		return
	}

	switch msgType {
	case MessageTypeMove:
		s.handleSudokuMove(lobby, sender, message)
	case MessageTypeRequestState:
		fmt.Println("Received request_state message")
		s.handleRequestState(lobby, sender)
	default:
		// For backwards compatibility, broadcast other message types as before
		s.broadcastMessage(lobby, sender, message)
	}
}

func (s *Server) handleSudokuMove(lobby *Lobby, sender *websocket.Conn, message []byte) {
	var moveMsg SudokuMoveMessage
	if err := json.Unmarshal(message, &moveMsg); err != nil {
		sendErrorResponse(sender, "Invalid move format")
		return
	}

	// Validate and make the move
	err := lobby.domainLobby.Puzzle.MakeMove(moveMsg.Row, moveMsg.Col, moveMsg.Value)
	if err != nil {
		// Send error only to the sender
		response := SudokuResponseMessage{
			Type:    MessageTypeError,
			Row:     moveMsg.Row,
			Col:     moveMsg.Col,
			Value:   moveMsg.Value,
			Success: false,
			Error:   err.Error(),
		}
		responseJSON, _ := json.Marshal(response)
		sender.WriteMessage(websocket.TextMessage, responseJSON)
		return
	}

	// Move is valid, broadcast success to all clients including sender
	response := SudokuResponseMessage{
		Type:    MessageTypeSuccess,
		Row:     moveMsg.Row,
		Col:     moveMsg.Col,
		Value:   moveMsg.Value,
		Success: true,
	}
	responseJSON, _ := json.Marshal(response)
	
	lobby.clientsMutex.RLock()
	defer lobby.clientsMutex.RUnlock()
	
	for client := range lobby.clients {
		client.WriteMessage(websocket.TextMessage, responseJSON)
	}
}

func (s *Server) broadcastMessage(lobby *Lobby, sender *websocket.Conn, message []byte) {
	lobby.clientsMutex.RLock()
	defer lobby.clientsMutex.RUnlock()

	for client := range lobby.clients {
		if client != sender {
			client.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func sendErrorResponse(conn *websocket.Conn, errMsg string) {
	response := SudokuResponseMessage{
		Type:    MessageTypeError,
		Success: false,
		Error:   errMsg,
	}
	responseJSON, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, responseJSON)
}

// handleRequestState sends the current state of the Sudoku board to the requesting client
func (s *Server) handleRequestState(lobby *Lobby, sender *websocket.Conn) {
	fmt.Println("Processing request_state message")
	
	// Debug: Print the current puzzle state
	fmt.Println("Current puzzle state:")
	for i := 0; i < domain.BoardSize; i++ {
		fmt.Println(lobby.domainLobby.Puzzle.Board[i])
	}
	
	// Convert the board to a 2D slice for JSON marshaling
	board := make([][]int, domain.BoardSize)
	for i := range board {
		board[i] = make([]int, domain.BoardSize)
		for j := range board[i] {
			board[i][j] = lobby.domainLobby.Puzzle.Board[i][j]
		}
	}

	// Create the state message
	stateMsg := SudokuStateMessage{
		Type:  MessageTypeState,
		Board: board,
	}

	// Marshal to JSON
	stateMsgJSON, err := json.Marshal(stateMsg)
	if err != nil {
		fmt.Println("Error marshaling board state:", err)
		sendErrorResponse(sender, "Failed to encode board state")
		return
	}

	fmt.Println("Sending state message to client")
	
	// Send only to the requesting client
	err = sender.WriteMessage(websocket.TextMessage, stateMsgJSON)
	if err != nil {
		fmt.Println("Error sending state message:", err)
	} else {
		fmt.Println("State message sent successfully")
	}
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// Apply CORS middleware to all routes
	lobbyHandler := http.HandlerFunc(s.HandleLobby)
	mux.Handle("/lobby", corsMiddleware(lobbyHandler))
}