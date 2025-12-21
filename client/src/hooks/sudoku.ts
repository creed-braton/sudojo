import { useEffect, useState } from "react";
import { isFinished, type Sudoku } from "../api/types";
import usePencil, { type PencilProps } from "./pencil";

export type Position = {
  row: number;
  column: number;
};

export type Mode = "pencil" | "ping" | "default";

export type SudokuProps = {
  cursor: Position | null;
  mode: Mode;
  select: (row: number, column: number) => void;
  input: (value: number) => void;
  notes: Map<string, Set<number>>;
  togglePencil: () => void;
  togglePing: () => void;
  finished: boolean;
};

const useSudoku = (
  initial: Sudoku | undefined,
  current: Sudoku | undefined,
  insert: (row: number, column: number, value: number) => void,
  ping: (row: number, column: number) => void,
): SudokuProps => {
  const pencil: PencilProps = usePencil(initial);
  const [cursor, setCursor] = useState<Position | null>(null);
  const [mode, setMode] = useState<Mode>("default");
  const [finished, setFinished] = useState<boolean>(false);

  const select = (row: number, column: number): void => {
    if (mode === "ping") {
      ping(row, column);
    } else {
      if (row > 9 || row < 0 || column > 9 || column < 0) return;
      if (initial && initial[row][column] !== 0) return;
      setCursor({ row: row, column: column } as Position);
    }
  };

  const input = (value: number): void => {
    if (!initial || !current || finished) return;
    if (mode === "ping") return;
    if (!cursor) return;
    if (initial[cursor.row][cursor.column] !== 0) return;

    mode === "pencil"
      ? pencil.input(cursor.row, cursor.column, value)
      : insert(cursor.row, cursor.column, value);
  };

  const togglePencil = (): void =>
    mode !== "pencil" ? setMode("pencil") : setMode("default");

  const togglePing = (): void => {
    if (mode !== "ping") {
      setMode("ping");
      setCursor(null);
    } else {
      setMode("default");
    }
  };

  useEffect((): void => {
    current && isFinished(current) && setFinished(true);
  }, [current]);

  return {
    cursor,
    mode,
    select,
    input,
    notes: pencil.notes,
    togglePencil,
    togglePing,
    finished,
  };
};

export default useSudoku;
