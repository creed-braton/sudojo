package event

import (
	"errors"
	"sync"
)

var (
	ErrClosedBus = errors.New("event bus is closed")
	ErrFullBus   = errors.New("event bus is full")
)

type Payload interface{}

type Event interface {
	Type() string
	Sender() string
	Receiver() string
	Trace() string
	Error() string
	Payload() Payload
}

type event struct {
	eventType string
	sender    string
	receiver  string
	trace     string
	errorMsg  string
	payload   Payload
}

func New(eventType, sender, receiver, trace, errorMsg string, payload Payload) *event {
	return &event{
		sender:    sender,
		receiver:  receiver,
		trace:     trace,
		eventType: eventType,
		errorMsg:  errorMsg,
		payload:   payload,
	}
}

func (e *event) Sender() string {
	return e.sender
}

func (e *event) Receiver() string {
	return e.receiver
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

type EventBus interface {
	Send(event Event) error
	Receive() (Event, error)
	Close()
}

type eventBus struct {
	events chan Event
	once   sync.Once
	closed chan struct{}
}

func NewEventBus() *eventBus {
	return &eventBus{
		events: make(chan Event, 256),
		closed: make(chan struct{}),
	}
}

func (b *eventBus) Send(event Event) error {
	select {
	case b.events <- event:
		return nil
	case <-b.closed:
		return ErrClosedBus
	default:
		return ErrFullBus
	}
}

func (b *eventBus) Receive() (Event, error) {
	select {
	case event := <-b.events:
		return event, nil
	case <-b.closed:
		return nil, ErrClosedBus
	}
}

func (b *eventBus) Close() {
	b.once.Do(func() {
		close(b.closed)
		close(b.events)
	})
}

type Fanout interface {
	Register(id string, bus EventBus)
	Deregister(id string)
}

type fanout struct {
	src    EventBus
	lock   sync.RWMutex
	routes map[string]EventBus
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
	defer f.lock.RUnlock()

	// targeted dispatch
	if recv := event.Receiver(); recv != "" {
		if bus := f.routes[recv]; bus != nil {
			if err := bus.Send(event); err == ErrClosedBus {
				go f.Deregister(recv)
			}
		}
		return
	}

	// broadcast
	for id, bus := range f.routes {
		if err := bus.Send(event); err == ErrClosedBus {
			go f.Deregister(id)
		}
	}
}

func (f *fanout) poll() {
	defer f.close()
	for {
		event, err := f.src.Receive()
		// stop when source bus has been closed
		if err == ErrClosedBus {
			return
		}
		if err == nil {
			f.dispatch(event)
		}
	}
}

func NewFanout(src EventBus) *fanout {
	f := &fanout{
		src:    src,
		routes: make(map[string]EventBus),
	}
	go f.poll()
	return f
}

func (f *fanout) Register(id string, bus EventBus) {
	f.lock.Lock()
	f.routes[id] = bus
	f.lock.Unlock()
}

func (f *fanout) Deregister(id string) {
	f.lock.Lock()
	delete(f.routes, id)
	f.lock.Unlock()
}
