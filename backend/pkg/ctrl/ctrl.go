package ctrl

import (
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sync"
)

type Controller interface {
	Close()
	Pump() error
	Lobby() lobby.Lobby
	Broadcast(e event.Event) error
	Create(name string) (string, error)
	Join(token string) (Player, error)
}

type controller struct {
	lobby  lobby.Lobby
	buffer event.Buffer
	fanout event.Fanout
	active map[string]struct{}
	lock   sync.RWMutex
}

var _ Controller = &controller{}

func New(lobby lobby.Lobby) *controller {
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

func (c *controller) Lobby() lobby.Lobby {
	return c.lobby
}

func (c *controller) Broadcast(e event.Event) error {
	return c.buffer.Send(e)
}

func (c *controller) Create(name string) (string, error) {
	return c.lobby.Create(name)
}

func (c *controller) leave(token string) {
	event := event.New().SetType(event.LeaveEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(c.lobby.Leave(token)))
	c.Broadcast(event)
}

func (c *controller) Join(token string) (Player, error) {
	players, err := c.lobby.Join(token)
	if err != nil {
		return nil, err
	}
	buffer := event.NewBuffer()
	c.fanout.Register(token, buffer)

	event := event.New().SetType(event.JoinEvent).SetSender(token).
		SetPayload(event.NewPayload().SetPlayers(players))

	if err := c.Broadcast(event); err != nil {
		return nil, err
	}

	return NewPlayer(token, c.lobby, buffer, c.Broadcast, c.leave), nil
}
