package database

import (
	"context"
	"fmt"
	"sudojo/internal/domain/game"
	"sudojo/internal/domain/lobby"
	"sudojo/internal/domain/sudoku"

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
	var maxPlayer int
	var initial, current, solution [][]int

	err := db.conn.QueryRow(
		context.Background(),
		`SELECT started_at, initial_board, current_board, solution, finished_at, max_player
		 FROM lobbies WHERE id = $1;`,
		id.String(),
	).Scan(
		&created,
		&initial,
		&current,
		&solution,
		&finished,
		&maxPlayer,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	l := &lobby.Lobby{
		Id: id.String(),
		GameManager: &lobby.GameManager{
			Game: &game.Game{
				Initial:  sudoku.Load(initial),
				Current:  sudoku.Load(current),
				Solution: sudoku.Load(solution),
			},
		},
		PlayerManager: &lobby.PlayerManager{
			Players:   make(map[string]*lobby.Player),
			MaxPlayer: maxPlayer,
		},
	}

	if finished.Valid {
		l.GameManager.Game.Finished = &finished.Int64
	}

	rows, err := db.conn.Query(
		context.Background(),
		`SELECT token, name FROM players WHERE lobby_id = $1;`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &lobby.Player{}
		if err := rows.Scan(&p.Token, &p.Name); err != nil {
			return nil, err
		}
		l.PlayerManager.Players[p.Token] = p
	}

	return l, nil
}

func (db *postgres) InsertLobby(l *lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO lobbies (id, started_at, initial_board, current_board, solution) 
		 VALUES ($1, $2, $3, $4, $5);`,
		l.Id,
		l.GameManager.Game.Started,
		l.GameManager.Game.Initial.Int(),
		l.GameManager.Game.Current.Int(),
		l.GameManager.Game.Solution.Int(),
	)
	return err
}

func (db *postgres) UpdateLobby(l *lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`UPDATE lobbies SET current_board = $2, finished_at = $3 WHERE id = $1;`,
		l.Id,
		l.GameManager.Game.Current.Int(),
		l.GameManager.Game.Finished,
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
