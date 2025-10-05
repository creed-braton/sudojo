import { useEffect, useState, type ReactElement } from "react";
import { useLocation } from "react-router-dom";
import type { Sudoku, PencilMarks } from "../../types";
import Board from "../Board/Board";
import styles from "./Lobby.module.css";
import InputBar from "../InputBar/InputBar";

type Position = {
  row: number;
  column: number;
};

const Lobby = ({
  joinLobby,
  sendMove,
  initialState,
  currentState,
}: {
  joinLobby: (id: string) => void;
  sendMove: (row: number, column: number, value: number) => void;
  initialState: Sudoku | undefined;
  currentState: Sudoku | undefined;
}): ReactElement => {
  const [position, setPosition] = useState<Position | undefined>(undefined);
  const [pencilMode, setPencilMode] = useState<boolean>(false);
  const [pencilMarks, setPencilMarks] = useState<PencilMarks>({});
  const location = useLocation();

  const handlePencilInput = (row: number, column: number, value: number) => {
    if (!currentState || !initialState) return;

    // Don't allow pencil marks if there's already a big number in the cell
    if (currentState[row][column] !== 0) return;

    const key = `${row}-${column}`;
    setPencilMarks((prev) => {
      const newMarks = { ...prev };
      if (!newMarks[key]) {
        newMarks[key] = new Set();
      }

      // Toggle the pencil mark
      if (newMarks[key].has(value)) {
        newMarks[key].delete(value);
        if (newMarks[key].size === 0) {
          delete newMarks[key];
        }
      } else {
        newMarks[key].add(value);
      }

      return newMarks;
    });
  };

  const handleMove = (row: number, column: number, value: number) => {
    if (pencilMode) {
      if (value === 0) {
        // Delete button in pencil mode: clear all pencil marks in the cell
        // but only if there's no big number in the cell
        if (currentState && currentState[row][column] === 0) {
          const key = `${row}-${column}`;
          setPencilMarks((prev) => {
            const newMarks = { ...prev };
            delete newMarks[key];
            return newMarks;
          });
        }
        // Don't send delete to server in pencil mode
      } else {
        handlePencilInput(row, column, value);
      }
    } else {
      // Clear pencil marks when placing a big number
      const key = `${row}-${column}`;
      setPencilMarks((prev) => {
        const newMarks = { ...prev };
        delete newMarks[key];
        return newMarks;
      });
      sendMove(row, column, value);
    }
  };

  useEffect((): void => {
    const id = location.pathname.split("/")[2];
    joinLobby(id);
  }, [location.pathname]);

  return (
    <div className={styles.lobby}>
      {initialState && currentState && (
        <>
          <Board
            position={position}
            setPosition={setPosition}
            initialState={initialState}
            currentState={currentState}
            pencilMarks={pencilMarks}
          />
          <InputBar
            position={position}
            sendMove={handleMove}
            pencilMode={pencilMode}
            setPencilMode={setPencilMode}
          />
        </>
      )}
    </div>
  );
};

export default Lobby;
export type { Position };
