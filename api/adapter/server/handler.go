package server

import (
	"encoding/json"
	"net/http"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"

	"github.com/google/uuid"
)

func (s *server) postLobby(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Strict    bool `json:"strict"`
		MaxPlayer int  `json:"max_player"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := s.tenant.Create(req.Strict, req.MaxPlayer)
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
	Current   [][]int   `json:"current_board"`
	Initial   [][]int   `json:"initial_board"`
	Started   *int64    `json:"started_at"`
	Finished  *int64    `json:"finished_at"`
	History   []history `json:"history"`
	Strict    bool      `json:"strict"`
	MaxPlayer int       `json:"max_player"`
}

type history struct {
	Name      string     `json:"player_name"`
	Artifacts []artifact `json:"artifacts"`
}

type artifact struct {
	Timestamp int64 `json:"timestamp"`
	Row       int   `json:"row"`
	Column    int   `json:"column"`
	Value     int   `json:"value"`
}

func convert(l lobby.Lobby) *response {
	artifacts := l.History().Artifacts()
	histories := make([]history, 0, len(l.Players()))
	for _, p := range l.Players() {
		token := p.Token()
		player := artifacts[token]

		artifacts := make([]artifact, 0, len(player))
		for _, a := range player {
			artifacts = append(artifacts, artifact{
				Timestamp: a.Timestamp(),
				Row:       a.Row(),
				Column:    a.Column(),
				Value:     a.Value(),
			})
		}

		if artifacts == nil {
			artifacts = []artifact{}
		}

		histories = append(histories, history{
			Name:      p.Name(),
			Artifacts: artifacts,
		})
	}

	var current, initial [][]int
	if l.Game().Current() != nil {
		current = l.Game().Current().Int()
	}
	if l.Game().Initial() != nil {
		initial = l.Game().Initial().Int()
	}

	return &response{
		Current:   current,
		Initial:   initial,
		Started:   l.Game().Started(),
		Finished:  l.Game().Finished(),
		History:   histories,
		Strict:    l.Strict(),
		MaxPlayer: l.Size(),
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
