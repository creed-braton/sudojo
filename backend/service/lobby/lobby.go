package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type Client struct {
	token  string
	conn   *websocket.Conn
	out    chan *Outbound
	closed sync.Once
}

func (c *Client) Close() {
	c.closed.Do(func() {
		c.conn.Close()
	})
}

type lobby struct {
	game    *game.Game
	clients map[string]*Client
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

func (s *service) writePump(client *Client) {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
		client.Close()
	}()

	for {
		select {
		case msg := <-client.out:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("failed serializing outbound message: %v", err)
				continue
			}
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
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

func (s *service) readPump(lobby *lobby, client *Client) {
	defer func() {
		client.Close()
	}()

	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		msgType, data, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.TextMessage {
			msg := &Inbound{}

			if err := json.Unmarshal(data, msg); err != nil {
				client.out <- &Outbound{Error: "invalid json format"}
				continue
			}

			switch msg.Type {
			case MsgTypeState:
				client.out <- &Outbound{
					Initial: lobby.game.Initial,
					Current: lobby.game.Current,
				}

			case MsgTypeMove:
				update, err := lobby.game.Move(msg.Row, msg.Column, msg.Value)
				if err != nil {
					client.out <- &Outbound{Error: err.Error()}
				}

				if update {
					lobby.broadcast(&Outbound{Current: lobby.game.Current})
				}

				if lobby.game.Finished != nil {
					lobby.lock.Lock()
					for _, c := range lobby.clients {
						c.Close()
					}
					lobby.lock.Unlock()

					s.lock.Lock()
					delete(s.lobbies, lobby.game.Id.String())
					s.lock.Unlock()
				}

			default:
				client.out <- &Outbound{Error: "invalid message type"}
			}
		}
	}
}

func (s *service) postLobby(w http.ResponseWriter, r *http.Request) {
	l := &lobby{
		game:    game.New(),
		clients: make(map[string]*Client),
	}
	s.lock.Lock()
	s.lobbies[l.game.Id.String()] = l
	s.lock.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(l.game.Id.String()))
}

func newToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
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
	lobby.lock.Lock()
	defer func() {
		lobby.lock.Unlock()
	}()

	client := lobby.clients[token]
	if client == nil {
		token = newToken()
		client = &Client{token: token, out: make(chan *Outbound, 256)}
	} else {
		client.Close()
	}

	conn, err := s.upgrader.Upgrade(w, r, http.Header{
		"Set-Cookie": {(&http.Cookie{
			Name:     "session_token",
			Value:    token,
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
	client.conn = conn

	lobby.clients[token] = client

	go s.writePump(client)
	go s.readPump(lobby, client)
}
