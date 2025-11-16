package lobby

import (
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"

	"github.com/gorilla/websocket"
)

type Service interface {
	Shutdown()
	Id() string
	CreatePlayer(name string) (string, error)
	Join(token string, conn *websocket.Conn) error
}

type service struct {
	lobby lobby.Lobby
}

var _ Service = &service{}

func (s *service) pump() {
	for {
		if err := s.lobby.Pump(); err != nil {
			return
		}
	}
}

func New(maxPlayer int, strict bool) *service {
	s := &service{
		lobby: lobby.Open(maxPlayer, strict),
	}
	go s.pump()
	return s
}

func (s *service) Shutdown() {
	s.lobby.Close()
}

func (s *service) Id() string {
	return s.lobby.Id()
}

func (s *service) CreatePlayer(name string) (string, error) {
	return s.lobby.CreatePlayer(name)
}

func (s *service) sendPump(bus event.EventBus, client socket.Client) {
	defer func() {
		bus.Close()
		client.Close()
	}()

	for {
		e, err := bus.Receive()
		if err != nil {
			return
		}
		if err := client.Send(e); err != nil {
			return
		}
	}
}

func (s *service) handler(token string, bus event.EventBus, client socket.Client) {
	defer func() {
		bus.Close()
		s.lobby.Leave(token)
	}()

	for {
		msg, err := client.Receive()
		if err != nil {
			return
		}
		if msg.Type == event.InsertEvent {
			e, err := s.lobby.Insert(*msg.Row, *msg.Column, *msg.Value, token, msg.Trace)
			if err != nil {
				return
			}
			if e != nil {
				if err := bus.Send(e); err != nil {
					return
				}
			}
		} else if msg.Type == event.PingEvent {
			e, err := s.lobby.Ping(*msg.Row, *msg.Column, token, msg.Trace)
			if err != nil {
				return
			}
			if e != nil {
				if err := bus.Send(e); err != nil {
					return
				}
			}
		} else {
			err := bus.Send(s.lobby.State(token, msg.Trace))
			if err != nil {
				return
			}
		}
	}
}

func (s *service) Join(token string, conn *websocket.Conn) error {
	bus := event.NewEventBus()
	err := s.lobby.Join(token, bus)
	if err != nil {
		conn.Close()
		return err
	}

	client := socket.NewClient(conn)
	go client.WritePump()
	go client.ReadPump()
	go s.sendPump(bus, client)
	go s.handler(token, bus, client)

	return nil
}
