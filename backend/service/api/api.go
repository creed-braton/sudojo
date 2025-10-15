package api

import (
	"net/http"
	"sudojo/service/data"

	"github.com/google/uuid"
)

type service struct {
	data *data.Service
}

func New(data *data.Service) *service {
	return &service{
		data: data,
	}
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/lobbies": {
			"POST": s.postLobby,
		},
		"/lobbies/{id}": {
			"PATCH": s.patchLobby,
		},
		"/lobbies/{id}/stats": {
			"GET": s.getStats,
		},
	}
}

func (s *service) postLobby(w http.ResponseWriter, r *http.Request) {
	id, err := s.data.CreateLobby()
	if err != nil {
		http.Error(w, "internal server error", 500)
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (s *service) patchLobby(w http.ResponseWriter, r *http.Request) {
	input := r.PathValue("id")
	id, err := uuid.Parse(input)
	if err != nil {
		http.Error(w, "invalid lobby id format", 400)
		return
	}

	token, err := s.data.CreatePlayer(
		id, r.URL.Query().Get("name"),
	)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token))
}

func (s *service) getStats(w http.ResponseWriter, r *http.Request) {
	input := r.PathValue("id")
	id, err := uuid.Parse(input)
	if err != nil {
		http.Error(w, "invalid lobby id format", 400)
		return
	}

	b, err := s.data.Summary(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if b == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
