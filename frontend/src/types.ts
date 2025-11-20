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

export type Position = {
  row: number;
  column: number;
  timestamp: number;
};

export type Conflict = Position & {
  message: string;
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

export type LobbyProps = {
  current: Sudoku | null;
  initial: Sudoku | null;
  insert: (row: number, col: number, val: number) => void;
  ping: (row: number, col: number) => void;
  conflictEvent: Conflict | null;
  pingEvent: Position | null;
  players: Player[];
  maxPlayer: number;
  strict: boolean | null;
};
