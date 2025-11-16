import { type ReactElement } from "react";
import type { Cell, ConflictEvent, Sudoku } from "../../types";
import styles from "./Board.module.css";

const Board = ({
  position,
  setPosition,
  initialBoard,
  currentBoard,
  notes,
  conflictEvent,
}: {
  position: Cell | undefined;
  setPosition: (state: Cell) => void;
  initialBoard: Sudoku;
  currentBoard: Sudoku;
  notes: Map<string, Set<number>>;
  conflictEvent: ConflictEvent | undefined;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.board}`}>
      <div className={styles.sudoku}>
        {currentBoard.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => (
              <div
                key={`${rowIndex}-${colIndex}-${conflictEvent?.timeStamp}`}
                className={`
    ${styles.cell}
    ${position && position.row === rowIndex && position.column === colIndex ? " " + styles.selected : ""}
    ${initialBoard[rowIndex][colIndex] === 0 ? styles.userValue : styles.initialValue}
    ${
      conflictEvent &&
      conflictEvent.cell &&
      conflictEvent.cell[0] === rowIndex &&
      conflictEvent.cell[1] === colIndex
        ? " " + styles.conflict
        : ""
    }
  `}
                onClick={() => {
                  initialBoard &&
                    initialBoard[rowIndex][colIndex] === 0 &&
                    setPosition({
                      row: rowIndex,
                      column: colIndex,
                    } as Cell);
                }}
              >
                {cell > 0 ? (
                  cell
                ) : (
                  <div className={styles.pencilGrid}>
                    {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((value: number) => {
                      const key: string = `${rowIndex}-${colIndex}`;
                      const hasNote: boolean =
                        notes.get(key)?.has(value) || false;
                      return (
                        <div
                          key={value}
                          className={`${styles.pencilCell} ${
                            hasNote ? styles.pencilMark : ""
                          }`}
                        >
                          {hasNote ? value : ""}
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
