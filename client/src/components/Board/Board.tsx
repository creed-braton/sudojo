import { useState, type ReactElement } from "react";
import { createPortal } from "react-dom";
import styles from "./Board.module.css";
import type { Sudoku } from "../../api/types";
import type { Position } from "../../hooks/sudoku";
import type { Animation } from "../../api/socket";
import DeleteIcon from "../../icons/DeleteIcon";

export type Insertion = {
  row: number;
  column: number;
  value: number;
  playerName: string;
  playerColor: string;
  timestamp: number;
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
  const [tooltip, setTooltip] = useState<{
    text: string;
    x: number;
    y: number;
  } | null>(null);

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
                  onMouseEnter={(e) => {
                    if (hasInsertion) {
                      setTooltip({
                        text: insertion.playerName,
                        x: e.clientX + 10,
                        y: e.clientY - 10,
                      });
                    }
                  }}
                  onMouseMove={(e) => {
                    if (hasInsertion && tooltip) {
                      setTooltip({
                        text: insertion.playerName,
                        x: e.clientX + 10,
                        y: e.clientY - 10,
                      });
                    }
                  }}
                  onMouseLeave={() => setTooltip(null)}
                  onClick={() => {
                    initial && select(rowIndex, colIndex);
                  }}
                >
                  {(() => {
                    if (hasInsertion) {
                      return insertion.value;
                    }
                    if (cell > 0) {
                      const isInitial = initial[rowIndex][colIndex] > 0;
                      const isMistake = insertions && !isInitial && !insertion;
                      if (isMistake) {
                        return (
                          <DeleteIcon
                            style={{
                              display: "flex",
                              justifyContent: "center",
                              alignItems: "center",
                              color: "var(--text-muted)",
                            }}
                          />
                        );
                      }
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
      {tooltip &&
        createPortal(
          <div
            className={styles.tooltip}
            style={{
              left: tooltip.x,
              top: tooltip.y,
            }}
          >
            {tooltip.text}
          </div>,
          document.body,
        )}
    </div>
  );
};

export default Board;
