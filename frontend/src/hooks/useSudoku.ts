import { useState } from "react";
import type { LobbyProps, Position } from "../types";
import usePencil, { type PencilProps } from "./usePencil";

export type SudokuProps = LobbyProps & {
  selected: Position | null;
  select: (row: number, column: number) => void;
  input: (value: number) => void;
  notes: Map<string, Set<number>>;
  pencilMode: boolean;
  pingMode: boolean;
  togglePencil: () => void;
  togglePing: () => void;
};

const useSudoku = (lobby: LobbyProps): SudokuProps => {
  const [selected, setSelected] = useState<Position | null>(null);
  const pencil: PencilProps = usePencil(lobby.initial);
  const [pencilMode, setPencilMode] = useState<boolean>(false);
  const [pingMode, setPingMode] = useState<boolean>(false);

  const input = (value: number): void => {
    if (!lobby.initial || !lobby.current) return;
    if (pingMode) return;
    if (!selected) return;
    if (lobby.initial[selected.row][selected.column] !== 0) return;

    pencilMode
      ? pencil.input(selected.row, selected.column, value)
      : lobby.insert(selected.row, selected.column, value);
  };

  const select = (row: number, column: number): void => {
    if (pingMode) {
      lobby.ping(row, column);
    } else {
      if (row > 9 || row < 0 || column > 9 || column < 0) return;
      if (lobby.initial && lobby.initial[row][column] !== 0) return;
      setSelected({ row: row, column: column } as Position);
    }
  };

  const togglePencil = (): void => {
    if (!pencilMode) {
      setPencilMode(true);
      setPingMode(false);
    } else {
      setPencilMode(false);
    }
  };

  const togglePing = (): void => {
    if (!pingMode) {
      setPencilMode(false);
      setPingMode(true);
      setSelected(null);
    } else {
      setPingMode(false);
    }
  };

  return {
    ...lobby,
    selected,
    select,
    input,
    notes: pencil.notes,
    pencilMode,
    pingMode,
    togglePencil,
    togglePing,
  };
};

export default useSudoku;
