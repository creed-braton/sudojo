package tenant

import (
	"fmt"
	"log/slog"
	"sudojo/adp/database"
	"sudojo/adp/metrics"
	"sudojo/adp/socket"
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
	"sudojo/svc/session"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Session(id string) (session.Service, error)
	Create(config lobby.Config, difficulty string) (string, error)
	Lobby(id string) (lobby.Lobby, error)
}

type service struct {
	db       database.Database
	logger   *slog.Logger
	metrics  metrics.Metrics
	sessions map[string]session.Service
	lock     sync.RWMutex
}

var _ Service = &service{}

func (s *service) prune(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.lock.Lock()
			for id, svc := range s.sessions {
				lifetime := time.Now().UTC().UnixNano() - svc.CreatedAt()

				if svc.Lobby().Game().FinishedAt() != nil {
					svc.Close(socket.CloseFinished, "game is finished")
					delete(s.sessions, id)
				} else if svc.Count() < 1 && lifetime > int64(5*time.Minute) {
					svc.Close(socket.CloseIdle, "lobby is idle")
					delete(s.sessions, id)
				}
			}
			s.lock.Unlock()
		}
	}
}

func New(
	db database.Database,
	logger *slog.Logger,
	metrics metrics.Metrics,
	interval int,
) *service {
	s := &service{
		db:       db,
		logger:   logger,
		metrics:  metrics,
		sessions: make(map[string]session.Service),
	}
	go s.prune(interval)
	return s
}

func (s *service) Session(id string) (session.Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	svc, exist := s.sessions[id]
	if !exist {
		lobby, err := s.db.Lobby(id)
		if err != nil {
			s.logger.Error(fmt.Sprintf("retrieve lobby: %v", err))
			return nil, err
		}

		if lobby == nil {
			return nil, nil
		}
		if lobby.Game().FinishedAt() != nil {
			return nil, game.ErrFinished
		}

		svc = session.New(
			lobby, s.db, slog.With("lobby_id", lobby.Id()), s.metrics,
		)
		s.sessions[lobby.Id()] = svc
		return svc, nil
	}

	return svc, nil
}

func (s *service) Create(
	config lobby.Config, difficulty string,
) (string, error) {
	if !game.ValidDifficulty(difficulty) {
		return "", game.ErrDifficulty
	}

	var g game.Game
	if difficulty == game.Joker {
		g = game.Generate(time.Now().UTC().UnixNano())
	} else {
		var err error
		g, err = s.db.SampleGame(difficulty)
		if err == database.ErrDiffNotFound {
			return "", err
		}
		if err != nil {
			s.logger.Error(fmt.Sprintf("sample game: %v", err))
			return "", err
		}
	}
	g.Start(time.Now().UTC().UnixNano())

	lobby := lobby.New(
		uuid.NewString(), config, g,
		history.New([]history.Artifact{}),
		make(map[string]player.Player),
	)
	if err := s.db.InsertLobby(lobby); err != nil {
		s.logger.Error(fmt.Sprintf("insert lobby: %v", err))
		return "", err
	}
	s.logger.Info("lobby created", "lobby_id", lobby.Id())

	svc := session.New(
		lobby, s.db, slog.With("lobby_id", lobby.Id()), s.metrics,
	)
	s.lock.Lock()
	s.sessions[lobby.Id()] = svc
	s.lock.Unlock()

	return lobby.Id(), nil
}

func (s *service) Lobby(id string) (lobby.Lobby, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var lobby lobby.Lobby
	if svc, exist := s.sessions[id]; exist {
		lobby = svc.Lobby()
	} else {
		var err error
		lobby, err = s.db.Lobby(id)
		if err != nil {
			slog.Error(fmt.Sprintf("retrieve lobby: %v", err), "lobby_id", id)
			return nil, err
		}
	}

	if lobby == nil {
		return nil, nil
	}
	if lobby.Game().FinishedAt() == nil {
		return nil, game.ErrNotFinished
	}

	return lobby, nil
}
