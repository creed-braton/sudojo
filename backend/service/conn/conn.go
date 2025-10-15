package conn

import (
	"log"
	"net/http"
	"sudojo/domain/lobby"
	"sudojo/service/data"
	"sync"
	"time"

	"github.com/google/uuid"
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
	clients  map[string]*client
	data     *data.Service
	lock     sync.RWMutex
}

func New(data *data.Service) *service {
	s := &service{
		data:    data,
		clients: make(map[string]*client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	go func() {
		ticker := time.NewTicker(45 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			s.lock.Lock()
			for token, client := range s.clients {
				select {
				case <-client.done:
					delete(s.clients, token)
				default:
					continue
				}
			}
			s.lock.Unlock()
		}
	}()

	return s
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/lobbies/{id}/ws": {
			"GET": s.connect,
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
				log.Printf("ERROR: failed sending message: %v", err)
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ERROR: failed sending ping: %v", err)
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
				log.Printf("ERROR: %v", err)
			}
		}
	}
}

func (s *service) connect(w http.ResponseWriter, r *http.Request) {
	input := r.PathValue("id")
	id, err := uuid.Parse(input)
	if err != nil {
		http.Error(w, "invalid lobby id format", 400)
		return
	}

	token := r.URL.Query().Get("token")
	if len(token) != 32 {
		http.Error(w, "invalid token format", 400)
		return
	}

	lobby, err := s.data.Lobby(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if lobby == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	player := lobby.Player(token)
	if player == nil {
		http.Error(w, "player not found", 404)
		return
	}
	err = lobby.Join(player)
	if err != nil {
		log.Printf("WARNING: failed rejoining lobby: %v", err)
		http.Error(w, "internal server error", 500)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: failed creating connection: %v", err)
		http.Error(w, "internal server error", 500)
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
