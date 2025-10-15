import { useState } from "react";
import useClient, { type ClientProps } from "./useClient";
import type { PencilProps } from "./usePencil";
import usePencil from "./usePencil";
import { ConflictEvent, type Sudoku } from "../types";

export type SudokuProps = {
  connect: (id: string, token: string) => void;
  initialBoard: Sudoku | null;
  currentBoard: Sudoku | null;
  input: (row: number, column: number, value: number) => void;
  conflictEvent: ConflictEvent | undefined;
  notes: Map<string, Set<number>>;
  pencilMode: boolean;
  toggleMode: () => void;
};

const useSudoku = () => {
  const client: ClientProps = useClient();
  const pencil: PencilProps = usePencil(client.initialBoard);
  const [pencilMode, setPencilMode] = useState<boolean>(false);

  const input = (row: number, column: number, value: number): void => {
    if (!client.initialBoard || !client.currentBoard) return;
    if (row > 9 || row < 0 || column > 9 || column < 0) return;
    if (client.initialBoard[row][column] !== 0) return;

    pencilMode
      ? pencil.input(row, column, value)
      : client.input(row, column, value);
  };

  const toggleMode = (): void => {
    setPencilMode(!pencilMode);
  };

  return {
    connect: client.connect,
    initialBoard: client.initialBoard,
    currentBoard: client.currentBoard,
    input,
    conflictEvent: client.conflictEvent,
    notes: pencil.notes,
    pencilMode,
    toggleMode,
  };
};

export default useSudoku;
