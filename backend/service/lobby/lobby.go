package lobby

import (
	"log/slog"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"

	"github.com/gorilla/websocket"
)

type Service interface {
	Id() string
	Shutdown()
	CreatePlayer(name string) (string, error)
	JoinPlayer(token string, conn *websocket.Conn) error
}

type service struct {
	lobby  lobby.Lobby
	bus    event.EventBus
	fanout event.Fanout
	logger *slog.Logger
}

var _ Service = &service{}

func (s *service) pump() {
	for {
		if err := s.fanout.Pump(); err != nil {
			s.logger.Info(err.Error())
		}
	}
}

func New() *service {
	lobby := lobby.Open(8, false)
	bus := event.NewEventBus()
	s := &service{
		lobby:  lobby,
		logger: slog.With("lobby_id", lobby.Id()),
		bus:    bus,
		fanout: event.NewFanout(bus),
	}
	go s.pump()
	return s
}

func (s *service) Id() string {
	return s.lobby.Id()
}

func (s *service) Shutdown() {

}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.lobby.Create(name)
	if err != nil {
		return "", err
	}
	s.logger.Info("player created", "player_token", token)
	return token, nil
}

func (s *service) handler(token string, bus event.EventBus, client socket.Client, logger *slog.Logger) {
	logger.Info("handler started")
	defer func() {
		bus.Close()
		s.fanout.Deregister(token)
		p, err := s.lobby.Leave(token)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		err = s.bus.Send(event.New(event.LeaveEvent, token, "", "", p))
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Info("handler terminated")
	}()

	for {
		msg, err := client.Receive()
		if err != nil {
			return
		}
		if msg.Type == event.InsertEvent {
			p, err := s.lobby.Insert(*msg.Row, *msg.Column, *msg.Value)
			if err != nil {
				bus.Send(event.New(
					event.InsertEvent,
					token, msg.Trace,
					err.Error(), p,
				))
				continue
			}
			if p != nil {
				bus.Send(event.New(event.InsertEvent, token, msg.Trace, "", p))
			}
		} else if msg.Type == event.PingEvent {
			p, err := s.lobby.Ping(*msg.Row, *msg.Column)
			if err != nil {
				bus.Send(event.New(
					event.PingEvent,
					token, msg.Trace,
					err.Error(), p,
				))
				continue
			}
			if p != nil {
				bus.Send(event.New(event.PingEvent, token, msg.Trace, "", p))
			}
		} else {
			bus.Send(event.New(event.StateEvent, token, msg.Trace, "", s.lobby.State()))
		}
	}
}

func (s *service) JoinPlayer(token string, conn *websocket.Conn) error {
	p, err := s.lobby.Join(token)
	if err != nil {
		conn.Close()
		return err
	}

	bus := event.NewEventBus()
	s.fanout.Register(token, bus)
	err = s.bus.Send(event.New(event.JoinEvent, token, "", "", p))
	if err != nil {
		return err
	} else {
		s.logger.Info("player joined", "token", token)
	}

	client := socket.NewClient(conn)
	logger := s.logger.With("player_token", token).With("client_id", client.Id())

	go func() {
		logger.Info("write pump started")
		if err := client.WritePump(); err != nil {
			logger.Info("write pump terminated")
		}
	}()
	go func() {
		logger.Info("read pump started")
		if err := client.ReadPump(); err != nil {
			logger.Info("read pump terminated")
		}
	}()

	go func() {
		logger.Info("send pump started")
		defer func() {
			bus.Close()
			client.Close()
			logger.Info("send pump terminated")
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
	}()
	go s.handler(token, bus, client, logger)

	return nil
}
