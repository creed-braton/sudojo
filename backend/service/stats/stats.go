package stats

import (
	"log"
	"net/http"
	"sudojo/adapter/database"
	"sudojo/domain/lobby"
	"time"
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
	return map[string]map[string]http.HandlerFunc{}
}
