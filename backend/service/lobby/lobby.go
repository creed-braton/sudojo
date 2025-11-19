package lobby

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/adapter/socket"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sudojo/service/player"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Service interface {
	Id() string
	Shutdown() error
	LastEvent() int64
	CreatePlayer(name string) (string, error)
	JoinPlayer(token string, conn *websocket.Conn) error
}

type service struct {
	lobby     lobby.Lobby
	buffer    event.Buffer
	fanout    event.Fanout
	logger    *slog.Logger
	db        database.Database
	lastEvent atomic.Int64
}

var _ Service = &service{}

func (s *service) pump() {
	for {
		if err := s.fanout.Pump(); err != nil {
			return
		}
	}
}

func (s *service) event() {
	s.lastEvent.Store(time.Now().UTC().Unix())
}

func New(lobby lobby.Lobby, db database.Database) *service {
	buffer := event.NewBuffer()
	s := &service{
		lobby:  lobby,
		logger: slog.With("lobby_id", lobby.Id()),
		buffer: buffer,
		fanout: event.NewFanout(buffer),
		db:     db,
	}
	s.event()
	go s.pump()
	return s
}

func (s *service) Id() string {
	return s.lobby.Id()
}

func (s *service) Shutdown() error {
	err := s.db.UpdateLobby(s.lobby)
	if err != nil {
		s.logger.Error(err.Error())
	}
	s.buffer.Close()
	s.logger.Info("shut down lobby")
	return err
}

func (s *service) LastEvent() int64 {
	return s.lastEvent.Load()
}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.lobby.Create(name)
	if err != nil {
		return "", err
	}
	if err := s.db.InsertPlayer(s.Id(), token, name); err != nil {
		return "", err
	}
	s.event()
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
	buffer := event.NewBuffer()
	s.fanout.Register(token, buffer)

	err = s.buffer.Send(event.New().
		SetType(event.JoinEvent).
		SetSender(token).
		SetPayload(p))
	if err != nil {
		return err
	} else {
		logger.Info("player joined")
	}

	s.event()
	player.New(
		buffer, client, s.lobby, s.buffer.Send,
		s.event, token, logger,
	).Start()

	return nil
}
