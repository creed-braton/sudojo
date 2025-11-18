package server

import (
	"net/http"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"

	"github.com/google/uuid"
)

func (s *server) postLobby(w http.ResponseWriter, r *http.Request) {
	id, err := s.tenant.Create()
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (s *server) patchLobby(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	name := r.URL.Query().Get("name")

	l := s.tenant.Lobby(id)
	if l == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	token, err := l.CreatePlayer(name)
	if err != nil {
		if err == lobby.ErrLobbyFull {
			http.Error(w, err.Error(), 409)
		} else if err == player.ErrInvalidChar ||
			err == player.ErrNameTooLong {
			http.Error(w, err.Error(), 404)
		} else {
			http.Error(w, "internal server error", 500)
		}
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token))
}

func (s *server) getLobby(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	token := r.URL.Query().Get("token")
	if len(token) != 32 {
		http.Error(w, "invalid token format", 400)
		return
	}

	lobby := s.tenant.Lobby(id)
	if lobby == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	err = lobby.JoinPlayer(token, conn)
	if err == player.ErrPlayerNotFound {
		http.Error(w, err.Error(), 404)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
}
