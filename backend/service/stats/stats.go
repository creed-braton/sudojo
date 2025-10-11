package stats

import (
	"log"
	"net/http"
	"sudojo/adapter/database"
	"sudojo/domain/lobby"
	"time"

	"github.com/google/uuid"
)

type service struct {
	logger chan *lobby.Log
	db     database.Database
}

func (s *service) storer() {
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

func New(logger chan *lobby.Log, db database.Database) *service {
	s := &service{
		logger: logger,
		db:     db,
	}
	go s.storer()
	return s
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/lobbies/{id}/stats": {
			"GET": s.getStats,
		},
	}
}

func (s *service) getStats(w http.ResponseWriter, r *http.Request) {
	input := r.PathValue("id")
	id, err := uuid.Parse(input)
	if err != nil {
		http.Error(w, "invalid lobby id format", 400)
		return
	}

	lobby, err := s.db.Lobby(id)
	if err != nil {
		log.Printf("ERROR: failed loading lobby from db: %v", err)
		http.Error(w, "internal server error", 500)
		return
	}
	if lobby == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	logs, err := s.db.Logs(id.String())
	if err != nil {
		log.Printf("ERROR: failed loading logs from db: %v", err)
		http.Error(w, "internal server error", 500)
		return
	}

	b, err := lobby.Summary(logs)
	if err != nil {
		log.Printf("ERROR: failed serializing summary: %v", err)
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
