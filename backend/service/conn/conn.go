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
	token string
	conn  *websocket.Conn
	once  sync.Once
	done  chan struct{}
}

func (c *client) close() {
	c.once.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

type service struct {
	upgrader websocket.Upgrader
	lobbies  map[string]*lobby.Lobby
	clients  map[string]*client
	lock     sync.RWMutex
}

func (s *service) cleaner() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.lock.Lock()
		// clean up idle clients
		for token, client := range s.clients {
			select {
			case <-client.done:
				delete(s.clients, token)
			default:
				continue
			}
		}

		// clean up idle lobbies
		for id, lobby := range s.lobbies {
			if idle := lobby.Idle(); idle {
				delete(s.lobbies, id)
			}
		}
		s.lock.Unlock()
	}
}

func New() *service {
	s := &service{
		lobbies: make(map[string]*lobby.Lobby),
		clients: make(map[string]*client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	go s.cleaner()
	return s
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
		client.close()
		ticker.Stop()
	}()

	for {
		select {
		case <-client.done:
			return

		case msg, ok := <-player.Out:
			if !ok {
				return
			}

			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("failed to send message: %v", err)
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("failed to send ping: %v", err)
				return
			}
		}
	}
}

func (s *service) readPump(lobby *lobby.Lobby, player *lobby.Player, client *client) {
	defer func() {
		client.close()
		lobby.Leave(player)
	}()

	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		t, msg, err := client.conn.ReadMessage() // blocking read, reason why we don't need to check client.done
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
	} else {
		err := lobby.Rejoin(player)
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
	client := &client{token: player.Token, conn: conn, done: make(chan struct{})}

	s.lock.Lock()
	// clean up old client if there is any
	old := s.clients[player.Token]
	if old != nil {
		old.close()
	}
	// register new client
	s.clients[player.Token] = client
	s.lock.Unlock()

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
