package database

import (
	"context"
	"fmt"
	"sudojo/domain/lobby"

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
