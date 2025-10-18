package database

import (
	"context"
	"fmt"
	"sudojo/domain/game"
	"sudojo/domain/lobby"
	"sudojo/domain/sudoku"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	conn *pgxpool.Pool
}

func New(host, port, name, user, pass string) (*postgres, error) {
	conn, err := pgxpool.New(
		context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name),
	)
	if err != nil {
		return nil, err
	}
	return &postgres{conn: conn}, nil
}

func (db *postgres) Close() {
	db.conn.Close()
}

func (db *postgres) Lobby(id uuid.UUID) (*lobby.Lobby, error) {
	var created int64
	var finished pgtype.Int8
	var initial, current, solution [][]int

	err := db.conn.QueryRow(
		context.Background(),
		`SELECT created_at, initial_board, current_board, 
		solution, finished_at FROM lobbies WHERE id = $1;`,
		id.String(),
	).Scan(
		&created,
		&initial,
		&current,
		&solution,
		&finished,
	)

	if err != nil && err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	l := &lobby.Lobby{
		Id: id.String(),
		Game: &game.Game{
			Id:       id,
			Created:  created,
			Initial:  sudoku.Load(initial),
			Current:  sudoku.Load(current),
			Solution: sudoku.Load(solution),
		},
		Players: make(map[string]*lobby.Player),
	}
	if finished.Valid {
		l.Game.Finished = &finished.Int64
	}

	rows, err := db.conn.Query(
		context.Background(),
		`SELECT token, name FROM players WHERE lobby_id = $1;`,
		id,
	)
	defer rows.Close()

	if err != nil {
		return nil, nil
	}

	for rows.Next() {
		p := &lobby.Player{}
		if err := rows.Scan(&p.Token, &p.Name); err != nil {
			return nil, err
		}
		l.Players[p.Token] = p
	}

	return l, nil
}

func (db *postgres) InsertLobby(lobby *lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO lobbies (id, created_at, initial_board, current_board, solution) 
		VALUES ($1, $2, $3, $4, $5);`,
		lobby.Id,
		lobby.Game.Created,
		lobby.Game.Initial.Int(),
		lobby.Game.Current.Int(),
		lobby.Game.Solution.Int(),
	)
	return err
}

func (db *postgres) UpdateLobby(lobby *lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`UPDATE lobbies SET current_board = $2, finished_at = $3 WHERE id = $1;`,
		lobby.Id,
		lobby.Game.Current.Int(),
		lobby.Game.Finished,
	)
	return err
}

func (db *postgres) InsertPlayer(id string, player *lobby.Player) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO players (lobby_id, token, name) 
		VALUES ($1, $2, $3);`,
		id,
		player.Token,
		player.Name,
	)
	return err
}

func (db *postgres) Logs(id string) ([]*lobby.Log, error) {
	rows, err := db.conn.Query(
		context.Background(),
		`SELECT player_token, timestamp, row, col, value FROM logs 
		WHERE lobby_id = $1 ORDER BY timestamp ASC;`,
		id,
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	logs := []*lobby.Log{}
	for rows.Next() {
		l := &lobby.Log{LobbyId: id}
		if err := rows.Scan(
			&l.Player, &l.Time, &l.Row,
			&l.Column, &l.Value,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (db *postgres) InsertLogs(logs []*lobby.Log) error {
	if len(logs) < 1 {
		return nil
	}

	rows := make([][]interface{}, 0, len(logs))
	for _, log := range logs {
		rows = append(rows, []interface{}{
			log.LobbyId,
			log.Player,
			log.Time,
			log.Row,
			log.Column,
			log.Value,
		})
	}

	_, err := db.conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"logs"},
		[]string{"lobby_id", "player_token", "timestamp", "row", "col", "value"},
		pgx.CopyFromRows(rows),
	)
	return err
}
