package lobby

import (
	"encoding/json"
	"fmt"
	"sudojo/domain/sudoku"
)

type Log struct {
	LobbyId string `json:"-"`
	Row     int    `json:"row"`
	Column  int    `json:"column"`
	Value   int    `json:"value"`
	Player  string `json:"-"`
	Time    int64  `json:"timestamp"`
}

type Score struct {
	PlayerName string `json:"player_name"`
	Points     []*Log `json:"points"`
	Mistakes   []*Log `json:"mistakes"`
}

type Summary struct {
	Current  *sudoku.Sudoku `json:"board"`
	Created  int64          `json:"created_at"`
	Finished *int64         `json:"finished_at"`
	Scores   []*Score       `json:"scores"`
}

func (l *Lobby) Summary(logs []*Log) ([]byte, error) {
	check := make(map[string]struct{})
	sum := &Summary{
		Current:  l.Game.Current,
		Created:  l.Game.Created,
		Finished: l.Game.Finished,
	}
	players := make(map[string]*Score)
	for t, p := range l.Players {
		players[t] = &Score{
			PlayerName: p.Name,
			Mistakes:   []*Log{},
			Points:     []*Log{},
		}
	}

	for _, log := range logs {
		if log.Value == sudoku.EmptyCell {
			continue
		}
		correct := l.Game.Solution[log.Row][log.Column] == log.Value
		id := fmt.Sprintf(
			"%d%d%d%t",
			log.Row,
			log.Column,
			log.Value,
			correct,
		)
		if _, exist := check[id]; !exist {
			check[id] = struct{}{}
			if correct {
				players[log.Player].Points = append(
					players[log.Player].Points, log,
				)
			} else {
				players[log.Player].Mistakes = append(
					players[log.Player].Mistakes, log,
				)
			}
		}
	}

	for _, s := range players {
		sum.Scores = append(sum.Scores, s)
	}

	return json.Marshal(sum)
}
