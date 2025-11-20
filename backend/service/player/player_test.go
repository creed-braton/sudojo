package player

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sudojo/pkg/sudoku"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func setup() (string, func()) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		client := socket.NewClient(conn)
		lobby := lobby.Open(8, false)
		buffer := event.NewBuffer()
		New(
			buffer, client,
			lobby, buffer.Send,
			func() {}, "",
			slog.With(),
		).Start()
	}))

	return "ws" + server.URL[len("http"):], server.Close
}

func TestService(t *testing.T) {
	url, shutdown := setup()
	defer shutdown()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Errorf("dial error: %v", err)
	}
	defer conn.Close()

	msg := &socket.Message{Type: event.StateEvent}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("unexpected error serializing message '%v'", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		t.Fatalf("unexpected error writing message '%v'", err)
	}
	_, b, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message '%v'", err)
	}
	if err := json.Unmarshal(b, msg); err != nil {
		t.Fatalf("unexpected error deserializing message '%v'", err)
	}

	initial := sudoku.NewFromInts(msg.Initial)
	solution := sudoku.New()
	initial.Copy(solution)
	solution.UniqueSolution()

	for i := range len(initial) {
		for j := range len(initial[i]) {
			msg := &socket.Message{
				Type: event.InsertEvent,
				Row:  &i, Column: &j,
				Value: &solution[i][j],
			}
			time.Sleep(500 * time.Millisecond)
			b, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("unexpected error serializing message '%v'", err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				t.Fatalf("unexpected error writing message '%v'", err)
			}

			_, b, err = conn.ReadMessage()
			if err != nil {
				t.Fatalf("unexpected error reading message '%v'", err)
			}
			msg = &socket.Message{}
			if err := json.Unmarshal(b, msg); err != nil {
				t.Fatalf("unexpected error deserializing message '%v'", err)
			}

			if initial.Cell(i, j) == sudoku.EmptyCell && msg.Error != "" {
				t.Errorf("unexpected error received from service '%s'", msg.Error)
			} else if initial.Cell(i, j) != sudoku.EmptyCell && msg.Error == "" {
				t.Error("expected initial clue error")
			}

			if msg.Error == "" && msg.Current == nil {
				t.Error("expected current board state")
			}
		}
	}
}
