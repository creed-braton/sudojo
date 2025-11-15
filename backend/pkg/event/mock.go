package event

import "sudojo/pkg/sudoku"

type mockPayload struct{}

func NewMockPayload() *mockPayload {
	return &mockPayload{}
}

func (p *mockPayload) Current() sudoku.Sudoku {
	return nil
}

func (p *mockPayload) Initial() sudoku.Sudoku {
	return nil
}

func (p *mockPayload) Row() *int {
	return nil
}

func (p *mockPayload) Column() *int {
	return nil
}

func (p *mockPayload) Value() *int {
	return nil
}

func (p *mockPayload) Conflict() string {
	return ""
}

func (p *mockPayload) Players() []*PlayerStatus {
	return []*PlayerStatus{}
}

type mockEventChan struct {
	send  func(e Event)
	recv  func() (Event, error)
	wait  func() (Event, error)
	close func()
}

var _ EventChan = &mockEventChan{}

func NewMockEventChan(
	send func(e Event), close func(),
	recv, wait func() (Event, error),
) *mockEventChan {
	return &mockEventChan{
		send:  send,
		recv:  recv,
		wait:  wait,
		close: close,
	}
}

func (ch *mockEventChan) Send(e Event) {
	ch.send(e)
}

func (ch *mockEventChan) NonBlockRecv() (Event, error) {
	return ch.recv()
}

func (ch *mockEventChan) Receive() (Event, error) {
	return ch.recv()
}

func (ch *mockEventChan) Close() {}

type mockEventBus struct {
	register   func(id string, pub, sub EventChan)
	deregister func(id string)
	pump       func()
	close      func()
}

var _ EventBus = &mockEventBus{}

func NewMockEventBus(
	register func(id string, pub, sub EventChan),
	deregister func(id string),
	pump, close func(),
) *mockEventBus {
	return &mockEventBus{
		register:   register,
		deregister: deregister,
		pump:       pump,
		close:      close,
	}
}

func (b *mockEventBus) Register(id string, pub, sub EventChan) {
	b.register(id, pub, sub)
}

func (b *mockEventBus) Deregister(id string) {
	b.deregister(id)
}

func (b *mockEventBus) Pump() {
	b.pump()
}

func (b *mockEventBus) Close() {
	b.close()
}
