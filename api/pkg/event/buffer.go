package event

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type BufferClosedError struct {
	reason int
}

func (e *BufferClosedError) Error() string {
	return fmt.Sprintf("event buffer is closed, reason: %d", e.reason)
}

func (e *BufferClosedError) Reason() int {
	return e.reason
}

var (
	ErrBufferFull = errors.New("event buffer is full")
)

// Buffered channel which is safe for concurrent use and can be closed without
// causing other senders to panic. Close reason can be passed and will be propagated
// through the BufferClosedError to receivers and senders.
type Buffer interface {
	// Puts non-blockingly the event into the channel. Also doesn't block if the channel
	// is full but will return ErrBufferFull. Returns a BufferClosedError if the channel
	// is closed.
	Send(event Event) error
	// Gets an event for the channel and blocks if the channel is empty until one event
	// arrives. Returns a BufferClosedError if the channel is closed.
	Receive() (Event, error)
	// Closes the buffered channel with provided reason. All sender and receiver will
	// get a BufferClosedError which contains the provided reason.
	Close(reason int)
}

type buffer struct {
	events chan Event
	lock   sync.RWMutex
	once   sync.Once
	closed chan struct{}
	reason atomic.Int32
}

var _ Buffer = &buffer{}

// Creates a buffered channel with the provided size.
func NewBuffer(size int) *buffer {
	return &buffer{
		events: make(chan Event, size),
		closed: make(chan struct{}),
	}
}

func (b *buffer) Send(event Event) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	select {
	case <-b.closed:
		return &BufferClosedError{reason: int(b.reason.Load())}
	default:
	}

	select {
	case b.events <- event:
		return nil
	default:
		return ErrBufferFull
	}
}

func (b *buffer) Receive() (Event, error) {
	event, ok := <-b.events
	if !ok {
		return nil, &BufferClosedError{reason: int(b.reason.Load())}
	}
	return event, nil
}

func (b *buffer) Close(reason int) {
	b.once.Do(func() {
		b.lock.Lock()
		b.reason.Store(int32(reason))
		close(b.closed)
		close(b.events)
		b.lock.Unlock()
	})
}
