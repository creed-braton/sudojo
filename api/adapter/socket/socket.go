package socket

import (
	"encoding/json"
	"errors"
	"fmt"
	"sudojo/pkg/event"
	"sudojo/pkg/manager"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	readDeadline  = 60
	writeDeadline = 10
	pingInterval  = 20
	rateLimit     = 4
	burstLimit    = 7
)

type Client interface {
	WritePump() error
	ReadPump() error
}

type client struct {
	player  manager.Player
	conn    *websocket.Conn
	write   chan []byte
	limiter *rate.Limiter
}

var _ Client = &client{}

func NewClient(player manager.Player, conn *websocket.Conn) *client {
	return &client{
		player:  player,
		conn:    conn,
		write:   make(chan []byte, 256),
		limiter: rate.NewLimiter(rateLimit, burstLimit),
	}
}

func (c *client) WritePump() error {
	ticker := time.NewTicker(pingInterval * time.Second)
	defer func() {
		ticker.Stop()
		msg := websocket.FormatCloseMessage(
			websocket.CloseNormalClosure, "",
		)
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			msg,
			time.Now().Add(time.Second),
		)
		time.Sleep(time.Second)
		c.conn.Close()
	}()

	ch := make(chan event.Event, 256)
	go func() {
		for {
			event, err := c.player.Receive()
			if err != nil {
				close(ch)
				return
			}
			ch <- event
		}
	}()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return errors.New("event channel has been closed")
			}

			msg := newMessage(event)
			b, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return fmt.Errorf("failed sending message: %w", err)
			}

		case b, ok := <-c.write:
			if !ok {
				return errors.New("write buffer has been closed")
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return fmt.Errorf("failed sending message: %w", err)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("failed sending ping: %w", err)
			}
		}
	}
}

func (c *client) ReadPump() error {
	defer func() {
		close(c.write)
	}()

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(readDeadline * time.Second))
		return nil
	})

	for {
		c.conn.SetReadDeadline(time.Now().Add(readDeadline * time.Second))
		t, b, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed receiving message: %w", err)
		}

		if t != websocket.TextMessage {
			continue
		}

		if !c.limiter.Allow() {
			select {
			case c.write <- rateLimitMsg:
			default:
			}
			continue
		}

		msg := &Message{}
		if err := json.Unmarshal(b, msg); err != nil || len(msg.Type) < 1 {
			select {
			case c.write <- invalidMsg:
			default:
			}
			continue
		}

		if msg.Type == event.InsertEvent {
			if msg.Row != nil && msg.Column != nil && msg.Value != nil {
				if err := c.player.Insert(*msg.Row, *msg.Column, *msg.Value, ""); err != nil {
					return err
				}
				continue
			}
		} else if msg.Type == event.PingEvent {
			if msg.Row != nil && msg.Column != nil {
				if err := c.player.Ping(*msg.Row, *msg.Column, ""); err != nil {
					return err
				}
				continue
			}
		} else if msg.Type == event.StateEvent {
			if err := c.player.State(""); err != nil {
				return err
			}
			continue
		}

		select {
		case c.write <- invalidMsg:
		default:
		}
	}
}
