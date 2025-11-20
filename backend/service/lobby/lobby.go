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

// Coordinates player connections, event distribution, and lobby state persistence.
type Service interface {
	// Returns the unique identifier of the lobby.
	Id() string
	// Persists the current lobby state to the database and closes the event buffer.
	// Returns an error if the lobby update in the database fails.
	Shutdown() error
	// Returns the Unix timestamp of the last recorded event.
	LastEvent() int64
	// Creates a new player with the given name and persists it to the database.
	// Returns the player's authentication token or an error if creation fails.
	CreatePlayer(name string) (string, error)
	// Registers a player's WebSocket connection, broadcasts a join event, and starts
	// the player service. Returns player.ErrPlayerNotFound if the token is invalid.
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

// Continuously distributes events from the broadcast buffer to all registered players.
func (s *service) pump() {
	for {
		if err := s.fanout.Pump(); err != nil {
			return
		}
	}
}

// Updates the last event timestamp to the current time.
func (s *service) event() {
	s.lastEvent.Store(time.Now().UTC().Unix())
}

// Returns a new lobby service and starts the event pump goroutine.
func New(lobby lobby.Lobby, db database.Database, logger *slog.Logger) *service {
	buffer := event.NewBuffer()
	s := &service{
		lobby:  lobby,
		logger: logger,
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
