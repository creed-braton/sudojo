package database

import "context"

func (db *database) createSchema() error {
	_, err := db.conn.Exec(
		context.Background(),
		`CREATE SCHEMA IF NOT EXISTS sudojo;

		DO $$ BEGIN
			CREATE TYPE sudojo.difficulty AS ENUM 
			('beginner', 'easy', 'medium', 'hard', 'expert', 'extreme', 'joker');
		EXCEPTION
			WHEN duplicate_object THEN NULL;
		END $$;

		CREATE TABLE IF NOT EXISTS sudojo.games (
			hash VARCHAR(16) PRIMARY KEY,
			initial_board INT[9][9] NOT NULL,
			solution INT[9][9] NOT NULL,
			difficulty sudojo.difficulty NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_games_difficulty 
		ON sudojo.games(difficulty);

		CREATE TABLE IF NOT EXISTS sudojo.lobbies (
			id VARCHAR(36) PRIMARY KEY,
			max_player INT NOT NULL,
			strict_mode BOOLEAN NOT NULL DEFAULT TRUE,
			ping_allowed BOOLEAN NOT NULL DEFAULT TRUE,
			notes_allowed BOOLEAN NOT NULL DEFAULT TRUE,
			started_at BIGINT DEFAULT NULL,
			finished_at BIGINT DEFAULT NULL,
			game_hash VARCHAR(16) REFERENCES sudojo.games(hash),
			current_board INT[9][9] NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sudojo.players (
			token VARCHAR(32) PRIMARY KEY,
			lobby_id VARCHAR(36) NOT NULL,
			name VARCHAR(16) NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_players_lobby_id 
		ON sudojo.players(lobby_id);

		CREATE TABLE IF NOT EXISTS sudojo.artifacts (
			lobby_id VARCHAR(36) NOT NULL,
			player_token VARCHAR(32) NOT NULL,
			timestamp BIGINT NOT NULL,
			"row" INT NOT NULL,
			"column" INT NOT NULL,
			"value" INT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_artifacts_lobby_id 
		ON sudojo.artifacts(lobby_id);`,
	)

	return err
}
