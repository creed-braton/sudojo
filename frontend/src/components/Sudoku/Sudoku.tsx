import type { ReactElement } from "react";
import styles from "./styles.module.css";

type SudokuProps = {
  board: number[][];
  initialBoard?: number[][];
  selectedCell?: { row: number; col: number } | null;
  invalidCell?: { row: number; col: number } | null;
  onCellSelect?: (row: number, col: number) => void;
};

const Sudoku = ({
  board,
  initialBoard = Array(9).fill(Array(9).fill(0)),
  selectedCell,
  invalidCell,
  onCellSelect,
}: SudokuProps): ReactElement => {
  const handleCellClick = (rowIndex: number, colIndex: number) => {
    if (onCellSelect) {
      onCellSelect(rowIndex, colIndex);
    }
  };

  return (
    <div className={styles.sudoku}>
      {board.map((row, rowIndex) => (
        <div key={`row-${rowIndex}`} className={styles.row}>
          {row.map((cell, colIndex) => {
            const isSelected =
              selectedCell?.row === rowIndex && selectedCell?.col === colIndex;
            const isInvalid =
              invalidCell?.row === rowIndex && invalidCell?.col === colIndex;
            const isInitialValue =
              initialBoard.length > rowIndex &&
              initialBoard[rowIndex].length > colIndex &&
              initialBoard[rowIndex][colIndex] !== 0;

            return (
              <div
                key={`cell-${rowIndex}-${colIndex}`}
                className={`${styles.cell} ${
                  (Math.floor(rowIndex / 3) + Math.floor(colIndex / 3)) % 2 ===
                  0
                    ? styles.lightBlock
                    : styles.darkBlock
                } ${
                  colIndex === 2 || colIndex === 5 ? styles.rightBorder : ""
                } ${
                  rowIndex === 2 || rowIndex === 5 ? styles.bottomBorder : ""
                } ${isSelected ? styles.selected : ""}
                  ${isInvalid ? styles.invalidMove : ""}
                  ${isInitialValue ? styles.initialValue : styles.userValue}`}
                onClick={() => handleCellClick(rowIndex, colIndex)}
              >
                {cell !== 0 ? cell : ""}
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
};

export default Sudoku;
