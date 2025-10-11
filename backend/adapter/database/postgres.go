package database

import (
	"context"
	"fmt"
	"sudojo/domain/lobby"

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

func (db *postgres) Lobby(id uuid.UUID, logger chan *lobby.Log) (*lobby.Lobby, error) {
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

	var l *lobby.Lobby
	if finished.Valid {
		l = lobby.Load(
			id, logger, created, &finished.Int64,
			initial, current, solution,
		)
	} else {
		l = lobby.Load(
			id, logger, created, nil,
			initial, current, solution,
		)
	}

	rows, err := db.conn.Query(
		context.Background(),
		`SELECT token, name FROM players WHERE lobby_id = $1;`,
		id.String(),
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		p := &lobby.Player{Out: make(chan []byte)}
		if err := rows.Scan(&p.Token, &p.Name); err != nil {
			return nil, err
		}
		l.Load(p)
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
