import type { ReactElement } from "react";
import styles from "./Board.module.css";
import type { Sudoku } from "../../api/types";
import type { Position } from "../../hooks/sudoku";
import type { Animation } from "../../api/socket";

export type Insertion = {
  row: number;
  column: number;
  value: number;
  playerName: string;
  playerColor: string;
};

const Board = ({
  cursor,
  select,
  initial,
  current,
  notes,
  animations,
  insertions,
}: {
  cursor: Position | null;
  select: (row: number, column: number) => void;
  initial: Sudoku;
  current: Sudoku;
  notes: Map<string, Set<number>>;
  animations: Map<string, Animation>;
  insertions?: Insertion[] | null;
}): ReactElement => {
  const insertionMap = new Map<string, Insertion>();
  if (insertions) {
    for (const insertion of insertions) {
      insertionMap.set(`${insertion.row}-${insertion.column}`, insertion);
    }
  }
  return (
    <div className={`glass-container ${styles.board}`}>
      <div className={styles.sudoku}>
        {current.map((row: number[], rowIndex: number) => (
          <div key={`${rowIndex}`} className={styles.row}>
            {row.map((cell: number, colIndex: number) => {
              const key: string = `${rowIndex}-${colIndex}`;
              const animation: Animation | undefined = animations.get(key);

              const insertion = insertionMap.get(key);
              const hasInsertion = insertions && insertion;

              return (
                <div
                  key={`${rowIndex}-${colIndex}`}
                  className={`
                    ${styles.cell}
                    ${cursor?.row === rowIndex && cursor?.column === colIndex ? styles.cursor : ""}
                    ${!hasInsertion && (initial[rowIndex][colIndex] > 0 ? styles.initialValue : styles.userValue)}
                    ${animation?.type === "conflict" ? styles.conflict : ""}
                    ${animation?.type === "ping" ? styles.ping : ""}
                  `}
                  style={
                    hasInsertion ? { color: insertion.playerColor } : undefined
                  }
                  title={hasInsertion ? insertion.playerName : undefined}
                  onClick={() => {
                    initial && select(rowIndex, colIndex);
                  }}
                >
                  {(() => {
                    if (hasInsertion) {
                      return insertion.value;
                    }
                    if (cell > 0) {
                      return cell;
                    }
                    return (
                      <div className={styles.pencilGrid}>
                        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((value: number) => {
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
                    );
                  })()}
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
};

export default Board;
