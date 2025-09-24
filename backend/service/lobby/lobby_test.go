package lobby

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sudojo/domain/sudoku"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func setupTestServer() *httptest.Server {
	service := New()
	router := http.NewServeMux()

	for path, methods := range service.Routes() {
		for method, handler := range methods {
			pattern := fmt.Sprintf("%s %s", method, path)
			router.HandleFunc(pattern, handler)
		}
	}

	server := httptest.NewServer(router)
	return server
}

func createLobby(t *testing.T, server *httptest.Server) uuid.UUID {
	res, err := http.Post(server.URL+"/lobbies", "text/plain", nil)
	if err != nil {
		t.Fatalf("failed to make post call: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("received wrong status code, got: %d, want: %d", res.StatusCode, http.StatusCreated)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	id, err := uuid.Parse(string(body))
	if err != nil {
		t.Fatalf("server returned an invalid uuid in response: %v", err)
	}

	return id
}

func joinLobby(t *testing.T, server *httptest.Server, id uuid.UUID) *websocket.Conn {
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.Path = "/lobbies/" + id.String()
	ws, res, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("websocket dial failed (status: %d): %v", res.StatusCode, err)
	}
	return ws
}

func sendMessage(t *testing.T, ws *websocket.Conn, msg *Request) {
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to serialize request: %v", err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}
}

func readMessage(t *testing.T, ws *websocket.Conn) *Response {
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	r := &Response{}
	if err := json.Unmarshal(msg, r); err != nil {
		t.Fatalf("failed to serialize response: %v", err)
	}

	return r
}

func TestInvalidJson(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	id := createLobby(t, server)

	ws := joinLobby(t, server, id)
	defer ws.Close()

	if err := ws.WriteMessage(websocket.TextMessage, []byte("test")); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	msg := readMessage(t, ws)
	got := msg.Error
	want := "invalid json format"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("{}")); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	msg = readMessage(t, ws)
	got = msg.Error
	want = "invalid message type"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestSoloLobby(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	id := createLobby(t, server)

	ws := joinLobby(t, server, id)
	defer ws.Close()

	req := &Request{Type: "state"}
	sendMessage(t, ws, req)
	msg := readMessage(t, ws)
	puzzle := msg.Current
	solution := sudoku.New()
	puzzle.Copy(solution)
	if solution.Complete() {
		t.Fatal("received puzzle is complete")
	}
	if !solution.UniqueSolution() {
		t.Fatal("received puzzle doesn't have a unique solution")
	}

	for !puzzle.Complete() {
		for row := 0; row < sudoku.BoardSize; row++ {
			for col := 0; col < sudoku.BoardSize; col++ {
				if puzzle[row][col] == sudoku.EmptyCell {
					puzzle[row][col] = solution[row][col]
					req := &Request{Type: "move", Row: row, Column: col, Value: solution[row][col]}
					sendMessage(t, ws, req)
					msg := readMessage(t, ws)
					if !msg.Current.Is(puzzle) {
						t.Fatalf(
							"current state wasn't correctly updated at row: %d, column: %d, got: %d, want: %d",
							row, col, msg.Current[row][col], puzzle[row][col],
						)
					}
				}
			}
		}
	}
}
