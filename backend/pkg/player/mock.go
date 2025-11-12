package player

import "sudojo/pkg/event"

type mockPlayer struct {
	token string
	name  string
	send  func(e event.Event) error
	recv  func() (event.Event, error)
}

var _ Player = &mockPlayer{}

func NewMock(
	token, name string,
	send func(e event.Event) error,
	recv func() (event.Event, error),
) *mockPlayer {
	return &mockPlayer{
		token: token,
		name:  name,
		send:  send,
		recv:  recv,
	}
}

func (p *mockPlayer) Token() string {
	return p.token
}

func (p *mockPlayer) Name() string {
	return p.name
}

func (p *mockPlayer) Send(e event.Event) error {
	return p.send(e)
}

func (p *mockPlayer) Receive() (event.Event, error) {
	return p.recv()
}

func (p *mockPlayer) Close() {
}
