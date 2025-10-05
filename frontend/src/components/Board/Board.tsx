import type { ReactElement } from "react";
import type { Position } from "../Lobby/Lobby";
import type { Sudoku, PencilMarks } from "../../types";
import styles from "./Board.module.css";

const Board = ({
  position,
  setPosition,
  initialState,
  currentState,
  pencilMarks,
}: {
  position: Position | undefined;
  setPosition: (state: Position) => void;
  initialState: Sudoku;
  currentState: Sudoku;
  pencilMarks: PencilMarks;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.board}`}>
      <div className={styles.sudoku}>
        {currentState.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => (
              <div
                key={`${rowIndex}-${colIndex}`}
                className={`${styles.cell}${
                  position &&
                  position.row === rowIndex &&
                  position.column === colIndex
                    ? " " + styles.selected
                    : ""
                } ${
                  initialState[rowIndex][colIndex] === 0
                    ? styles.userValue
                    : styles.initialValue
                }`}
                onClick={() => {
                  initialState &&
                    initialState[rowIndex][colIndex] === 0 &&
                    setPosition({
                      row: rowIndex,
                      column: colIndex,
                    } as Position);
                }}
              >
                {cell > 0 ? (
                  cell
                ) : (
                  <div className={styles.pencilGrid}>
                    {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => {
                      const key = `${rowIndex}-${colIndex}`;
                      const hasPencilMark = pencilMarks[key]?.has(num);
                      return (
                        <div
                          key={num}
                          className={`${styles.pencilCell} ${
                            hasPencilMark ? styles.pencilMark : ""
                          }`}
                        >
                          {hasPencilMark ? num : ""}
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
};

export default Board;
