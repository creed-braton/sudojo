import { useEffect, useState } from "react";
import type { Sudoku } from "../types";

export type PencilProps = {
  notes: Map<string, Set<number>>;
  input: (row: number, column: number, value: number) => void;
};

const usePencil = (initialBoard: Sudoku | null) => {
  const [notes, setNotes] = useState<Map<string, Set<number>>>(new Map());

  useEffect((): void => {
    setNotes(new Map());
  }, [initialBoard]);

  const input = (row: number, column: number, value: number): void => {
    if (!initialBoard) return;
    if (row > 9 || row < 0 || column > 9 || column < 0) return;
    if (initialBoard[row][column] !== 0) return;

    const key: string = `${row}-${column}`;
    const newNotes: Map<string, Set<number>> = new Map(notes);

    if (value === 0) {
      newNotes.set(key, new Set());
      setNotes(newNotes);
      return;
    }

    const cell = new Set(newNotes.get(key) ?? []);
    if (cell.has(value)) {
      cell.delete(value);
    } else {
      cell.add(value);
    }

    newNotes.set(key, cell);
    setNotes(newNotes);
  };

  return { notes, input };
};

export default usePencil;
