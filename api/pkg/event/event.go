package event

import (
	"errors"
	"sync"
)

const (
	LeaveEvent  = "leave"
	JoinEvent   = "join"
	StateEvent  = "state"
	InsertEvent = "insert"
	PingEvent   = "ping"
	bufferSize  = 256
)

var (
	ErrClosedBuffer = errors.New("event buffer is closed")
	ErrFullBuffer   = errors.New("event buffer is full")
)

// Represents a message containing type, sender, receiver, trace, error
// message, and a payload. Provides methods for accessing event metadata.
type Event interface {
	// Type of the event.
	Type() string
	// Sets the type of the event.
	SetType(eventType string) Event
	// Entity causing the event.
	Sender() string
	// Sets the entity causing the event.
	SetSender(sender string) Event
	// Trace of the event to identify and keep track of it.
	Trace() string
	// Sets the trace ID of the event.
	SetTrace(trace string) Event
	// Error message if event could not be properly processed.
	Error() string
	// Sets the error message.
	SetError(msg string) Event
	// Data attached ot the event.
	Payload() Payload
	// Sets the payload data.
	SetPayload(payload Payload) Event
	// Creation timestamp of the event.
	Timestamp() int64
	// Sets the creation timestamp.
	SetTimestamp(ts int64) Event
}

type event struct {
	eventType string
	sender    string
	trace     string
	errorMsg  string
	payload   Payload
	timestamp int64
}

var _ Event = &event{}

// Returns an empty event that can be build using setter methods.
func New() *event {
	return &event{}
}

func (e *event) Sender() string {
	return e.sender
}

func (e *event) SetSender(sender string) Event {
	e.sender = sender
	return e
}

func (e *event) Trace() string {
	return e.trace
}

func (e *event) SetTrace(trace string) Event {
	e.trace = trace
	return e
}

func (e *event) Type() string {
	return e.eventType
}

func (e *event) SetType(eventType string) Event {
	e.eventType = eventType
	return e
}

func (e *event) Error() string {
	return e.errorMsg
}

func (e *event) SetError(msg string) Event {
	e.errorMsg = msg
	return e
}

func (e *event) Payload() Payload {
	return e.payload
}

func (e *event) SetPayload(payload Payload) Event {
	e.payload = payload
	return e
}

func (e *event) Timestamp() int64 {
	return e.timestamp
}

func (e *event) SetTimestamp(ts int64) Event {
	e.timestamp = ts
	return e
}

// Represents a buffered channel for sending and receiving events. Provides
// thread-safe methods for event transmission and supports graceful shutdown.
type Buffer interface {
	// Sends an event to the bus. Returns ErrFullBuffer if the buffer is full
	// or ErrClosedBuffer if the buffer has been closed.
	Send(event Event) error
	// Receives an event from the buffer, blocking until one is available.
	// Returns ErrClosedBuffer if the buffer has been closed.
	Receive() (Event, error)
	// Closes the event buffer, preventing further sends and receives.
	Close()
}

type buffer struct {
	events chan Event
	lock   sync.RWMutex
	once   sync.Once
	closed chan struct{}
}

var _ Buffer = &buffer{}

// Returns a new event buffer with a size of 256 events.
func NewBuffer() *buffer {
	return &buffer{
		events: make(chan Event, bufferSize),
		closed: make(chan struct{}),
	}
}

func (b *buffer) Send(event Event) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	select {
	case <-b.closed:
		return ErrClosedBuffer
	default:
	}

	select {
	case b.events <- event:
		return nil
	default:
		return ErrFullBuffer
	}
}

func (b *buffer) Receive() (Event, error) {
	event, ok := <-b.events
	if !ok {
		return nil, ErrClosedBuffer
	}
	return event, nil
}

func (b *buffer) Close() {
	b.once.Do(func() {
		b.lock.Lock()
		close(b.closed)
		close(b.events)
		b.lock.Unlock()
	})
}

// Represents a router that distributes events from a source buffer to multiple
// registered destination buffers.
type Fanout interface {
	// Registers a destination buffer under the specified id for event routing.
	Register(id string, buffer Buffer)
	// Removes the buffer with the specified id from event routing.
	Deregister(id string)
	// Retrieves an event from the source buffer and distributes it to all
	// registered target event buffers. Returns an ErrClosedBuffer if the source
	// buffer is closed.
	Pump() error
}

type fanout struct {
	src    Buffer
	lock   sync.RWMutex
	routes map[string]Buffer
}

var _ Fanout = &fanout{}

// Returns a new fanout router that distributes events from the source buffer
// to registered destinations.
func NewFanout(src Buffer) *fanout {
	return &fanout{
		src:    src,
		routes: make(map[string]Buffer),
	}
}

func (f *fanout) Register(id string, buffer Buffer) {
	f.lock.Lock()
	if old := f.routes[id]; old != nil {
		old.Close()
	}
	f.routes[id] = buffer
	f.lock.Unlock()
}

func (f *fanout) Deregister(id string) {
	f.lock.Lock()
	delete(f.routes, id)
	f.lock.Unlock()
}

func (f *fanout) close() {
	f.lock.RLock()
	defer f.lock.RUnlock()
	for _, buffer := range f.routes {
		buffer.Close()
	}
}

func (f *fanout) dispatch(event Event) {
	f.lock.RLock()
	routes := make(map[string]Buffer, len(f.routes))
	for k, v := range f.routes {
		routes[k] = v
	}
	f.lock.RUnlock()

	// broadcast
	for id, buffer := range routes {
		if err := buffer.Send(event); err == ErrClosedBuffer {
			f.Deregister(id)
		}
	}
}

func (f *fanout) Pump() error {
	event, err := f.src.Receive()
	// stop when source buffer has been closed
	if err == ErrClosedBuffer {
		f.close()
	}
	if err != nil {
		return err
	}
	f.dispatch(event)
	return nil
}
