package data

import (
	"log"
	"sudojo/adapter/database"
	"sudojo/domain/lobby"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	logger  chan *lobby.Log
	lobbies map[string]*lobby.Lobby
	lock    sync.RWMutex
	db      database.Database
}

func (s *Service) cleaner() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.lock.Lock()
		for id, lobby := range s.lobbies {
			if idle := lobby.Idle(); idle {
				delete(s.lobbies, id)
				if err := s.db.UpdateLobby(lobby); err != nil {
					log.Printf("ERROR: failed writing lobby to db: %v", err)
				}
			}
		}
		s.lock.Unlock()
	}
}

func (s *Service) storer() {
	buffer := make([]*lobby.Log, 0, 1024)
	flush := func() {
		if err := s.db.InsertLogs(buffer); err != nil {
			log.Printf("ERROR: failed writing logs to db: %v", err)
		}
		buffer = buffer[:0]
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case log, ok := <-s.logger:
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, log)

			if len(buffer) >= 1024 {
				flush()
			}
		case <-ticker.C:
			flush()
		default:
		}
	}
}

func New(db database.Database) *Service {
	s := &Service{
		db:      db,
		logger:  make(chan *lobby.Log),
		lobbies: make(map[string]*lobby.Lobby),
	}

	go s.cleaner()
	go s.storer()

	return s
}

func (s *Service) Lobby(id uuid.UUID) (*lobby.Lobby, error) {
	s.lock.RLock()
	lobby, exists := s.lobbies[id.String()]
	s.lock.RUnlock()

	if !exists {
		var err error
		lobby, err = s.db.Lobby(id)
		if err != nil {
			log.Printf("ERROR: failed loading lobby from db: %v", err)
			return nil, err
		}

		if lobby == nil {
			return nil, nil
		}

		lobby.Init(s.logger)
		s.lock.Lock()
		s.lobbies[id.String()] = lobby
		s.lock.Unlock()
	}

	return lobby, nil
}

func (s *Service) CreateLobby() (string, error) {
	l := lobby.New(s.logger)
	err := s.db.InsertLobby(l)
	if err != nil {
		log.Printf("ERROR: failed writing lobby to db: %v", err)
		return "", err
	}

	s.lock.Lock()
	s.lobbies[l.Id] = l
	s.lock.Unlock()

	return l.Id, nil
}

func (s *Service) CreatePlayer(id uuid.UUID, name string) (string, error) {
	lobby, err := s.Lobby(id)
	if err != nil {
		return "", err
	}

	player, err := lobby.Create(name)
	if err != nil {
		return "", err
	}

	err = s.db.InsertPlayer(lobby.Id, player)
	if err != nil {
		log.Printf("ERROR: failed writing player to db: %v", err)
		return "", err
	}

	return player.Token, nil
}

func (s *Service) Summary(id uuid.UUID) ([]byte, error) {
	lobby, err := s.Lobby(id)
	if err != nil {
		return nil, err
	}
	if lobby == nil {
		return nil, nil
	}

	logs, err := s.db.Logs(id.String())
	if err != nil {
		log.Printf("ERROR: failed loading logs from db: %v", err)
		return nil, err
	}

	b, err := lobby.Summary(logs)
	if err != nil {
		log.Printf("ERROR: failed serializing summary: %v", err)
		return nil, err
	}

	return b, nil
}
