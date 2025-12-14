export type Sudoku = number[][];

export const isSudoku = (value: unknown): value is Sudoku => {
  if (!Array.isArray(value)) return false;
  if (value.length !== 9) return false;

  for (const row of value) {
    if (!Array.isArray(row)) return false;
    if (row.length !== 9) return false;

    for (const cell of row) {
      if (typeof cell !== "number") return false;
      if (!Number.isInteger(cell)) return false;
      if (cell < 0 || cell > 9) return false;
    }
  }

  return true;
};

export type Player = {
  name: string;
  active: boolean;
};

export const isPlayer = (value: unknown): value is Player => {
  if (typeof value !== "object" || value === null) return false;
  const player = value as Record<string, unknown>;
  return typeof player.name === "string" && typeof player.active === "boolean";
};

export type Lobby = {
  current_state: Sudoku;
  initial_state: Sudoku;
  players: Player[];
  max_player: number;
  strict: boolean;
};

export const isLobby = (value: unknown): value is Lobby => {
  if (typeof value !== "object" || value === null) return false;
  const lobby = value as Record<string, unknown>;

  if (!("current_state" in lobby)) return false;
  if (!("initial_state" in lobby)) return false;
  if (!("players" in lobby)) return false;
  if (!("max_player" in lobby)) return false;
  if (!("strict" in lobby)) return false;

  if (!isSudoku(lobby.current_state)) return false;
  if (!isSudoku(lobby.initial_state)) return false;

  if (!Array.isArray(lobby.players)) return false;
  if (!lobby.players.every(isPlayer)) return false;

  if (typeof lobby.max_player !== "number") return false;
  if (!Number.isInteger(lobby.max_player)) return false;
  if (lobby.max_player < 1) return false;

  if (typeof lobby.strict !== "boolean") return false;

  return true;
};
