package tenant

import (
	"sudojo/service/lobby"
	"sync"
)

type Service interface {
	Create() string
	Lobby(id string) lobby.Service
}

type service struct {
	lobbies map[string]lobby.Service
	lock    sync.RWMutex
}

var _ Service = &service{}

func New() *service {
	return &service{
		lobbies: make(map[string]lobby.Service),
	}
}

func (s *service) Create() string {
	l := lobby.New()
	s.lock.Lock()
	s.lobbies[l.Id()] = l
	s.lock.Unlock()
	return l.Id()
}

func (s *service) Lobby(id string) lobby.Service {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.lobbies[id]
}
