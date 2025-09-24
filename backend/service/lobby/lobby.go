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

type client struct {
	conn *websocket.Conn
	lock sync.Mutex
}

func (c *client) send(r *Response) {
	res, _ := json.Marshal(r)
	c.lock.Lock()
	c.conn.WriteMessage(websocket.TextMessage, res)
	c.lock.Unlock()
}

type lobby struct {
	game    *game.Game
	clients map[*websocket.Conn]*client
	lock    sync.RWMutex
}

func (l *lobby) broadcast(r *Response) {
	l.lock.RLock()
	for _, c := range l.clients {
		c.send(r)
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
	client := &client{conn: conn}
	lobby.clients[conn] = client
	lobby.lock.Unlock()

	go s.handleConnection(client, lobby)
}

const (
	MsgTypeMove  = "move"
	MsgTypeState = "state"
)

type Request struct {
	Type   string `json:"type"`
	Row    int    `json:"row"`
	Column int    `json:"column"`
	Value  int    `json:"value"`
}

type Response struct {
	Initial *sudoku.Sudoku `json:"initial_state,omitempty"`
	Current *sudoku.Sudoku `json:"current_state,omitempty"`
	Error   string         `json:"error,omitempty"`
}

func (s *service) handleConnection(client *client, lobby *lobby) {
	defer func() {
		log.Println("closing client connection")
		client.conn.Close()
		lobby.lock.Lock()
		delete(lobby.clients, client.conn)
		lobby.lock.Unlock()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// ping go routine
	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer func() {
			ticker.Stop()
			client.conn.Close()
		}()
		for {
			select {
			case <-ticker.C:
				client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("couldn't send a ping: %v\n", err)
					return
				}
			}
		}
	}()

	for {
		msgType, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err.Error())
			break
		}

		if msgType == websocket.TextMessage {
			msg := &Request{}

			if err := json.Unmarshal(message, msg); err != nil {
				client.send(&Response{Error: "invalid json format"})
				continue
			}

			switch msg.Type {
			case MsgTypeState:
				s.handleStateMsg(lobby, client)
			case MsgTypeMove:
				s.handleMoveMsg(msg, lobby, client)
			default:
				client.send(&Response{Error: "invalid message type"})
			}
		}
	}
}

func (s *service) handleStateMsg(lobby *lobby, client *client) {
	client.send(&Response{
		Initial: lobby.game.Initial,
		Current: lobby.game.Current,
	})
}

func (s *service) handleMoveMsg(msg *Request, lobby *lobby, client *client) {
	update, err := lobby.game.Move(msg.Row, msg.Column, msg.Value)
	if err != nil {
		client.send(&Response{Error: err.Error()})
		return
	}
	if update {
		lobby.broadcast(&Response{Current: lobby.game.Current})
	}
}
