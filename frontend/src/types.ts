export class ApiError extends Error {
  status: number;
  message: string;
  constructor(status: number, message: string) {
    super();
    this.status = status;
    this.message = message;
  }
}

export class ConflictEvent extends Event {
  cell: number[];
  message: string;
  constructor(cell: number[], message: string) {
    super("conflict");
    this.cell = cell;
    this.message = message;
  }
}

export type Sudoku = number[][];

export type Cell = {
  row: number;
  column: number;
};

export type Message = {
  type: string;
  initial_state: Sudoku | null;
  current_state: Sudoku | null;
  error: string | null;
  conflict: string | null;
  cell: number[];
};

type Log = {
  timestamp: number;
  row: number;
  column: number;
  value: number;
};

export type Score = {
  points: Log[];
  mistakes: Log[];
  player_name: string;
};

export type GameStats = {
  board: Sudoku;
  created_at: number;
  finished_at: number | null;
  scores: Score[];
};

export type Points = {
  player: string;
  points: number;
};
