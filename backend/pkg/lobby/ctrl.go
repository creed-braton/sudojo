package lobby

import (
	"sudojo/pkg/event"
	"sync"
)

type Controller interface {
	Close()
	Pump() error
	Lobby() Lobby
	Broadcast(e event.Event) error
	Create(name string) (string, error)
	Join(token string) (Player, error)
}

type controller struct {
	lobby  Lobby
	buffer event.Buffer
	fanout event.Fanout
	active map[string]struct{}
	lock   sync.RWMutex
}

var _ Controller = &controller{}

func NewController(lobby Lobby) *controller {
	buffer := event.NewBuffer()
	fanout := event.NewFanout(buffer)
	return &controller{lobby: lobby, buffer: buffer, fanout: fanout}
}

func (c *controller) Close() {
	c.buffer.Close()
}

func (c *controller) Pump() error {
	return c.fanout.Pump()
}

func (c *controller) Lobby() Lobby {
	return c.lobby
}

func (c *controller) Broadcast(e event.Event) error {
	return c.buffer.Send(e)
}

func (c *controller) Create(name string) (string, error) {
	return c.lobby.Create(name)
}

func (c *controller) leave(token string) {
	c.lobby.Leave(token)
	event := event.New().SetType(event.LeaveEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(c.lobby.Players()))
	c.Broadcast(event)
}

func (c *controller) Join(token string) (Player, error) {
	err := c.lobby.Player(token)
	if err != nil {
		return nil, err
	}
	c.lobby.Join(token)
	buffer := event.NewBuffer()
	c.fanout.Register(token, buffer)

	c.lock.RLock()
	defer c.lock.RUnlock()

	event := event.New().SetType(event.JoinEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(c.lobby.Players()))
	err = c.Broadcast(event)
	if err != nil {
		return nil, err
	}

	return NewPlayer(token, c.lobby, buffer, c.Broadcast, c.leave), nil
}
