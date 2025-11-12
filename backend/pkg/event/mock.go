package event

type mockEventBus struct {
	send func(e Event) error
	recv func() (Event, error)
}

var _ EventBus = &mockEventBus{}

func NewMockEventBus(send func(e Event) error, recv func() (Event, error)) *mockEventBus {
	return &mockEventBus{send: send, recv: recv}
}

func (b *mockEventBus) Send(e Event) error {
	return b.send(e)
}

func (b *mockEventBus) Receive() (Event, error) {
	return b.recv()
}

func (b *mockEventBus) Close() {}

type mockFanout struct {
	register   func(id string, bus EventBus)
	deregister func(id string)
	poll       func() error
}

var _ Fanout = &mockFanout{}

func NewMockFanout(
	register func(id string, bus EventBus),
	deregister func(id string),
	poll func() error,
) *mockFanout {
	return &mockFanout{register: register, deregister: deregister, poll: poll}
}

func (f *mockFanout) Register(id string, bus EventBus) {
	f.register(id, bus)
}

func (f *mockFanout) Deregister(id string) {
	f.deregister(id)
}

func (f *mockFanout) Poll() error {
	return f.poll()
}
