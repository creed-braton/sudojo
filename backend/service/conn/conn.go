package conn

import (
	"fmt"
	"log"
	"net/http"
	"sudojo/domain/lobby"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	once sync.Once
	done chan struct{}
}

func (c *client) close() {
	c.once.Do(func() {
		c.conn.Close()
		close(c.done)
	})
}

type service struct {
	upgrader websocket.Upgrader
	lobbies  map[string]*lobby.Lobby
	lock     sync.RWMutex
}

func New() *service {
	return &service{
		lobbies: make(map[string]*lobby.Lobby),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/lobbies": {
			"POST": s.postLobby,
		},
		"/lobbies/{id}": {
			"GET": s.getLobby,
		},
	}
}

func (s *service) writePump(player *lobby.Player, client *client) {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-client.done:
			return
		case msg := <-player.Out:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("failed to send message: %v", err)
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("failed to send ping: %v", err)
			}
		}
	}
}

func (s *service) readPump(lobby *lobby.Lobby, player *lobby.Player, client *client) {
	defer func() {
		client.close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		t, msg, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		if t == websocket.TextMessage {
			err := lobby.Process(msg, player)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}

func (s *service) getLobby(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.lock.RLock()
	lobby, exists := s.lobbies[id]
	s.lock.RUnlock()
	if !exists {
		http.Error(w, "lobby not found", 404)
		return
	}

	var token string
	cookie, err := r.Cookie("session_token")
	if err == nil {
		token = cookie.Value
	}

	player := lobby.Player(token)
	if player == nil {
		player, err = lobby.Join("")
		if err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
	}

	conn, err := s.upgrader.Upgrade(w, r, http.Header{
		"Set-Cookie": {(&http.Cookie{
			Name:     "session_token",
			Value:    player.Token,
			Path:     fmt.Sprintf("/api/lobbies/%s", id),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400, // 1 day, experimental
		}).String()},
	})
	if err != nil {
		log.Printf("failed to create connection: %v", err)
		return
	}
	client := &client{conn: conn, done: make(chan struct{})}

	go s.writePump(player, client)
	go s.readPump(lobby, player, client)
}

func (s *service) postLobby(w http.ResponseWriter, r *http.Request) {
	l := lobby.New()
	s.lock.Lock()
	s.lobbies[l.Id] = l
	s.lock.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(l.Id))
}
