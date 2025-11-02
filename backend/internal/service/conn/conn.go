package conn

import (
	"encoding/json"
	"log"
	"net/http"
	"sudojo/internal/adapter/database"
	"sudojo/internal/domain/lobby"
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
	db       database.Database
	upgrader websocket.Upgrader
	lock     sync.RWMutex
	workers  map[string]*worker
}

type worker struct {
	lobby   *lobby.Lobby
	clients map[string]*websocket.Conn
}

func New(db database.Database) *service {
	s := &service{
		db:      db,
		workers: make(map[string]*worker),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return s
}

func (s *service) NewWorker(l *lobby.Lobby) {
	w := &worker{lobby: l, clients: make(map[string]*websocket.Conn)}
	go w.writePump()
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/lobbies/{id}/ws": {
			"GET": s.connect,
		},
	}
}

func (w *worker) writePump() {
	for {
		select {
		case event, ok := <-w.lobby.Events:
			if !ok {
				return
			}

			b, err := json.Marshal(event)
			if err != nil {
				log.Printf("ERROR: failed serializing message: %v", err)
				continue
			}

			if len(event.Receiver) > 0 {
				conn := w.clients[event.Receiver]
				if conn == nil {
					continue
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Printf("ERROR: failed sending message: %v", err)
					return
				}
			}
		}
	}
}

func (s *service) connect(w http.ResponseWriter, r *http.Request) {

}
