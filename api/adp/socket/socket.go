package socket

import (
	"errors"
	"fmt"
	"sudojo/adp/serial"
	"sudojo/pkg/event"
	"sudojo/pkg/message"
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
}

type client struct {
	conn    *websocket.Conn
	serial  serial.Serial
	token   string
	buffer  event.Buffer
	msgChan chan message.Message
	leave   func()
	limiter *rate.Limiter
}

var _ Client = &client{}

func NewClient(
	conn *websocket.Conn,
	token string,
	buffer event.Buffer,
	msgChan chan message.Message,
	leave func(),
) *client {
	return &client{
		conn:    conn,
		serial:  serial.New(),
		token:   token,
		buffer:  buffer,
		msgChan: msgChan,
		leave:   leave,
		limiter: rate.NewLimiter(rateLimit, burstLimit),
	}
}

func (c *client) WritePump() error {
	ticker := time.NewTicker(pingInterval * time.Second)
	defer func() {
		ticker.Stop()
		msg := websocket.FormatCloseMessage(
			c.buffer.Reason()+4000, "",
		)
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			msg,
			time.Now().Add(time.Second),
		)
		time.Sleep(time.Second)
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.buffer.Chan():
			if !ok {
				return errors.New("event buffer is closed")
			}

			b, err := c.serial.MarshalEvent(event)
			if err != nil {
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return fmt.Errorf("failed sending message: %w", err)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeDeadline * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.buffer.Close(event.TimeoutReason)
				return fmt.Errorf("failed sending ping: %w", err)
			}
		}
	}
}

func (c *client) ReadPump() error {
	defer func() {
		close(c.msgChan)
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
			event := event.New(event.SystemEvent, time.Now().UTC().UnixNano(), "").
				SetError("rate limit exceeded")
			c.buffer.Send(event)
			continue
		}

		msg, err := c.serial.UnmarshalMsg(b)
		if err != nil {
			event := event.New(event.SystemEvent, time.Now().UTC().UnixNano(), "").
				SetError(err.Error())
			c.buffer.Send(event)
		}

		c.msgChan <- msg
	}
}
