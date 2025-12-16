export type Sudoku = number[][];

const isSudoku = (value: unknown): value is Sudoku => {
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

export const isFinished = (sudoku: Sudoku): boolean => {
  for (const row of sudoku) {
    for (const cell of row) {
      if (cell === 0) return false;
    }
  }
  return true;
};

export type Player = {
  name: string;
  active: boolean;
};

const isPlayer = (value: unknown): value is Player => {
  if (typeof value !== "object" || value === null) return false;
  const player = value as Record<string, unknown>;
  return typeof player.name === "string" && typeof player.active === "boolean";
};

export type Artifact = {
  timestamp: number;
  row: number;
  column: number;
  value: number;
};

const isArtifact = (value: unknown): value is Artifact => {
  if (typeof value !== "object" || value === null) return false;
  const artifact = value as Record<string, unknown>;
  return (
    typeof artifact.timestamp === "number" &&
    Number.isInteger(artifact.timestamp) &&
    artifact.timestamp >= 0 &&
    typeof artifact.row === "number" &&
    Number.isInteger(artifact.row) &&
    artifact.row >= 0 &&
    artifact.row < 9 &&
    typeof artifact.column === "number" &&
    Number.isInteger(artifact.column) &&
    artifact.column >= 0 &&
    artifact.column < 9 &&
    typeof artifact.value === "number" &&
    Number.isInteger(artifact.value) &&
    artifact.value >= 0 &&
    artifact.value <= 9
  );
};

export type History = {
  player_name: string;
  artifacts: Artifact[];
};

const isHistory = (value: unknown): value is History => {
  if (typeof value !== "object" || value === null) return false;
  const history = value as Record<string, unknown>;
  return (
    typeof history.player_name === "string" &&
    Array.isArray(history.artifacts) &&
    history.artifacts.every(isArtifact)
  );
};

export type Lobby = {
  current_board: Sudoku;
  initial_board: Sudoku;
  max_player: number;
  strict: boolean;
  history: History[];
  started_at: number;
  finished_at: number | null;
};

export const isLobby = (value: unknown): value is Lobby => {
  if (typeof value !== "object" || value === null) return false;
  const lobby = value as Record<string, unknown>;

  if (!("current_board" in lobby)) return false;
  if (!("initial_board" in lobby)) return false;
  if (!("max_player" in lobby)) return false;
  if (!("strict" in lobby)) return false;
  if (!("history" in lobby)) return false;
  if (!("started_at" in lobby)) return false;
  if (!("finished_at" in lobby)) return false;

  if (!isSudoku(lobby.current_board)) return false;
  if (!isSudoku(lobby.initial_board)) return false;

  if (typeof lobby.max_player !== "number") return false;
  if (!Number.isInteger(lobby.max_player)) return false;
  if (lobby.max_player < 1) return false;

  if (typeof lobby.strict !== "boolean") return false;

  if (!Array.isArray(lobby.history)) return false;
  if (!lobby.history.every(isHistory)) return false;

  if (typeof lobby.started_at !== "number") return false;
  if (!Number.isInteger(lobby.started_at)) return false;
  if (lobby.started_at < 0) return false;

  if (lobby.finished_at !== null) {
    if (typeof lobby.finished_at !== "number") return false;
    if (!Number.isInteger(lobby.finished_at)) return false;
    if (lobby.finished_at < 0) return false;
  }

  return true;
};

export type JoinMessage = {
  type: string;
  players: Player[];
};

export const isJoinMessage = (value: unknown): value is JoinMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "join") return false;

  if (!("players" in message)) return false;
  if (!Array.isArray(message.players)) return false;
  if (!message.players.every(isPlayer)) return false;

  return true;
};

export type LeaveMessage = {
  type: string;
  players: Player[];
};

export const isLeaveMessage = (value: unknown): value is LeaveMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "leave") return false;

  if (!("players" in message)) return false;
  if (!Array.isArray(message.players)) return false;
  if (!message.players.every(isPlayer)) return false;

  return true;
};

export type StateMessage = {
  type: string;
  trace?: string;
  current_board: Sudoku;
  initial_board: Sudoku;
  players: Player[];
  max_player: number;
  strict: boolean;
};

export const isStateMessage = (value: unknown): value is StateMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "state") return false;

  if (!("current_board" in message)) return false;
  if (!("initial_board" in message)) return false;
  if (!("players" in message)) return false;
  if (!("max_player" in message)) return false;
  if (!("strict" in message)) return false;

  if ("trace" in message && typeof message.trace !== "string") return false;
  if (!isSudoku(message.current_board)) return false;
  if (!isSudoku(message.initial_board)) return false;
  if (!Array.isArray(message.players)) return false;
  if (!message.players.every(isPlayer)) return false;
  if (typeof message.max_player !== "number") return false;
  if (!Number.isInteger(message.max_player)) return false;
  if (message.max_player < 1) return false;
  if (typeof message.strict !== "boolean") return false;

  return true;
};

export type InsertMessage = {
  type: string;
  trace?: string;
  error?: string;
  conflict?: string;
  current_board?: Sudoku;
  row?: number;
  column?: number;
  value?: number;
};

export const isInsertMessage = (value: unknown): value is InsertMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "insert") return false;

  if ("trace" in message && typeof message.trace !== "string") return false;
  if ("error" in message && typeof message.error !== "string") return false;
  if ("conflict" in message && typeof message.conflict !== "string")
    return false;
  if ("current_board" in message && !isSudoku(message.current_board))
    return false;

  if ("row" in message) {
    if (typeof message.row !== "number") return false;
    if (!Number.isInteger(message.row)) return false;
    if (message.row < 0 || message.row >= 9) return false;
  }

  if ("column" in message) {
    if (typeof message.column !== "number") return false;
    if (!Number.isInteger(message.column)) return false;
    if (message.column < 0 || message.column >= 9) return false;
  }

  if ("value" in message) {
    if (typeof message.value !== "number") return false;
    if (!Number.isInteger(message.value)) return false;
    if (message.value < 0 || message.value > 9) return false;
  }

  return true;
};

export type PingMessage = {
  type: string;
  trace?: string;
  error?: string;
  row?: number;
  column?: number;
};

export const isPingMessage = (value: unknown): value is PingMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "ping") return false;

  if ("trace" in message && typeof message.trace !== "string") return false;
  if ("error" in message && typeof message.error !== "string") return false;

  if ("row" in message) {
    if (typeof message.row !== "number") return false;
    if (!Number.isInteger(message.row)) return false;
    if (message.row < 0 || message.row >= 9) return false;
  }

  if ("column" in message) {
    if (typeof message.column !== "number") return false;
    if (!Number.isInteger(message.column)) return false;
    if (message.column < 0 || message.column >= 9) return false;
  }

  return true;
};

export type SystemMessage = {
  type: string;
  error: string;
};

export const isSystemMessage = (value: unknown): value is SystemMessage => {
  if (typeof value !== "object" || value === null) return false;
  const message = value as Record<string, unknown>;

  if (!("type" in message)) return false;
  if (typeof message.type !== "string") return false;
  if (message.type !== "system") return false;

  if (!("error" in message)) return false;
  if (typeof message.error !== "string") return false;

  return true;
};

export type Position = {
  row: number;
  column: number;
};
