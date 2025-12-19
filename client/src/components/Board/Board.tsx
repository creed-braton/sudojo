import type { ReactElement } from "react";
import styles from "./Board.module.css";
import type { Sudoku } from "../../api/types";
import type { Position } from "../../hooks/sudoku";

const Board = ({
  cursor,
  select,
  initial,
  current,
  notes,
}: {
  cursor: Position | null;
  select: (row: number, column: number) => void;
  initial: Sudoku;
  current: Sudoku;
  notes: Map<string, Set<number>>;
}): ReactElement => {
  return (
    <div className={`glass-container ${styles.board}`}>
      <div className={styles.sudoku}>
        {current.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => (
              <div
                key={`${rowIndex}-${colIndex}`}
                className={`
    ${styles.cell}
    ${cursor && cursor.row === rowIndex && cursor.column === colIndex ? " " + styles.selected : ""}
    ${initial[rowIndex][colIndex] === 0 ? styles.userValue : styles.initialValue}
  `}
                onClick={() => {
                  initial && select(rowIndex, colIndex);
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
