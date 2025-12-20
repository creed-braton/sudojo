package lobby

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sudojo/adapter/database"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sudojo/pkg/manager"
	"testing"

	"github.com/gorilla/websocket"
)

func connect(lobby Service, token string) (string, func()) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		err = lobby.JoinPlayer(token, conn)
		if err != nil {
			panic(err)
		}
	}))

	return "ws" + server.URL[len("http"):], server.Close
}

func TestService(t *testing.T) {
	lobby := New(
		manager.New(lobby.Open(false, 8)), database.NewMock(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	name := "test-player"
	token, err := lobby.CreatePlayer(name)
	if err != nil {
		t.Fatalf("unexpected error creating player '%v'", err.Error())
	}

	url, shutdown := connect(lobby, token)
	defer shutdown()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Errorf("dial error: %v", err)
	}
	defer conn.Close()

	_, b, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error reading message '%v'", err)
	}
	msg := &socket.Message{}
	if err := json.Unmarshal(b, msg); err != nil {
		t.Fatalf("unexpected error deserializing message '%v'", err)
	}

	if msg.Type != event.JoinEvent {
		t.Errorf("expected event type '%s', got '%s'", event.JoinEvent, msg.Type)
	}
	if len(msg.Players) != 1 {
		t.Fatalf("expected players length '%d', got '%d'", 1, len(msg.Players))
	}
	if !msg.Players[0].Active {
		t.Errorf("expected player active '%t', got '%t'", true, msg.Players[0].Active)
	}
	if msg.Players[0].Name != name {
		t.Errorf("expected player name '%s', got '%s'", name, msg.Players[0].Name)
	}
}
