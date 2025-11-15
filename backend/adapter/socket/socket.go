package socket

import (
	"encoding/json"
	"errors"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/sudoku"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	readDeadline  = 60
	writeDeadline = 10
	pingInterval  = 20
	rateLimit     = 2
	burstLimit    = 4
)

var (
	rateLimitMsg = []byte(`{"type":"system","error":"rate limit exceeded"}`)
	invalidMsg   = []byte(`{"type":"system","error":"invalid message format"}`)
)

type message struct {
	Type     string                `json:"type"`
	Trace    string                `json:"trace_id,omitempty"`
	Error    string                `json:"error,omitempty"`
	Current  sudoku.Sudoku         `json:"current_state,omitempty"`
	Initial  sudoku.Sudoku         `json:"initial_state,omitempty"`
	Conflict string                `json:"conflict,omitempty"`
	Row      *int                  `json:"row,omitempty"`
	Column   *int                  `json:"column,omitempty"`
	Value    *int                  `json:"value,omitempty"`
	Players  []*event.PlayerStatus `json:"players,omitempty"`
}

// Represents a WebSocket client connection. Provides methods for sending and
// receiving messages, as well as pumping messages through the internal channels.
type Client interface {
	// Returns the UUID of the client.
	Id() string
	// Closes the client connection and cleans up resources.
	Close()
	// Writes continuously messages from internal channels to WebSocket connection.
	// Also handles periodic pings to keep the connection alive. Closes read channel
	// if read deadline runs out.
	WritePump() error
	// Reads continuously messages from WebSocket connection and writes them into
	// internal channel. Enforces rate limiting and updates the read deadline
	// based on pong messages.
	ReadPump() error
	// Queues non-blockingly event into internal channel to be sent to the client.
	// Drops messages when internal channel is full. Returns error if event could
	// not be serialized into bytes.
	Send(e event.Event) error
	// Retrieves blockingly next message from the client. Returns error if read
	// channel has been closed.
	Receive() (*message, error)
}

type client struct {
	id      string
	read    chan *message
	write   chan []byte
	cross   chan []byte
	conn    *websocket.Conn
	once    sync.Once
	limiter *rate.Limiter
}

var _ Client = &client{}

// Returns a new Client instance with a UUID. It initializes internal channels for
// reading, writing, and cross-messages, sets up a rate limiter, and associates
// the provided WebSocket connection.
func NewClient(conn *websocket.Conn) *client {
	return &client{
		id:      uuid.NewString(),
		read:    make(chan *message, 256),
		write:   make(chan []byte, 256),
		cross:   make(chan []byte, 256),
		conn:    conn,
		limiter: rate.NewLimiter(rateLimit, burstLimit),
	}
}

func (c *client) Id() string {
	return c.id
}

func (c *client) Close() {
	c.once.Do(func() {
		close(c.write)
		c.conn.Close()
	})
}

func (c *client) WritePump() error {
	ticker := time.NewTicker(pingInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.write:
			if !ok {
				return errors.New("write buffer has been closed")
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return fmt.Errorf("failed sending message: %v", err)
			}

		case msg, ok := <-c.cross:
			if !ok {
				return errors.New("cross buffer has been closed")
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return fmt.Errorf("failed sending message: %v", err)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("failed sending ping: %v", err)
			}
		}
	}
}

func (c *client) ReadPump() error {
	defer func() {
		close(c.read)
		close(c.cross)
	}()

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(readDeadline * time.Second))
		return nil
	})

	for {
		c.conn.SetReadDeadline(time.Now().Add(readDeadline * time.Second))
		t, b, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed receiving message: %v", err)
		}

		if t != websocket.TextMessage {
			continue
		}

		if !c.limiter.Allow() {
			select {
			case c.cross <- rateLimitMsg:
			default:
			}
			continue
		}

		msg := &message{}
		if err := json.Unmarshal(b, msg); err != nil || len(msg.Type) < 1 {
			select {
			case c.cross <- invalidMsg:
			default:
			}
			continue
		}

		if msg.Type == event.InsertEvent {
			if msg.Row == nil || msg.Column == nil || msg.Value == nil {
				select {
				case c.cross <- invalidMsg:
				default:
				}
				continue
			}
		} else if msg.Type == event.PingEvent {
			if msg.Row == nil || msg.Column == nil {
				select {
				case c.cross <- invalidMsg:
				default:
				}
				continue
			}
		} else if msg.Type != event.StateEvent {
			select {
			case c.cross <- invalidMsg:
			default:
			}
			continue
		}

		c.read <- msg
	}
}

func (c *client) Send(e event.Event) error {
	msg := &message{
		Type:     e.Type(),
		Trace:    e.Trace(),
		Error:    e.Error(),
		Current:  e.Payload().Current(),
		Initial:  e.Payload().Initial(),
		Conflict: e.Payload().Conflict(),
		Row:      e.Payload().Row(),
		Column:   e.Payload().Column(),
		Value:    e.Payload().Value(),
		Players:  e.Payload().Players(),
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.write <- b:
	default:
	}

	return nil
}

func (c *client) Receive() (*message, error) {
	msg, ok := <-c.read
	if !ok {
		return nil, errors.New("read buffer has been closed")
	}
	return msg, nil
}
