import type { ReactElement } from "react";
import type { Conflict, Position, Sudoku } from "../../types";
import styles from "./Board.module.css";

const Board = ({
  selected,
  select,
  initialBoard,
  currentBoard,
  notes,
  conflictEvent,
  pingEvent,
}: {
  selected: Position | null;
  select: (row: number, column: number) => void;
  initialBoard: Sudoku;
  currentBoard: Sudoku;
  notes: Map<string, Set<number>>;
  conflictEvent: Conflict | null;
  pingEvent: Position | null;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.board}`}>
      <div className={styles.sudoku}>
        {currentBoard.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => (
              <div
                key={`${rowIndex}-${colIndex}-${conflictEvent?.timestamp}-${pingEvent?.timestamp}`}
                className={`
    ${styles.cell}
    ${selected && selected.row === rowIndex && selected.column === colIndex ? " " + styles.selected : ""}
    ${initialBoard[rowIndex][colIndex] === 0 ? styles.userValue : styles.initialValue}
    ${
      conflictEvent &&
      conflictEvent.row === rowIndex &&
      conflictEvent.column === colIndex
        ? " " + styles.conflict
        : ""
    }
    ${
      pingEvent && pingEvent.row === rowIndex && pingEvent.column === colIndex
        ? " " + styles.ping
        : ""
    }
  `}
                onClick={() => {
                  initialBoard && select(rowIndex, colIndex);
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
