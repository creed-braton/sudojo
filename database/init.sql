CREATE TYPE difficulty AS ENUM ('easy', 'medium', 'hard', 'extreme', 'joker');

CREATE TABLE IF NOT EXISTS games (
  hash VARCHAR(16) PRIMARY KEY,
  initial_board INT[9][9] NOT NULL,
  solution INT[9][9] NOT NULL,
  difficulty difficulty NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_games_difficulty ON games(difficulty);

CREATE TABLE IF NOT EXISTS lobbies (
  id VARCHAR(36) PRIMARY KEY,
  max_player INT NOT NULL,
  strict BOOLEAN NOT NULL DEFAULT FALSE,
  started_at BIGINT DEFAULT NULL,
  finished_at BIGINT DEFAULT NULL,
  game_hash VARCHAR(16) REFERENCES games(hash),
  current_board INT[9][9] NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
  token VARCHAR(32) PRIMARY KEY,
  lobby_id VARCHAR(36) NOT NULL,
  name VARCHAR(16) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_players_lobby_id ON players(lobby_id);

CREATE TABLE IF NOT EXISTS artifacts (
  lobby_id VARCHAR(36) NOT NULL,
  player_token VARCHAR(32) NOT NULL,
  timestamp BIGINT NOT NULL,
  "row" INT NOT NULL,
  "column" INT NOT NULL,
  "value" INT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_artifacts_lobby_id ON artifacts(lobby_id);
