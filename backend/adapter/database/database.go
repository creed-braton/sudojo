package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	Close()
}

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
