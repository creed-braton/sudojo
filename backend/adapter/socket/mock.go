package socket

import "sudojo/core/event"

type mockClient struct {
	close func()
	send  func(e event.Event) error
	recv  func() (*Message, error)
}

var _ Client = &mockClient{}

func NewMockClient(
	closeFn func(),
	send func(e event.Event) error,
	recv func() (*Message, error),
) *mockClient {
	return &mockClient{
		close: closeFn,
		send:  send,
		recv:  recv,
	}
}

func (c *mockClient) Id() string {
	return "mock-client"
}

func (c *mockClient) Close() {
	c.close()
}

func (c *mockClient) WritePump() error {
	return nil
}

func (c *mockClient) ReadPump() error {
	return nil
}

func (c *mockClient) Send(e event.Event) error {
	return c.send(e)
}

func (c *mockClient) Receive() (*Message, error) {
	return c.recv()
}
