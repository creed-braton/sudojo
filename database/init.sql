CREATE TABLE IF NOT EXISTS lobbies (
  id VARCHAR(36) PRIMARY KEY,
  max_player INT NOT NULL,
  strict BOOLEAN NOT NULL DEFAULT FALSE,
  started_at BIGINT DEFAULT NULL,
  finished_at BIGINT DEFAULT NULL,
  initial_board INT[9][9] NOT NULL,
  current_board INT[9][9] NOT NULL,
  solution INT[9][9] NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
  lobby_id VARCHAR(36) NOT NULL,
  token VARCHAR(32) PRIMARY KEY,
  name VARCHAR(16) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_players_lobby ON players (lobby_id);
