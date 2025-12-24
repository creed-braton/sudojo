package database

import (
	"context"
	"fmt"
	"sudojo/pkg/game"
	"sudojo/pkg/history"
	"sudojo/pkg/lobby"
	"sudojo/pkg/player"
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
	InsertPlayer(id, token, name string) error
	// Returns a random game with the provided difficulty.
	SampleGame(difficulty string) (game.Game, error)
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

func (db *database) players(id string) (map[string]player.Player, error) {
	rows, err := db.conn.Query(
		context.Background(),
		`SELECT token, name FROM players WHERE lobby_id = $1;`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make(map[string]player.Player)
	for rows.Next() {
		var token, name string
		if err := rows.Scan(&token, &name); err != nil {
			return nil, err
		}
		players[token] = player.New(token, name)
	}

	return players, nil
}

func (db *database) history(id string) (history.History, error) {
	rows, err := db.conn.Query(
		context.Background(),
		`SELECT player_token, timestamp, "row", "column", "value"
		 FROM artifacts 
		 WHERE lobby_id = $1 
		 ORDER BY timestamp ASC;`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artifacts := make([]history.Artifact, 0)
	for rows.Next() {
		var playerToken string
		var timestamp int64
		var row, column, value int
		if err := rows.Scan(&playerToken, &timestamp, &row, &column, &value); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, history.NewArtifact(playerToken, timestamp, row, column, value))
	}

	return history.New(artifacts), nil
}

func (db *database) Lobby(id string) (lobby.Lobby, error) {
	var difficulty string
	var start, finish pgtype.Int8
	var initial, current, solution [][]int
	strict, maxPlayer := false, 0
	err := db.conn.QueryRow(
		context.Background(),
		`SELECT games.initial_board, 
       lobbies.current_board, 
       games.solution, 
			 games.difficulty,
       lobbies.started_at, 
       lobbies.finished_at, 
       lobbies.strict, 
       lobbies.max_player 
		FROM lobbies 
		JOIN games ON lobbies.game_hash = games.hash
		WHERE lobbies.id = $1;`,
		id,
	).Scan(
		&initial,
		&current,
		&solution,
		&difficulty,
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
		started, finished, difficulty,
	)

	players, err := db.players(id)
	if err != nil {
		return nil, err
	}

	history, err := db.history(id)
	if err != nil {
		return nil, err
	}

	return lobby.New(id, game, players, history, strict, maxPlayer), nil
}

func (db *database) InsertLobby(lobby lobby.Lobby) error {
	hash := lobby.Game().Hash()
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO games (hash, initial_board, solution, difficulty) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (hash) DO NOTHING;`,
		hash,
		lobby.Game().Initial().Int(),
		lobby.Game().Solution().Int(),
		lobby.Game().Difficulty(),
	)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec(
		context.Background(),
		`INSERT INTO lobbies (
			id, max_player, strict, started_at, 
			game_hash, current_board
		) VALUES ($1, $2, $3, $4, $5, $6);`,
		lobby.Id(),
		lobby.Size(),
		lobby.Strict(),
		lobby.Game().Started(),
		hash,
		lobby.Game().Current().Int(),
	)
	return err
}

func (db *database) UpdateLobby(lobby lobby.Lobby) error {
	ctx := context.Background()
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`UPDATE lobbies SET current_board = $2, finished_at = $3 WHERE id = $1;`,
		lobby.Id(),
		lobby.Game().Current().Int(),
		lobby.Game().Finished(),
	)
	if err != nil {
		return err
	}

	artifacts := lobby.History().Flush()
	if len(artifacts) > 0 {
		batch := &pgx.Batch{}
		for _, artifact := range artifacts {
			batch.Queue(
				`INSERT INTO artifacts (lobby_id, player_token, 
				 timestamp, "row", "column", "value") 
				 VALUES ($1, $2, $3, $4, $5, $6);`,
				lobby.Id(),
				artifact.Player(),
				artifact.Timestamp(),
				artifact.Row(),
				artifact.Column(),
				artifact.Value(),
			)
		}
		br := tx.SendBatch(ctx, batch)
		err = br.Close()
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (db *database) InsertPlayer(id, token, name string) error {
	_, err := db.conn.Exec(
		context.Background(),
		`INSERT INTO players (lobby_id, token, name) 
		VALUES ($1, $2, $3);`,
		id, token, name,
	)
	return err
}

func (db *database) SampleGame(difficulty string) (game.Game, error) {
	var initial, solution [][]int
	err := db.conn.QueryRow(
		context.Background(),
		`SELECT initial_board, solution 
		 FROM games 
		 WHERE difficulty = $1 
		 ORDER BY RANDOM() 
		 LIMIT 1;`,
		difficulty,
	).Scan(&initial, &solution)
	if err == pgx.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return game.New(
		sudoku.NewFromInts(initial), // current starts as initial
		sudoku.NewFromInts(initial),
		sudoku.NewFromInts(solution),
		nil, nil, difficulty,
	), nil
}
