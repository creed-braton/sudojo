package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"sudojo/domain"
)

// Message types
const (
	MessageTypeMove         = "move"
	MessageTypeClear        = "clear"
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
	clients     map[*websocket.Conn]*Client
	clientsMutex sync.RWMutex
}

type Client struct {
	conn  *websocket.Conn
	mutex sync.Mutex
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
		clients:     make(map[*websocket.Conn]*Client),
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

	client := &Client{conn: conn}
	lobby.clientsMutex.Lock()
	lobby.clients[conn] = client
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

// SudokuClearMessage represents a clear request from a client
type SudokuClearMessage struct {
	Type string `json:"type"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
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
	Type         string     `json:"type"`
	Board        [][]int    `json:"board"`
	InitialBoard [][]int    `json:"initialBoard"` // Added to store the initial state
}

func (s *Server) handleLobbyMessage(lobby *Lobby, sender *websocket.Conn, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		s.sendErrorResponse(sender, "Invalid JSON format")
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		s.sendErrorResponse(sender, "Message type is required")
		return
	}

	switch msgType {
	case MessageTypeMove:
		s.handleSudokuMove(lobby, sender, message)
	case MessageTypeClear:
		s.handleSudokuClear(lobby, sender, message)
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
		s.sendErrorResponse(sender, "Invalid move format")
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
		s.safeWriteMessage(sender, responseJSON)
		return
	}

	// Move is valid, broadcast success to all clients including sender
	log.Printf("Successful move: row=%d, col=%d, value=%d in lobby %s",
		moveMsg.Row, moveMsg.Col, moveMsg.Value, lobby.domainLobby.ID)
	
	response := SudokuResponseMessage{
		Type:    MessageTypeSuccess,
		Row:     moveMsg.Row,
		Col:     moveMsg.Col,
		Value:   moveMsg.Value,
		Success: true,
	}
	responseJSON, _ := json.Marshal(response)
	
	lobby.clientsMutex.RLock()
	clients := make([]*Client, 0, len(lobby.clients))
	for _, client := range lobby.clients {
		clients = append(clients, client)
	}
	lobby.clientsMutex.RUnlock()
	
	log.Printf("Broadcasting move success to %d clients in lobby %s",
		len(clients), lobby.domainLobby.ID)
	
	// First, send the success response to all clients
	for _, client := range clients {
		s.safeWriteMessage(client.conn, responseJSON)
	}
	
	// Then, broadcast the updated board state to all clients
	board := make([][]int, domain.BoardSize)
	for i := range board {
		board[i] = make([]int, domain.BoardSize)
		for j := range board[i] {
			board[i][j] = lobby.domainLobby.Puzzle.Board[i][j]
		}
	}
	
	// Convert the initial board to a 2D slice for JSON marshaling
	initialBoard := make([][]int, domain.BoardSize)
	for i := range initialBoard {
		initialBoard[i] = make([]int, domain.BoardSize)
		for j := range initialBoard[i] {
			initialBoard[i][j] = lobby.domainLobby.InitialPuzzle.Board[i][j]
		}
	}
	
	// Create the state message with both current and initial boards
	stateMsg := SudokuStateMessage{
		Type:         MessageTypeState,
		Board:        board,
		InitialBoard: initialBoard,
	}
	
	// Marshal to JSON
	stateMsgJSON, err := json.Marshal(stateMsg)
	if err != nil {
		fmt.Println("Error marshaling board state after move:", err)
		return
	}
	
	log.Printf("Broadcasting updated board state to %d clients in lobby %s",
		len(clients), lobby.domainLobby.ID)
	
	// Broadcast the updated state to all clients
	for _, client := range clients {
		s.safeWriteMessage(client.conn, stateMsgJSON)
	}
}

func (s *Server) handleSudokuClear(lobby *Lobby, sender *websocket.Conn, message []byte) {
	var clearMsg SudokuClearMessage
	if err := json.Unmarshal(message, &clearMsg); err != nil {
		s.sendErrorResponse(sender, "Invalid clear format")
		return
	}

	// Validate and clear the cell using the initial board check
	err := lobby.domainLobby.Puzzle.ClearCellWithInitialCheck(clearMsg.Row, clearMsg.Col, lobby.domainLobby.InitialPuzzle)
	if err != nil {
		// Send error only to the sender
		response := SudokuResponseMessage{
			Type:    MessageTypeError,
			Row:     clearMsg.Row,
			Col:     clearMsg.Col,
			Value:   0, // Value is 0 for clear operations
			Success: false,
			Error:   err.Error(),
		}
		responseJSON, _ := json.Marshal(response)
		s.safeWriteMessage(sender, responseJSON)
		return
	}

	// Clear is valid, broadcast success to all clients including sender
	log.Printf("Successful clear: row=%d, col=%d in lobby %s",
		clearMsg.Row, clearMsg.Col, lobby.domainLobby.ID)
	
	response := SudokuResponseMessage{
		Type:    MessageTypeSuccess,
		Row:     clearMsg.Row,
		Col:     clearMsg.Col,
		Value:   0, // Value is 0 for clear operations
		Success: true,
	}
	responseJSON, _ := json.Marshal(response)
	
	lobby.clientsMutex.RLock()
	clients := make([]*Client, 0, len(lobby.clients))
	for _, client := range lobby.clients {
		clients = append(clients, client)
	}
	lobby.clientsMutex.RUnlock()
	
	log.Printf("Broadcasting clear success to %d clients in lobby %s",
		len(clients), lobby.domainLobby.ID)
	
	// First, send the success response to all clients
	for _, client := range clients {
		s.safeWriteMessage(client.conn, responseJSON)
	}
	
	// Then, broadcast the updated board state to all clients
	board := make([][]int, domain.BoardSize)
	for i := range board {
		board[i] = make([]int, domain.BoardSize)
		for j := range board[i] {
			board[i][j] = lobby.domainLobby.Puzzle.Board[i][j]
		}
	}
	
	// Convert the initial board to a 2D slice for JSON marshaling
	initialBoard := make([][]int, domain.BoardSize)
	for i := range initialBoard {
		initialBoard[i] = make([]int, domain.BoardSize)
		for j := range initialBoard[i] {
			initialBoard[i][j] = lobby.domainLobby.InitialPuzzle.Board[i][j]
		}
	}
	
	// Create the state message with both current and initial boards
	stateMsg := SudokuStateMessage{
		Type:         MessageTypeState,
		Board:        board,
		InitialBoard: initialBoard,
	}
	
	// Marshal to JSON
	stateMsgJSON, err := json.Marshal(stateMsg)
	if err != nil {
		fmt.Println("Error marshaling board state after clear:", err)
		return
	}
	
	log.Printf("Broadcasting updated board state to %d clients in lobby %s",
		len(clients), lobby.domainLobby.ID)
	
	// Broadcast the updated state to all clients
	for _, client := range clients {
		s.safeWriteMessage(client.conn, stateMsgJSON)
	}
}

func (s *Server) broadcastMessage(lobby *Lobby, sender *websocket.Conn, message []byte) {
	lobby.clientsMutex.RLock()
	clients := make([]*Client, 0, len(lobby.clients))
	for conn, client := range lobby.clients {
		if conn != sender {
			clients = append(clients, client)
		}
	}
	lobby.clientsMutex.RUnlock()

	for _, client := range clients {
		s.safeWriteMessage(client.conn, message)
	}
}

// safeWriteMessage writes a message to a WebSocket connection with proper synchronization
func (s *Server) safeWriteMessage(conn *websocket.Conn, message []byte) {
	// This is a simplified approach - in production you might want a more sophisticated solution
	defer func() {
		if r := recover(); r != nil {
			// Connection might be closed, ignore the error
		}
	}()
	
	conn.WriteMessage(websocket.TextMessage, message)
}

func (s *Server) sendErrorResponse(conn *websocket.Conn, errMsg string) {
	response := SudokuResponseMessage{
		Type:    MessageTypeError,
		Success: false,
		Error:   errMsg,
	}
	responseJSON, _ := json.Marshal(response)
	s.safeWriteMessage(conn, responseJSON)
}

// handleRequestState sends both the current state and initial state of the Sudoku board to the requesting client
func (s *Server) handleRequestState(lobby *Lobby, sender *websocket.Conn) {
	fmt.Println("Processing request_state message")
	
	// Debug: Print the current puzzle state
	fmt.Println("Current puzzle state:")
	for i := 0; i < domain.BoardSize; i++ {
		fmt.Println(lobby.domainLobby.Puzzle.Board[i])
	}
	
	// Convert the current board to a 2D slice for JSON marshaling
	board := make([][]int, domain.BoardSize)
	for i := range board {
		board[i] = make([]int, domain.BoardSize)
		for j := range board[i] {
			board[i][j] = lobby.domainLobby.Puzzle.Board[i][j]
		}
	}
	
	// Convert the initial board to a 2D slice for JSON marshaling
	initialBoard := make([][]int, domain.BoardSize)
	for i := range initialBoard {
		initialBoard[i] = make([]int, domain.BoardSize)
		for j := range initialBoard[i] {
			initialBoard[i][j] = lobby.domainLobby.InitialPuzzle.Board[i][j]
		}
	}

	// Create the state message with both current and initial boards
	stateMsg := SudokuStateMessage{
		Type:         MessageTypeState,
		Board:        board,
		InitialBoard: initialBoard,
	}

	// Marshal to JSON
	stateMsgJSON, err := json.Marshal(stateMsg)
	if err != nil {
		fmt.Println("Error marshaling board state:", err)
		s.sendErrorResponse(sender, "Failed to encode board state")
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

// HandleHealth is a simple health check endpoint that returns "healthy" with 200 status
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// Apply CORS middleware to all routes
	lobbyHandler := http.HandlerFunc(s.HandleLobby)
	healthHandler := http.HandlerFunc(s.HandleHealth)
	
	mux.Handle("/lobby", corsMiddleware(lobbyHandler))
	mux.Handle("/health", corsMiddleware(healthHandler))
}