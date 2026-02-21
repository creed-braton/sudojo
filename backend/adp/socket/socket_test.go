package socket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestReadWrite(t *testing.T) {
	upgrader := websocket.Upgrader{}
	serverReady := make(chan *socket)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade connection: %v", err)
			return
		}

		s := New(conn, 60, 20)
		serverReady <- s
		_ = s.Listen()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer client.Close()

	s := <-serverReady

	// Test server Send -> client receives
	sentMsg := &Message{Type: "test", Trace: "abc123"}
	if err := s.Send(sentMsg); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	_, data, err := client.ReadMessage()
	if err != nil {
		t.Fatalf("client failed to read message: %v", err)
	}

	var receivedMsg Message
	if err := json.Unmarshal(data, &receivedMsg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if receivedMsg.Type != sentMsg.Type || receivedMsg.Trace != sentMsg.Trace {
		t.Errorf("message mismatch: expected %+v, got %+v", sentMsg, receivedMsg)
	}

	// Test client sends -> server Receive
	clientMsg := &Message{Type: "response", Trace: "xyz789"}
	data, _ = json.Marshal(clientMsg)
	if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("client failed to write message: %v", err)
	}

	gotMsg, err := s.Receive()
	if err != nil {
		t.Fatalf("server failed to receive message: %v", err)
	}

	if gotMsg.Type != clientMsg.Type || gotMsg.Trace != clientMsg.Trace {
		t.Errorf("message mismatch: expected %+v, got %+v", clientMsg, gotMsg)
	}

	s.Close(websocket.CloseNormalClosure, "test done")
}

func TestClose(t *testing.T) {
	expectedCode := 4001
	expectedMsg := "server shutting down"

	upgrader := websocket.Upgrader{}
	serverReady := make(chan *socket)
	listenErr := make(chan error, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade connection: %v", err)
			return
		}

		s := New(conn, 60, 20)
		serverReady <- s
		listenErr <- s.Listen()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer client.Close()

	s := <-serverReady

	// Start a blocking Receive() call before Close() is called
	var wg sync.WaitGroup
	receiveErr := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := s.Receive()
		receiveErr <- err
	}()

	s.Close(expectedCode, expectedMsg)

	// Verify blocking Receive() unblocks and returns ErrClosed
	wg.Wait()
	if err := <-receiveErr; err != ErrClosed {
		t.Errorf("expected blocking Receive to return ErrClosed, got %v", err)
	}

	// Verify Listen() returns an error after Close()
	err = <-listenErr
	if err == nil {
		t.Fatal("expected Listen() to return an error after Close(), got nil")
	}

	// Verify client receives correct close code
	_, _, err = client.ReadMessage()
	if err == nil {
		t.Fatal("expected close error, got nil")
	}

	closeErr, ok := err.(*websocket.CloseError)
	if !ok {
		t.Fatalf("expected CloseError, got %T: %v", err, err)
	}

	if closeErr.Code != expectedCode {
		t.Errorf("expected close code %d, got %d", expectedCode, closeErr.Code)
	}

	if closeErr.Text != expectedMsg {
		t.Errorf("expected close message %q, got %q", expectedMsg, closeErr.Text)
	}
}

func TestPing(t *testing.T) {
	upgrader := websocket.Upgrader{}
	listenErr := make(chan error, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("failed to upgrade connection: %v", err)
			return
		}

		// Use short timeout (2s) and ping interval (1s)
		// so the test doesn't take too long
		s := New(conn, 2, 1)
		listenErr <- s.Listen()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer client.Close()

	// Override the default ping handler to NOT send a pong response
	client.SetPingHandler(func(string) error {
		return nil
	})

	// Keep reading to process control frames (pings)
	// This goroutine will exit when the connection closes
	clientClosed := make(chan error, 1)
	go func() {
		for {
			_, _, err := client.ReadMessage()
			if err != nil {
				clientClosed <- err
				return
			}
		}
	}()

	// Wait for the server to close due to pong timeout
	select {
	case err := <-listenErr:
		if err == nil {
			t.Fatal("expected Listen() to return an error due to pong timeout, got nil")
		}
		// Success - the connection was closed due to missing pong
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server to close connection")
	}

	// Verify client receives correct close code
	select {
	case err := <-clientClosed:
		closeErr, ok := err.(*websocket.CloseError)
		if !ok {
			t.Fatalf("expected CloseError, got %T: %v", err, err)
		}
		if closeErr.Code != CloseTimeout {
			t.Errorf("expected close code %d, got %d", CloseTimeout, closeErr.Code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client to detect closed connection")
	}
}
