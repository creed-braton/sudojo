package lobby

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/adapter/socket"
	"sudojo/pkg/ctrl"

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
	ctrl   ctrl.Controller
	logger *slog.Logger
	db     database.Database
}

var _ Service = &service{}

// Continuously distributes events from the broadcast buffer to all registered players.
func (s *service) pump() {
	for {
		if err := s.ctrl.Pump(); err != nil {
			return
		}
	}
}

// Returns a new lobby service and starts the event pump goroutine.
func New(ctrl ctrl.Controller, db database.Database, logger *slog.Logger) *service {
	s := &service{
		ctrl:   ctrl,
		logger: logger,
		db:     db,
	}
	go s.pump()
	return s
}

func (s *service) Id() string {
	return s.ctrl.Lobby().Id()
}

func (s *service) Shutdown() error {
	err := s.db.UpdateLobby(s.ctrl.Lobby())
	if err != nil {
		s.logger.Error(err.Error())
	}
	s.ctrl.Close()
	s.logger.Info("shut down lobby")
	return err
}

func (s *service) LastEvent() int64 {
	return s.ctrl.LastEvent()
}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.ctrl.Create(name)
	if err != nil {
		return "", err
	}
	if err := s.db.InsertPlayer(s.Id(), token, name); err != nil {
		return "", err
	}
	s.logger.Info("player created", "player_token", token)
	return token, nil
}

func (s *service) JoinPlayer(token string, conn *websocket.Conn) error {
	player, err := s.ctrl.Join(token)
	if err != nil {
		conn.Close()
		return err
	}

	client := socket.NewClient(player, conn)
	go client.WritePump()
	go client.ReadPump()

	return nil
}
