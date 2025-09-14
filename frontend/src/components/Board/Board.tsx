import type { ReactElement } from "react";
import styles from "./styles.module.css";
import Sudoku from "../Sudoku/Sudoku";

type BoardProps = {
  board: number[][];
  initialBoard: number[][];
  selectedCell: { row: number; col: number } | null;
  invalidCell?: { row: number; col: number } | null;
  onCellSelect?: (row: number, col: number) => void;
};

const Board = ({
  board,
  initialBoard,
  selectedCell,
  invalidCell,
  onCellSelect,
}: BoardProps): ReactElement => {
  return (
    <div className={`${styles.board} glassmorphism`}>
      <Sudoku
        board={board}
        initialBoard={initialBoard}
        selectedCell={selectedCell}
        invalidCell={invalidCell}
        onCellSelect={onCellSelect}
      />
    </div>
  );
};

export default Board;
