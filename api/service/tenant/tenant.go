package tenant

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/manager"
	svc "sudojo/service/lobby"
	"sync"
	"time"
)

// Manages lobby lifecycle and provides access to lobby services with automatic pruning
// of inactive lobbies.
type Service interface {
	// Creates a new lobby service. Returns the lobby identifier or an error if creation
	// fails.
	Create(strict bool, size int) (string, error)
	// Retrieves a lobby state from the database. Returns nil if none exists with specified
	// lobby ID or an error if database call fails.
	Lobby(id string) (lobby.Lobby, error)
	// Retrieves or loads a lobby service by identifier. Returns nil if the lobby does not
	// exist, or an error if loading from database fails.
	Service(id string) (svc.Service, error)
}

type service struct {
	db     database.Database
	routes map[string]svc.Service
	lock   sync.RWMutex
}

var _ Service = &service{}

// Periodically removes lobbies that have been inactive for more than 10 minutes.
func (s *service) pruner() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.lock.Lock()
			for id, r := range s.routes {
				tolerance := int64(600) // 10 minutes in seconds
				frame := time.Now().UTC().Unix() - r.LastEvent()
				if frame > tolerance || r.Lobby().Game().Finished() != nil {
					if err := r.Shutdown(); err == nil {
						delete(s.routes, id)
					}
				}
			}
			s.lock.Unlock()
		}
	}
}

// Returns a new tenant service with automatic lobby pruning enabled.
func New(db database.Database) *service {
	s := &service{db: db, routes: make(map[string]svc.Service)}
	go s.pruner()
	return s
}

func (s *service) Create(strict bool, size int) (string, error) {
	lobby := lobby.Open(strict, size)
	if err := s.db.InsertLobby(lobby); err != nil {
		slog.Error(err.Error())
		return "", err
	}
	route := svc.New(
		manager.New(lobby), s.db,
		slog.With("lobby_id", lobby.Id()),
	)
	s.lock.Lock()
	s.routes[lobby.Id()] = route
	s.lock.Unlock()
	return lobby.Id(), nil
}

func (s *service) Lobby(id string) (lobby.Lobby, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if svc, exist := s.routes[id]; exist {
		return svc.Lobby(), nil
	}

	lobby, err := s.db.Lobby(id)
	if err != nil {
		slog.Error(err.Error(), "lobby_id", id)
		return nil, err
	}
	if lobby == nil {
		return nil, nil
	}

	return lobby, nil
}

func (s *service) Service(id string) (svc.Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	route, exist := s.routes[id]
	if !exist {
		lobby, err := s.db.Lobby(id)
		if err != nil {
			slog.Error(err.Error(), "lobby_id", id)
			return nil, err
		}
		if lobby == nil {
			return nil, nil
		}
		if lobby.Game().Finished() != nil {
			return nil, game.ErrFinished
		}

		route := svc.New(
			manager.New(lobby), s.db,
			slog.With("lobby_id", lobby.Id()),
		)
		s.routes[lobby.Id()] = route
		return route, nil
	}

	return route, nil
}
