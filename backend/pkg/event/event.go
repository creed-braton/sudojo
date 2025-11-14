package event

import (
	"errors"
	"sync"
)

const (
	SystemEvent = "system"
	LeaveEvent  = "leave"
	JoinEvent   = "join"
	StateEvent  = "state"
	InsertEvent = "insert"
	PingEvent   = "ping"
	bufferSize  = 256
)

var (
	ErrClosedChan = errors.New("event channel is closed")
)

// Represents arbitrary data carried within an event.
type Payload interface {
	// Serializes the data struct into bytes.
	Marshal() ([]byte, error)
}

// Represents a message containing type, sender, receiver, trace ID,
// error message, broadcast flag and a payload.
type Event interface {
	// Type of the event.
	Type() string
	// Entity causing the event.
	Sender() string
	// Trace id of the event to keep track of it.
	Trace() string
	// Error message if event could not be properly processed.
	Error() string
	// Flag wether event should be broadcasted to all consumers.
	Broadcast() bool
	// Data attached ot the event.
	Payload() Payload
}

type event struct {
	eventType string
	sender    string
	trace     string
	errorMsg  string
	broadcast bool
	payload   Payload
}

var _ Event = &event{}

// Returns a new event with the provided properties.
func New(
	eventType, sender, trace, errorMsg string,
	broadcast bool, payload Payload,
) *event {
	return &event{
		sender:    sender,
		trace:     trace,
		eventType: eventType,
		errorMsg:  errorMsg,
		broadcast: broadcast,
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

func (e *event) Broadcast() bool {
	return e.broadcast
}

func (e *event) Payload() Payload {
	return e.payload
}

// Represents channel for events with non-blocking and blocking communication operations.
type EventChan interface {
	// Queues event non-blockingly in the channel. Drops event if channel is full.
	Send(e Event)
	// Retrieves event non-blockingly from the channel. Returns ErrClosedChan if
	// channel has been closed.
	NonBlockRecv() (Event, error)
	// Retrieves event blockingly from the channel. Returns ErrClosedChan if
	// channel has been closed.
	Receive() (Event, error)
	// Closes the channel.
	Close()
}

type eventChan struct {
	ch   chan Event
	once sync.Once
}

var _ EventChan = &eventChan{}

func NewEventChan() *eventChan {
	return &eventChan{ch: make(chan Event, bufferSize)}
}

func (ch *eventChan) Send(e Event) {
	select {
	case ch.ch <- e:
	default:
	}
}

func (ch *eventChan) NonBlockRecv() (Event, error) {
	select {
	case e, ok := <-ch.ch:
		if !ok {
			return nil, ErrClosedChan
		}
		return e, nil
	default:
	}
	return nil, nil
}

func (ch *eventChan) Receive() (Event, error) {
	e, ok := <-ch.ch
	if !ok {
		return nil, ErrClosedChan
	}
	return e, nil
}

func (ch *eventChan) Close() {
	ch.once.Do(func() {
		close(ch.ch)
	})
}

// Represents a communication hub connecting event publishers and subscribers.
type EventBus interface {
	// Registers a new publisher and subscriber pair under the given identifier.
	// If present closes old subscriber channel under the identifier.
	Register(id string, pub, sub EventChan)
	// Deregisters the publisher and subscriber associated with the given identifier.
	// Also closes subscriber channel.
	Deregister(id string)
	// Pumps events from all publishers and dispatches them to corresponding subscribers.
	Pump()
	// Closes the bus and all registered subscribers.
	Close()
}

type eventBus struct {
	publisher  map[string]EventChan
	subscriber map[string]EventChan
	lock       sync.Mutex
	once       sync.Once
}

var _ EventBus = &eventBus{}

func NewEventBus() *eventBus {
	return &eventBus{
		publisher:  make(map[string]EventChan),
		subscriber: make(map[string]EventChan),
	}
}

func (b *eventBus) Register(id string, pub, sub EventChan) {
	b.lock.Lock()
	defer b.lock.Unlock()

	delete(b.publisher, id)
	if s, exist := b.subscriber[id]; exist {
		delete(b.subscriber, id)
		s.Close()
	}

	b.publisher[id] = pub
	b.subscriber[id] = sub
}

func (b *eventBus) Deregister(id string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if s, exist := b.subscriber[id]; exist {
		delete(b.subscriber, id)
		s.Close()
	}
}

func (b *eventBus) dispatch(e Event) {
	if e.Broadcast() {
		for _, s := range b.subscriber {
			s.Send(e)
		}
	} else {
		if s, exist := b.subscriber[e.Sender()]; exist {
			s.Send(e)
		}
	}
}

func (b *eventBus) Pump() {
	b.lock.Lock()
	defer b.lock.Unlock()

	for id, p := range b.publisher {
		e, err := p.NonBlockRecv()
		if err != nil {
			delete(b.publisher, id)
			if sub, exist := b.subscriber[id]; exist {
				delete(b.subscriber, id)
				sub.Close()
			}
			continue
		}
		if e != nil {
			b.dispatch(e)
		}
	}
}

func (b *eventBus) Close() {
	b.once.Do(func() {
		b.lock.Lock()
		defer b.lock.Unlock()

		for id, s := range b.subscriber {
			delete(b.subscriber, id)
			s.Close()
		}
	})
}
