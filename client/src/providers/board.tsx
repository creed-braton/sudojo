import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useSocket, type SocketContextProps } from "./socket";
import { useNotes, type NotesContextProps } from "./notes";

export type Cell = {
  value: number;
  initial: boolean;
  notes: Set<number>;
};

export type Position = {
  row: number;
  column: number;
};

export type BoardContextProps = {
  board: Cell[][] | null;
  cursor: Position | null;
  setCursor: (state: Position | null) => void;
};

const BoardContext = createContext<BoardContextProps | null>(null);

type BoardProviderProps = {
  children: ReactNode;
};

export const BoardProvider = ({ children }: BoardProviderProps) => {
  const { id, initial, current }: SocketContextProps = useSocket();
  const { notes }: NotesContextProps = useNotes();

  const board: Cell[][] | null = useMemo(() => {
    if (initial === null || current === null) return null;
    const board: Cell[][] = [];

    for (let row: number = 0; row < 9; row++) {
      board.push([]);
      for (let col: number = 0; col < 9; col++) {
        const key = `${row}-${col}`;
        board[row].push({
          value: current[row][col],
          initial: initial[row][col] !== 0,
          notes: notes.get(key) ?? new Set(),
        });
      }
    }

    return board;
  }, [initial, current, notes]);

  const [cursor, setCursor] = useState<Position | null>(null);

  useEffect((): void => {
    setCursor(null);
  }, [id]);

  const value: BoardContextProps = {
    board,
    cursor,
    setCursor,
  };

  return (
    <BoardContext.Provider value={value}>{children}</BoardContext.Provider>
  );
};

export const useBoard = (): BoardContextProps => {
  const context = useContext(BoardContext);
  if (!context) {
    throw new Error("useBoard must be used within a BoardProvider");
  }
  return context;
};
