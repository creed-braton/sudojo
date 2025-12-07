package tenant

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/pkg/ctrl"
	"sudojo/pkg/lobby"
	svc "sudojo/service/lobby"
	"sync"
	"time"
)

// Manages lobby lifecycle and provides access to lobby services with automatic pruning
// of inactive lobbies.
type Service interface {
	// Creates a new lobby service. Returns the lobby identifier or an error if
	// creation fails.
	Create() (string, error)
	// Retrieves or loads a lobby service by identifier. Returns nil if the lobby
	// does not exist, or an error if loading from database fails.
	Lobby(id string) (svc.Service, error)
}

type service struct {
	db      database.Database
	lobbies map[string]svc.Service
	lock    sync.RWMutex
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
			for id, l := range s.lobbies {
				tolerance := int64(600) // 10 minutes in seconds
				frame := time.Now().UTC().Unix() - l.LastEvent()
				if frame > tolerance {
					if err := l.Shutdown(); err == nil {
						delete(s.lobbies, id)
					}
				}
			}
			s.lock.Unlock()
		}
	}
}

// Returns a new tenant service with automatic lobby pruning enabled.
func New(db database.Database) *service {
	s := &service{
		db:      db,
		lobbies: make(map[string]svc.Service),
	}
	go s.pruner()
	return s
}

func (s *service) Create() (string, error) {
	l := lobby.Open(false, 8)
	if err := s.db.InsertLobby(l); err != nil {
		slog.Error(err.Error())
		return "", err
	}
	ctrl := ctrl.New(l)
	lobby := svc.New(ctrl, s.db, slog.With("lobby_id", l.Id()))
	s.lock.Lock()
	s.lobbies[lobby.Id()] = lobby
	s.lock.Unlock()
	return lobby.Id(), nil
}

func (s *service) Lobby(id string) (svc.Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	lobbySvc, exist := s.lobbies[id]
	if !exist {
		l, err := s.db.Lobby(id)
		if err != nil {
			slog.Error(err.Error(), "lobby_id", id)
			return nil, err
		}
		if l == nil {
			return nil, nil
		}

		ctrl := ctrl.New(l)
		lobbySvc := svc.New(ctrl, s.db, slog.With("lobby_id", l.Id()))
		s.lobbies[l.Id()] = lobbySvc
		return lobbySvc, nil
	}

	return lobbySvc, nil
}
