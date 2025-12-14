package server

import (
	"encoding/json"
	"net/http"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"

	"github.com/google/uuid"
)

func (s *server) postLobby(w http.ResponseWriter, r *http.Request) {
	id, err := s.tenant.Create(false, 8)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (s *server) postPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}
	name := r.URL.Query().Get("name")

	l, err := s.tenant.Service(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
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

type response struct {
	Id       string    `json:"id"`
	Current  [][]int   `json:"current_state"`
	Initial  [][]int   `json:"initial_state"`
	Started  *int64    `json:"started,omitempty"`
	Finished *int64    `json:"finished,omitempty"`
	Players  []history `json:"players"`
	Strict   bool      `json:"strict"`
	Size     int       `json:"size"`
}

type history struct {
	Name      string     `json:"name"`
	Artifacts []artifact `json:"artifacts"`
}

type artifact struct {
	Timestamp int64 `json:"timestamp"`
	Row       int   `json:"row"`
	Column    int   `json:"column"`
	Value     int   `json:"value"`
}

func convert(l lobby.Lobby) *response {
	// Group artifacts by player token
	artifacts := make(map[string][]artifact)
	for _, a := range l.History().Artifacts() {
		token := a.Player()
		artifacts[token] = append(artifacts[token], artifact{
			Timestamp: a.Timestamp(),
			Row:       a.Row(),
			Column:    a.Column(),
			Value:     a.Value(),
		})
	}

	// Build player history list
	histories := make([]history, 0, len(l.Players()))
	for _, p := range l.Players() {
		token := p.Token()
		artifacts := artifacts[token]
		if artifacts == nil {
			artifacts = []artifact{}
		}
		histories = append(histories, history{
			Name:      p.Name(),
			Artifacts: artifacts,
		})
	}

	// Get current and initial states
	var current, initial [][]int
	if l.Game().Current() != nil {
		current = l.Game().Current().Int()
	}
	if l.Game().Initial() != nil {
		initial = l.Game().Initial().Int()
	}

	return &response{
		Id:       l.Id(),
		Current:  current,
		Initial:  initial,
		Started:  l.Game().Started(),
		Finished: l.Game().Finished(),
		Players:  histories,
		Strict:   l.Strict(),
		Size:     l.Size(),
	}
}

func (s *server) getLobby(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid lobby id", 400)
		return
	}

	lobby, err := s.tenant.Lobby(id)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}
	if lobby == nil {
		http.Error(w, "lobby not found", 404)
		return
	}

	res := convert(lobby)
	b, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *server) getConn(w http.ResponseWriter, r *http.Request) {
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

	svc, err := s.tenant.Service(id)
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

	err = svc.JoinPlayer(token, conn)
	if err == lobby.ErrPlayerNotFound {
		http.Error(w, err.Error(), 404)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", 500)
	}
}
