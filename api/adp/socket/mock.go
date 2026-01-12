package socket

import (
	"sync"

	"github.com/google/uuid"
)

type mock struct {
	id      string
	in, out chan *Message
	closed  chan struct{}
	once    sync.Once

	// Exposed for test assertions
	CloseCode int
	CloseMsg  string
}

var _ Socket = &mock{}

func NewMock(in, out chan *Message) *mock {
	return &mock{
		id:     uuid.NewString(),
		in:     in,
		out:    out,
		closed: make(chan struct{}),
	}
}

func (s *mock) Id() string {
	return s.id
}

func (s *mock) Close(code int, msg string) {
	s.once.Do(func() {
		s.CloseCode = code
		s.CloseMsg = msg
		close(s.closed)
	})
}

func (s *mock) Closed() <-chan struct{} {
	return s.closed
}

func (s *mock) Listen() error {
	<-s.closed
	return nil
}

func (s *mock) Send(msg *Message) error {
	select {
	case <-s.closed:
		return ErrClosed
	case s.out <- msg:
		return nil
	default:
		return ErrBufferFull
	}
}

func (s *mock) Receive() (*Message, error) {
	select {
	case <-s.closed:
		return nil, ErrClosed
	case msg, ok := <-s.in:
		if !ok {
			s.Close(CloseTimeout, "client closed")
			return nil, ErrClosed
		}
		return msg, nil
	}
}
