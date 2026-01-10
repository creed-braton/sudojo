package event

import (
	"errors"
	"sync"
)

var (
	ErrHubClosed = errors.New("event hub is closed")
)

// Manages pool of event buffers and offers broadcast functionality.
type Hub interface {
	// Registers a buffer under the specified ID for event broadcasting. If a
	// buffer is already registered under the ID, the old buffer gets closed with
	// TakeoverReason and replaced with the new buffer. Returns ErrHubClosed if
	// the hub has been closed.
	Register(id string, buffer Buffer) error
	// Removes the buffer with the specified ID from the hub without closing it.
	Deregister(id string)
	// Sends the provided event to all open buffers in the hub. De-registers closed
	// buffers from the hub.
	Broadcast(event Event)
	// Closes all registered buffers with the provided reason and marks the hub
	// as closed. Subsequent calls to Register will return ErrHubClosed.
	Close(reason int)
	// Sends the event to a specific buffer under the provided ID. If no buffer is
	// registered under that ID, the message is silently dropped. De-registers
	// closed buffers from the hub.
	Send(id string, event Event)
}

type hub struct {
	buffers map[string]Buffer
	closed  bool
	lock    sync.Mutex
}

var _ Hub = &hub{}

// Creates an empty event hub.
func NewHub() *hub {
	return &hub{buffers: make(map[string]Buffer)}
}

func (h *hub) Register(id string, buffer Buffer) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.closed {
		return ErrHubClosed
	}

	if old := h.buffers[id]; old != nil {
		old.Close(TakeoverReason)
	}
	h.buffers[id] = buffer
	return nil
}

func (h *hub) Deregister(id string) {
	h.lock.Lock()
	delete(h.buffers, id)
	h.lock.Unlock()
}

func (h *hub) Broadcast(event Event) {
	h.lock.Lock()
	for id, buffer := range h.buffers {
		if _, ok := buffer.Send(event).(*BufferClosedError); ok {
			delete(h.buffers, id)
		}
	}
	h.lock.Unlock()
}

func (h *hub) Send(id string, event Event) {
	h.lock.Lock()
	defer h.lock.Unlock()

	buffer, ok := h.buffers[id]
	if !ok {
		return
	}

	if _, ok := buffer.Send(event).(*BufferClosedError); ok {
		delete(h.buffers, id)
	}
}

func (h *hub) Close(reason int) {
	h.lock.Lock()
	h.closed = true
	for _, buffer := range h.buffers {
		buffer.Close(reason)
	}
	h.lock.Unlock()
}
