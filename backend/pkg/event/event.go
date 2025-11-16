package event

import (
	"errors"
	"sync"
)

const (
	LeaveEvent   = "leave"
	JoinEvent    = "join"
	StateEvent   = "state"
	InsertEvent  = "insert"
	PingEvent    = "ping"
	eventBusSize = 256
)

var (
	ErrClosedBus = errors.New("event bus is closed")
	ErrFullBus   = errors.New("event bus is full")
)

// Represents a message containing type, sender, receiver, trace id, error
// message, and a payload. Provides methods for accessing event metadata.
type Event interface {
	// Type of the event.
	Type() string
	// Entity causing the event.
	Sender() string
	// Trace id of the event to keep track of it.
	Trace() string
	// Error message if event could not be properly processed.
	Error() string
	// Data attached ot the event.
	Payload() Payload
}

type event struct {
	eventType string
	sender    string
	trace     string
	errorMsg  string
	payload   Payload
}

var _ Event = &event{}

// Returns a new event with the provided type, sender, trace id, error message, and payload.
func New(eventType, sender, trace, errorMsg string, payload Payload) *event {
	return &event{
		sender:    sender,
		trace:     trace,
		eventType: eventType,
		errorMsg:  errorMsg,
		payload:   payload,
	}
}

func (e *event) Sender() string {
	return e.sender
}

func (e *event) Trace() string {
	return e.trace
}

func (e *event) Type() string {
	return e.eventType
}

func (e *event) Error() string {
	return e.errorMsg
}

func (e *event) Payload() Payload {
	return e.payload
}

// Represents a channel for sending and receiving events. Provides thread-safe
// methods for event transmission and supports graceful shutdown.
type EventBus interface {
	// Sends an event to the bus. Returns ErrFullBus if the buffer is full
	// or ErrClosedBus if the bus has been closed.
	Send(event Event) error
	// Receives an event from the bus, blocking until one is available.
	// Returns ErrClosedBus if the bus has been closed.
	Receive() (Event, error)
	// Closes the event bus, preventing further sends and receives.
	Close()
}

type eventBus struct {
	events chan Event
	lock   sync.RWMutex
	once   sync.Once
	closed chan struct{}
}

var _ EventBus = &eventBus{}

// Returns a new event bus with a buffer size of 256 events.
func NewEventBus() *eventBus {
	return &eventBus{
		events: make(chan Event, eventBusSize),
		closed: make(chan struct{}),
	}
}

func (b *eventBus) Send(event Event) error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	select {
	case <-b.closed:
		return ErrClosedBus
	default:
	}

	select {
	case b.events <- event:
		return nil
	default:
		return ErrFullBus
	}
}

func (b *eventBus) Receive() (Event, error) {
	event, ok := <-b.events
	if !ok {
		return nil, ErrClosedBus
	}
	return event, nil
}

func (b *eventBus) Close() {
	b.once.Do(func() {
		b.lock.Lock()
		close(b.closed)
		close(b.events)
		b.lock.Unlock()
	})
}

// Represents a router that distributes events from a source bus to multiple
// registered destination buses.
type Fanout interface {
	// Registers a destination bus under the specified id for event routing.
	Register(id string, bus EventBus)
	// Removes the bus with the specified id from event routing.
	Deregister(id string)
	// Retrieves an event from the source bus and distributes it to all registered
	// target event buses. Returns an ErrClosedBus if the source bus is closed.
	Pump() error
}

type fanout struct {
	src    EventBus
	lock   sync.RWMutex
	routes map[string]EventBus
}

var _ Fanout = &fanout{}

// Returns a new fanout router that distributes events from the source bus to
// registered destinations.
func NewFanout(src EventBus) *fanout {
	return &fanout{
		src:    src,
		routes: make(map[string]EventBus),
	}
}

func (f *fanout) Register(id string, bus EventBus) {
	f.lock.Lock()
	if old := f.routes[id]; old != nil {
		old.Close()
	}
	f.routes[id] = bus
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
	for _, bus := range f.routes {
		bus.Close()
	}
}

func (f *fanout) dispatch(event Event) {
	f.lock.RLock()
	routes := make(map[string]EventBus, len(f.routes))
	for k, v := range f.routes {
		routes[k] = v
	}
	f.lock.RUnlock()

	// broadcast
	for id, bus := range routes {
		if err := bus.Send(event); err == ErrClosedBus {
			f.Deregister(id)
		}
	}
}

func (f *fanout) Pump() error {
	event, err := f.src.Receive()
	// stop when source bus has been closed
	if err == ErrClosedBus {
		f.close()
	}
	if err != nil {
		return err
	}
	f.dispatch(event)
	return nil
}
