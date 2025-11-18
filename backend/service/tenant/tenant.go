package tenant

import (
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/pkg/lobby"
	svc "sudojo/service/lobby"
	"sync"
)

type Service interface {
	Create() (string, error)
	Lobby(id string) svc.Service
}

type service struct {
	db      database.Database
	lobbies map[string]svc.Service
	lock    sync.RWMutex
}

var _ Service = &service{}

func New(db database.Database) *service {
	return &service{
		db:      db,
		lobbies: make(map[string]svc.Service),
	}
}

func (s *service) Create() (string, error) {
	lobby := lobby.Open(8, false)
	if err := s.db.InsertLobby(lobby); err != nil {
		slog.Error(err.Error())
		return "", err
	}
	l := svc.New(lobby, s.db)
	s.lock.Lock()
	s.lobbies[l.Id()] = l
	s.lock.Unlock()
	return l.Id(), nil
}

func (s *service) Lobby(id string) svc.Service {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.lobbies[id]
}
