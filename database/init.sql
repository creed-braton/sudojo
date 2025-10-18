CREATE TABLE IF NOT EXISTS lobbies (
  id VARCHAR(36) PRIMARY KEY,
  created_at BIGINT NOT NULL,
  finished_at BIGINT DEFAULT NULL,
  initial_board INT[9][9] NOT NULL,
  current_board INT[9][9] NOT NULL,
  solution INT[9][9] NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
  lobby_id VARCHAR(36) NOT NULL,
  token VARCHAR(32) PRIMARY KEY,
  name VARCHAR(12) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_players_lobby ON players (lobby_id);

CREATE TABLE IF NOT EXISTS logs (
  lobby_id VARCHAR(36) NOT NULL,
  player_token VARCHAR(32) NOT NULL,
  timestamp BIGINT NOT NULL,
  row INT NOT NULL,
  col INT NOT NULL,
  value INT NOT NULL
);


CREATE INDEX IF NOT EXISTS idx_logs_lobby ON logs (lobby_id);
