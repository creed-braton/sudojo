package event

type mockPayload struct{}

type mockBuffer struct {
	send func(e Event) error
	recv func() (Event, error)
}

var _ Buffer = &mockBuffer{}

func NewMockBuffer(send func(e Event) error, recv func() (Event, error)) *mockBuffer {
	return &mockBuffer{send: send, recv: recv}
}

func (b *mockBuffer) Send(e Event) error {
	return b.send(e)
}

func (b *mockBuffer) Receive() (Event, error) {
	return b.recv()
}

func (b *mockBuffer) Close() {}

type mockFanout struct {
	register   func(id string, buffer Buffer)
	deregister func(id string)
	pump       func() error
}

var _ Fanout = &mockFanout{}

func NewMockFanout(
	register func(id string, buffer Buffer),
	deregister func(id string),
	pump func() error,
) *mockFanout {
	return &mockFanout{register: register, deregister: deregister, pump: pump}
}

func (f *mockFanout) Register(id string, buffer Buffer) {
	f.register(id, buffer)
}

func (f *mockFanout) Deregister(id string) {
	f.deregister(id)
}

func (f *mockFanout) Pump() error {
	return f.pump()
}
