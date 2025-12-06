package database

import (
	"context"
	"fmt"
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
	"sudojo/pkg/sudoku"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	// Closses the connection pool to the database.
	Close()
	// Returns lobby with provided ID or nil if lobby
	// doesn't exist.
	Lobby(id string) (lobby.Lobby, error)
	// Inserts provided lobby in the database. Returns an error
	// if not successful.
	InsertLobby(lobby lobby.Lobby) error
	// Updates the game state and timestamps of provided lobby
	// in the database.
	UpdateLobby(lobby lobby.Lobby) error
	// Inserts provided player attributes in the database under
	// the lobby ID.
	InsertPlayer(lobbyId, token, name string) error
}

type database struct {
	conn *pgxpool.Pool
}

var _ Database = &database{}

// Creates a connection pool to the database. Returns an error
// if no connection could be established.
func New(host, port, name, user, pass string) (*database, error) {
	conn, err := pgxpool.New(
		context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, name),
	)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		conn.Close()
		return nil, err
	}
	return &database{conn: conn}, nil
}

func (db *database) Close() {
	db.conn.Close()
}

func (db *database) Lobby(id string) (lobby.Lobby, error) {
	var start, finish pgtype.Int8
	var initial, current, solution [][]int
	strict, maxPlayer := false, 0

	err := db.conn.QueryRow(
		context.Background(),
		`SELECT initial_board, current_board, solution, started_at, 
		finished_at, strict, max_player FROM lobbies WHERE id = $1;`,
		id,
	).Scan(
		&initial,
		&current,
		&solution,
		&start,
		&finish,
		&strict,
		&maxPlayer,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var started, finished *int64
	if start.Valid {
		started = &start.Int64
	}
	if finish.Valid {
		finished = &finish.Int64
	}

	game := game.New(
		sudoku.NewFromInts(current),
		sudoku.NewFromInts(initial),
		sudoku.NewFromInts(solution),
		started, finished,
	)

	rows, err := db.conn.Query(
		context.Background(),
		`SELECT token, name FROM players WHERE lobby_id = $1;`,
		id,
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	players := make(map[string]string)
	for rows.Next() {
		var token, name string
		if err := rows.Scan(&token, &name); err != nil {
			return nil, err
		}
		players[token] = name
	}

	return lobby.New(id, game, players, strict, maxPlayer), nil
}

func (db *database) InsertLobby(lobby lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO lobbies (
			id, started_at, initial_board, 
			current_board, solution, max_player
		) VALUES ($1, $2, $3, $4, $5, $6);`,
		lobby.Id(),
		lobby.Game().Started(),
		lobby.Game().Initial().Int(),
		lobby.Game().Current().Int(),
		lobby.Game().Solution().Int(),
		lobby.Size(),
	)
	return err
}

func (db *database) UpdateLobby(lobby lobby.Lobby) error {
	_, err := db.conn.Exec(
		context.Background(),
		`UPDATE lobbies SET current_board = $2, finished_at = $3 WHERE id = $1;`,
		lobby.Id(),
		lobby.Game().Current().Int(),
		lobby.Game().Finished(),
	)
	return err
}

func (db *database) InsertPlayer(lobbyId, token, name string) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO players (lobby_id, token, name) 
		VALUES ($1, $2, $3);`,
		lobbyId, token, name,
	)
	return err
}
