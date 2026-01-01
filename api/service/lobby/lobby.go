package lobby

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/adapter/socket"
	"sudojo/pkg/lobby"
	"sudojo/pkg/manager"

	"github.com/gorilla/websocket"
)

// Coordinates player connections, event distribution, and lobby state persistence.
type Service interface {
	// Returns the unique identifier of the lobby.
	Id() string
	// Persists the current lobby state to the database and closes the event buffer.
	// Returns an error if the lobby update in the database fails.
	Shutdown(reason int) error
	// Returns the Unix timestamp of the last recorded event.
	LastEvent() int64
	// Returns the lobby state of the service.
	Lobby() lobby.Lobby
	// Creates a new player with the given name and persists it to the database.
	// Returns the player's authentication token or an error if creation fails.
	CreatePlayer(name string) (string, error)
	// Registers a player's WebSocket connection, broadcasts a join event, and starts
	// the player service. Returns player.ErrPlayerNotFound if the token is invalid.
	JoinPlayer(token string, conn *websocket.Conn) error
}

type service struct {
	manager manager.Manager
	logger  *slog.Logger
	db      database.Database
}

var _ Service = &service{}

// Continuously distributes events from the broadcast buffer to all registered players.
func (s *service) pump() {
	for {
		if err := s.manager.Pump(); err != nil {
			return
		}
	}
}

// Returns a new lobby service and starts the event pump goroutine.
func New(mng manager.Manager, db database.Database, logger *slog.Logger) *service {
	s := &service{
		manager: mng,
		logger:  logger,
		db:      db,
	}
	go s.pump()
	return s
}

func (s *service) Id() string {
	return s.manager.Lobby().Id()
}

func (s *service) Shutdown(reason int) error {
	err := s.db.UpdateLobby(s.manager.Lobby())
	if err != nil {
		s.logger.Error(err.Error())
	}
	s.manager.Close(reason)
	s.logger.Info("shut down lobby")
	return err
}

func (s *service) LastEvent() int64 {
	return s.manager.LastEvent()
}

func (s *service) Lobby() lobby.Lobby {
	return s.manager.Lobby()
}

func (s *service) CreatePlayer(name string) (string, error) {
	token, err := s.manager.Create(name)
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
	player, err := s.manager.Join(token)
	if err != nil {
		conn.Close()
		return err
	}

	client := socket.NewClient(player, conn)
	go client.WritePump()
	go client.ReadPump()

	return nil
}
