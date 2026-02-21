package server

import (
	"encoding/json"
	"sudojo/pkg/lobby"
)

type postLobbyReq struct {
	Strict     bool   `json:"strict_mode"`
	Pings      bool   `json:"pings_allowed"`
	Notes      bool   `json:"notes_allowed"`
	MaxSize    int    `json:"max_player"`
	Difficulty string `json:"difficulty"`
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

type config struct {
	Strict  bool `json:"strict_mode"`
	Pings   bool `json:"pings_allowed"`
	Notes   bool `json:"notes_allowed"`
	MaxSize int  `json:"max_player"`
}

type getLobbyRes struct {
	Current    [][]int   `json:"current_board"`
	Initial    [][]int   `json:"initial_board"`
	Started    *int64    `json:"started_at"`
	Finished   *int64    `json:"finished_at"`
	History    []history `json:"history"`
	Config     config    `json:"config"`
	MaxPlayer  int       `json:"max_player"`
	Difficulty string    `json:"difficulty"`
}

func marshalLobby(l lobby.Lobby) ([]byte, error) {
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

	config := config{
		Strict:  l.Config().Strict(),
		Pings:   l.Config().Pings(),
		Notes:   l.Config().Notes(),
		MaxSize: l.Config().MaxSize(),
	}

	res := &getLobbyRes{
		Current:    current,
		Initial:    initial,
		Started:    l.Game().StartedAt(),
		Finished:   l.Game().FinishedAt(),
		History:    histories,
		Config:     config,
		Difficulty: l.Game().Difficulty(),
	}

	return json.Marshal(res)
}
