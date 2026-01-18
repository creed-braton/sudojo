import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useSocket, type SocketContextProps } from "./socket";
import { useNotes, type NotesContextProps } from "./notes";

export type AnimationType = "conflict" | "ping";

export type Animation = {
  id: string;
  type: AnimationType;
};

export type Cell = {
  value: number;
  initial: boolean;
  animation: Animation | null;
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
  const { id, initial, current, setOnConflict, setOnPing }: SocketContextProps =
    useSocket();
  const { notes }: NotesContextProps = useNotes();

  const [animations, setAnimations] = useState<Map<string, Animation>>(
    new Map(),
  );

  const triggerAnimation = useCallback(
    (type: AnimationType, row: number, column: number): void => {
      const key: string = `${row}-${column}`;
      const id: string = crypto.randomUUID();

      setAnimations((prev): Map<string, Animation> => {
        const next: Map<string, Animation> = new Map(prev);
        next.set(key, { id: id, type });
        return next;
      });

      setTimeout((): void => {
        setAnimations((prev: Map<string, Animation>) => {
          const current: Animation | undefined = prev.get(key);
          if (current?.id !== id) return prev;
          const next: Map<string, Animation> = new Map(prev);
          next.delete(key);
          return next;
        });
      }, 400);
    },
    [],
  );

  const handleConflict = useCallback(
    (row: number, column: number): void => {
      triggerAnimation("conflict", row, column);
    },
    [triggerAnimation],
  );

  const handlePing = useCallback(
    (row: number, column: number): void => {
      triggerAnimation("ping", row, column);
    },
    [triggerAnimation],
  );

  useEffect((): void => {
    setOnConflict(handleConflict);
    setOnPing(handlePing);
  }, []);

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
          animation: animations.get(key) ?? null,
          notes: notes.get(key) ?? new Set(),
        });
      }
    }

    return board;
  }, [initial, current, animations, notes]);

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
