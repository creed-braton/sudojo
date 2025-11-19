package tenant

import (
	"fmt"
	"log/slog"
	"sudojo/adapter/database"
	"sudojo/pkg/lobby"
	svc "sudojo/service/lobby"
	"sync"
	"time"
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
				slog.Info(fmt.Sprintf("%d", tolerance))
				slog.Info(fmt.Sprintf("%d", frame))
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

func New(db database.Database) *service {
	s := &service{
		db:      db,
		lobbies: make(map[string]svc.Service),
	}
	go s.pruner()
	return s
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
