package socket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func setupServer() (string, func()) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		client := NewClient(conn)
		go client.ReadPump()
		go client.WritePump()
		go func() {
			for {
				msg, err := client.Receive()
				if err != nil {
					client.Close()
					return
				}
				client.Send(msg)
			}
		}()
	}))

	return "ws" + server.URL[len("http"):], server.Close
}

func setupClient() (string, func()) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		go func() {
			defer conn.Close()
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if string(msg) == "close" {
					return
				}
			}
		}()
	}))

	return "ws" + server.URL[len("http"):], server.Close
}

func TestConnection(t *testing.T) {
	url, teardown := setupServer()
	defer teardown()

	numMsg := burstLimit
	numClient := 128
	wg := sync.WaitGroup{}
	wg.Add(numClient)

	for range numClient {
		go func() {
			defer wg.Done()

			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				t.Errorf("dial error: %v", err)
			}
			defer conn.Close()

			for range numMsg {
				want := uuid.NewString()
				if err := conn.WriteMessage(websocket.TextMessage, []byte(want)); err != nil {
					t.Errorf("write message error: %v", err)
				}
				_, msg, err := conn.ReadMessage()
				if err != nil {
					t.Errorf("read message error: %v", err)
				}
				got := string(msg)
				if got != want {
					t.Errorf("expected: '%s', got: '%s'", want, got)
				}
			}
		}()
	}

	wg.Wait()
}

func TestTermination(t *testing.T) {
	url, teardown := setupClient()
	defer teardown()

	t.Run("client side connection close", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("dial error: %v", err)
			return
		}
		client := NewClient(conn)

		wg := sync.WaitGroup{}
		wg.Add(3)
		go func() {
			client.ReadPump()
			wg.Done()
		}()
		go func() {
			client.WritePump()
			wg.Done()
		}()
		go func() {
			for {
				if _, err := client.Receive(); err != nil {
					client.Close()
					break
				}
			}
			wg.Done()
		}()

		client.Send([]byte("close"))
		wg.Wait()
	})

	t.Run("server side connection close", func(t *testing.T) {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("dial error: %v", err)
			return
		}
		client := NewClient(conn)

		wg := sync.WaitGroup{}
		wg.Add(3)
		go func() {
			client.ReadPump()
			wg.Done()
		}()
		go func() {
			client.WritePump()
			wg.Done()
		}()
		go func() {
			for {
				if _, err := client.Receive(); err != nil {
					break
				}
			}
			wg.Done()
		}()

		client.Close()
		wg.Wait()
	})
}

func TestRateLimit(t *testing.T) {
	url, teardown := setupServer()
	defer teardown()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Errorf("dial error: %v", err)
	}
	defer conn.Close()

	overflow := 3
	total := burstLimit + overflow
	limit := 0

	for i := range total {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d", i))); err != nil {
			t.Fatalf("write error: %v", err)
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read error: %v", err)
		}

		if string(msg) == string(rateLimitMsg) {
			limit++
		}
	}

	if limit != overflow {
		t.Errorf("expected %d messages to be rate limited, got %d", overflow, limit)
	}
}
