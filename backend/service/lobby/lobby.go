package lobby

import (
	"log/slog"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sudojo/service/player"

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
	s.bus.Close()
}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.lobby.Create(name)
	if err != nil {
		return "", err
	}
	s.logger.Info("player created", "player_token", token)
	return token, nil
}

func (s *service) JoinPlayer(token string, conn *websocket.Conn) error {
	p, err := s.lobby.Join(token)
	if err != nil {
		conn.Close()
		return err
	}

	client := socket.NewClient(conn)
	logger := s.logger.With("player_token", token).With("client_id", client.Id())
	bus := event.NewEventBus()
	s.fanout.Register(token, bus)

	err = s.bus.Send(event.New().
		SetType(event.JoinEvent).
		SetSender(token).
		SetPayload(p))
	if err != nil {
		return err
	} else {
		logger.Info("player joined")
	}

	player.New(bus, client, s.lobby, s.bus.Send, token, logger).Start()

	return nil
}
