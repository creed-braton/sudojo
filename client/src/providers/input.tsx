import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useSocket, type SocketContextProps } from "./socket";
import { useBoard, type BoardContextProps } from "./board";
import { useNotes, type NotesContextProps } from "./notes";

export type InputContextProps = {
  togglePing: () => void;
  toggleNotes: () => void;
  setCursor: (row: number, column: number) => void;
  input: (value: number) => void;
  up: () => void;
  down: () => void;
  left: () => void;
  right: () => void;
};

const InputContext = createContext<InputContextProps | null>(null);

type InputProviderProps = {
  children: ReactNode;
};

type InputMode = "ping" | "notes" | "insert";

export const InputProvider = ({ children }: InputProviderProps) => {
  const socket: SocketContextProps = useSocket();
  const notes: NotesContextProps = useNotes();
  const board: BoardContextProps = useBoard();
  const [mode, setMode] = useState<InputMode>("insert");

  useEffect((): void => {
    setMode("insert");
  }, [socket.id]);

  const togglePing = (): void => {
    if (!socket.config?.pings_allowed) return;
    mode === "ping" ? setMode("insert") : setMode("ping");
    board.setCursor(null);
  };

  const toggleNotes = (): void => {
    if (!socket.config?.notes_allowed) return;
    mode === "notes" ? setMode("insert") : setMode("notes");
  };

  const setCursor = (row: number, column: number): void => {
    if (row < 0 || row > 8 || column < 0 || column > 8) return;
    mode === "ping"
      ? socket.ping(row, column)
      : board.setCursor({ row: row, column: column });
  };

  const input = (value: number): void => {
    if (board.board === null || board.cursor === null || mode === "ping")
      return;
    if (socket.config?.strict_mode && value === 0 && mode !== "notes") return;

    const row: number = board.cursor.row;
    const column: number = board.cursor.column;
    if (
      board.board[row][column].value !== 0 &&
      (mode === "notes" ||
        socket.config?.strict_mode ||
        board.board[row][column].initial)
    )
      return;

    mode === "notes"
      ? notes.insert(row, column, value)
      : socket.insert(row, column, value);
  };

  const up = (): void => {
    if (mode === "ping") return;

    if (board.cursor === null) {
      board.setCursor({ row: 8, column: 8 });
    } else if (board.cursor.row === 0) {
      board.setCursor({ row: 8, column: board.cursor.column });
    } else {
      board.setCursor({
        row: board.cursor.row - 1,
        column: board.cursor.column,
      });
    }
  };

  const down = (): void => {
    if (mode === "ping") return;

    if (board.cursor === null) {
      board.setCursor({ row: 0, column: 0 });
    } else if (board.cursor.row === 8) {
      board.setCursor({ row: 0, column: board.cursor.column });
    } else {
      board.setCursor({
        row: board.cursor.row + 1,
        column: board.cursor.column,
      });
    }
  };

  const left = (): void => {
    if (mode === "ping") return;

    if (board.cursor === null) {
      board.setCursor({ row: 8, column: 8 });
    } else if (board.cursor.column === 0) {
      board.setCursor({ row: board.cursor.row, column: 8 });
    } else {
      board.setCursor({
        row: board.cursor.row,
        column: board.cursor.column - 1,
      });
    }
  };

  const right = (): void => {
    if (mode === "ping") return;

    if (board.cursor === null) {
      board.setCursor({ row: 0, column: 0 });
    } else if (board.cursor.column === 8) {
      board.setCursor({ row: board.cursor.row, column: 0 });
    } else {
      board.setCursor({
        row: board.cursor.row,
        column: board.cursor.column + 1,
      });
    }
  };

  const value: InputContextProps = {
    togglePing,
    toggleNotes,
    setCursor,
    input,
    up,
    down,
    left,
    right,
  };

  return (
    <InputContext.Provider value={value}>{children}</InputContext.Provider>
  );
};

export const useInput = (): InputContextProps => {
  const context = useContext(InputContext);
  if (!context) {
    throw new Error("useInput must be used within a InputProvider");
  }
  return context;
};
