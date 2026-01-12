package socket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	CloseTimeout  = 4001
	CloseTakeover = 4002
	CloseFinished = 4003
	CloseNotFound = 4004
	CloseIdle     = 4005
	CloseStale    = 4006
)

var (
	ErrClosed     = errors.New("connection is closed")
	ErrBufferFull = errors.New("buffer is full")
)

type Socket interface {
	Id() string
	Close(code int, msg string)
	Listen() error
	Send(msg *Message) error
	Receive() (*Message, error)
}

type socket struct {
	id      string
	conn    *websocket.Conn
	in      chan *Message
	out     chan *Message
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
	closed  chan struct{}
	timeout time.Duration
	ping    time.Duration
}

var _ Socket = &socket{}

func New(conn *websocket.Conn, timeout, ping time.Duration) *socket {
	ctx, cancel := context.WithCancel(context.Background())

	return &socket{
		id:      uuid.NewString(),
		conn:    conn,
		in:      make(chan *Message, 32),
		out:     make(chan *Message, 32),
		ctx:     ctx,
		cancel:  cancel,
		closed:  make(chan struct{}),
		timeout: timeout,
		ping:    ping,
	}
}

func (s *socket) Id() string {
	return s.id
}

func (s *socket) Close(code int, msg string) {
	s.once.Do(func() {
		close(s.closed) // block sends first

		_ = s.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				code, msg,
			),
			time.Now().Add(time.Second),
		)
		time.Sleep(time.Second)

		s.cancel()
		s.conn.Close()
	})
}

func (s *socket) writePump() error {
	ticker := time.NewTicker(s.ping * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()

		case msg := <-s.out:
			b, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return fmt.Errorf("failed sending message: %w", err)
			}

		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.Close(websocket.CloseNormalClosure, "connection broken")
				return fmt.Errorf("failed sending ping: %w", err)
			}
		}
	}
}

func (s *socket) readPump() error {
	defer func() {
		close(s.in)
	}()

	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(s.timeout * time.Second))
		return nil
	})

	s.conn.SetReadDeadline(time.Now().Add(s.timeout * time.Second))

	for {
		t, b, err := s.conn.ReadMessage()
		if err != nil {
			s.Close(CloseTimeout, "heartbeat timeout")
			return err
		}

		if t != websocket.TextMessage {
			continue
		}

		msg := &Message{}
		if err := json.Unmarshal(b, msg); err != nil || len(msg.Type) < 1 {
			continue
		}

		s.in <- msg
	}
}

func (s *socket) Listen() error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.writePump()
	}()

	err := s.readPump()
	s.wg.Wait()
	return err
}

func (s *socket) Send(msg *Message) error {
	select {
	case <-s.closed:
		return ErrClosed
	case s.out <- msg:
		return nil
	default:
		return ErrBufferFull
	}
}

func (s *socket) Receive() (*Message, error) {
	msg, ok := <-s.in
	if !ok {
		return nil, ErrClosed
	}
	return msg, nil
}
