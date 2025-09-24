import type { ReactElement } from "react";
import type { Position } from "../Lobby/Lobby";
import type { Sudoku } from "../../types";
import styles from "./Board.module.css";

const Board = ({
  position,
  setPosition,
  initialState,
  currentState,
}: {
  position: Position | undefined;
  setPosition: (state: Position) => void;
  initialState: Sudoku;
  currentState: Sudoku;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.board}`}>
      <div className={styles.sudoku}>
        {currentState.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => (
              <div
                key={`${rowIndex}-${colIndex}`}
                className={`${styles.cell}${position && position.row === rowIndex && position.column === colIndex ? " " + styles.selected : ""} ${initialState[rowIndex][colIndex] === 0 ? styles.userValue : styles.initialValue}`}
                onClick={() => {
                  initialState &&
                    initialState[rowIndex][colIndex] === 0 &&
                    setPosition({
                      row: rowIndex,
                      column: colIndex,
                    } as Position);
                }}
              >
                {cell > 0 && cell}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
};

export default Board;
