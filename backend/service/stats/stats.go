package stats

import (
	"encoding/json"
	"net/http"
	"sudojo/domain/lobby"
	"sync"
)

type service struct {
	lobbies  map[string]*lobby.Report
	complete chan *lobby.Lobby
	lock     sync.RWMutex
}

func (s *service) exporter() {
	for l := range s.complete {
		rep := l.Export()
		if rep != nil {
			s.lock.Lock()
			s.lobbies[l.Id] = rep
			s.lock.Unlock()
		}
	}
}

func New(complete chan *lobby.Lobby) *service {
	s := &service{
		lobbies:  make(map[string]*lobby.Report),
		complete: complete,
	}
	go s.exporter()
	return s
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/stats/{id}": {
			"GET": s.getStats,
		},
	}
}

func (s *service) getStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.lock.RLock()
	defer s.lock.RUnlock()

	rep := s.lobbies[id]
	if rep == nil {
		http.Error(w, "game not found", 404)
		return
	}

	data, err := json.Marshal(rep)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
