package server

import (
	"encoding/json"
	"net/http"
	"sudojo/adp/database"
	"sudojo/adp/socket"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"

	"github.com/google/uuid"
)

func (s *server) postLobby(w http.ResponseWriter, r *http.Request) {
	req := &postLobbyReq{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	config, err := lobby.NewConfig(req.Strict, req.Pings, req.Notes, req.MaxSize)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	id, err := s.tenant.Create(config, req.Difficulty)
	if err == database.ErrDiffNotFound {
		http.Error(w, err.Error(), 404)
		return
	}
	if err == lobby.ErrInvalidSize {
		http.Error(w, err.Error(), 400)
		return
	}
	if err == game.ErrDifficulty {
		http.Error(w, err.Error(), 400)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (s *server) getPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "missing session token", 401)
		return
	}
	token := cookie.Value
	if len(token) != 32 {
		http.Error(w, "invalid token format", 400)
		return
	}

	svc, err := s.tenant.Session(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
	if svc == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	exist := svc.Player(token)
	if !exist {
		http.Error(w, "invalid token", 401)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) postPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	name := r.URL.Query().Get("name")

	svc, err := s.tenant.Session(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
	if svc == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	token, err := svc.CreatePlayer(name)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/api/lobbies/" + id,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   2592000, // 30 days
	})

	w.WriteHeader(http.StatusCreated)
}

func (s *server) getLobby(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}

	lobby, err := s.tenant.Lobby(id)
	if err == game.ErrNotFinished {
		http.Error(w, err.Error(), 409)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if lobby == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	b, err := marshalLobby(lobby)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *server) getSocket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "missing session token", 401)
		return
	}
	token := cookie.Value
	if len(token) != 32 {
		http.Error(w, "invalid token format", 400)
		return
	}

	svc, err := s.tenant.Session(id)
	if err == game.ErrFinished {
		http.Error(w, err.Error(), 410)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if svc == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	client := socket.New(conn, 60, 20)

	err = svc.JoinPlayer(token, client)
	if err == lobby.ErrPlayerNotFound {
		http.Error(w, err.Error(), 404)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
}
