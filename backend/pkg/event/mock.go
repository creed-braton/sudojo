package event

type mockEventChan struct {
	recv func() (Event, error)
}

var _ EventChan = &mockEventChan{}

func NewMockEventChan(recv func() (Event, error)) *mockEventChan {
	return &mockEventChan{recv: recv}
}

func (ch *mockEventChan) Send(e Event) {}

func (ch *mockEventChan) NonBlockRecv() (Event, error) {
	return ch.recv()
}

func (ch *mockEventChan) Receive() (Event, error) {
	return ch.recv()
}

func (ch *mockEventChan) Close() {}

type mockEventBus struct {
}

var _ EventBus = &mockEventBus{}

func (b *mockEventBus) Register(id string, pub, sub EventChan) {}

func (b *mockEventBus) Deregister(id string) {}

func (b *mockEventBus) Pump() {}

func (b *mockEventBus) Close() {}
