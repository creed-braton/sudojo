package lobby

import (
	"encoding/json"
	"log"
	"net/http"
	"sudojo/domain/game"
	"sudojo/domain/sudoku"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MsgTypeMove  = "move"
	MsgTypeState = "state"
)

type Inbound struct {
	Type   string `json:"type"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Value  int    `json:"value"`
}

type Outbound struct {
	Initial *sudoku.Sudoku `json:"initial_state,omitempty"`
	Current *sudoku.Sudoku `json:"current_state,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type client struct {
	conn *websocket.Conn
	out  chan *Outbound
}

type lobby struct {
	game    *game.Game
	clients map[*websocket.Conn]*client
	lock    sync.RWMutex
}

func (l *lobby) broadcast(msg *Outbound) {
	l.lock.RLock()
	for _, c := range l.clients {
		c.out <- msg
	}
	l.lock.RUnlock()
}

type service struct {
	upgrader websocket.Upgrader
	lobbies  map[string]*lobby
	lock     sync.RWMutex
}

func New() *service {
	return &service{
		lobbies: make(map[string]*lobby),
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

func (c *client) writePump() {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.out:
			if !ok {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("failed to send connection close message: %v\n", err)
				}
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("failed serializing outbound message: %v\n", err)
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message: %v\n", err)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("failed to send ping: %v\n", err)
			}
		}
	}
}

func (c *client) readPump(lobby *lobby) {
	defer func() {
		close(c.out)
		lobby.lock.Lock()
		delete(lobby.clients, c.conn)
		lobby.lock.Unlock()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseProtocolError,
				websocket.CloseUnsupportedData,
				websocket.CloseNoStatusReceived,
			) {
				log.Printf("websocket closed by client: %v", err)
				break
			}
			log.Printf("failed to read message: %v\n", err)
			break
		}

		if msgType == websocket.TextMessage {
			msg := &Inbound{}

			if err := json.Unmarshal(data, msg); err != nil {
				c.out <- &Outbound{Error: "invalid json format"}
				continue
			}

			switch msg.Type {
			case MsgTypeState:
				c.out <- &Outbound{
					Initial: lobby.game.Initial,
					Current: lobby.game.Current,
				}

			case MsgTypeMove:
				update, err := lobby.game.Move(msg.Row, msg.Column, msg.Value)
				if err != nil {
					c.out <- &Outbound{Error: err.Error()}
				}
				if update {
					lobby.broadcast(&Outbound{Current: lobby.game.Current})
				}

			default:
				c.out <- &Outbound{Error: "invalid message type"}
			}
		}
	}
}

func (s *service) postLobby(w http.ResponseWriter, r *http.Request) {
	l := &lobby{
		game:    game.New(),
		clients: make(map[*websocket.Conn]*client),
	}
	s.lock.Lock()
	s.lobbies[l.game.Id.String()] = l
	s.lock.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(l.game.Id.String()))
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

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	lobby.lock.Lock()
	client := &client{conn: conn, out: make(chan *Outbound, 256)}
	lobby.clients[conn] = client
	lobby.lock.Unlock()

	go client.writePump()
	go client.readPump(lobby)
}
