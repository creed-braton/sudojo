/**
 * Checks if placing a value at a specific position in the Sudoku board is valid
 * according to Sudoku rules (no duplicates in row, column, or 3x3 box).
 *
 * @param board The current Sudoku board
 * @param row Row index (0-8)
 * @param col Column index (0-8)
 * @param value Value to check (1-9)
 * @returns true if the move is valid, false otherwise
 */
export const isValidSudokuMove = (
  board: number[][],
  row: number,
  col: number,
  value: number
): boolean => {
  // Value 0 is always valid (clearing a cell)
  if (value === 0) {
    return true;
  }

  // Check if value is between 1 and 9
  if (value < 1 || value > 9) {
    return false;
  }

  // Check row
  for (let c = 0; c < 9; c++) {
    if (c !== col && board[row][c] === value) {
      return false; // Value already exists in this row
    }
  }

  // Check column
  for (let r = 0; r < 9; r++) {
    if (r !== row && board[r][col] === value) {
      return false; // Value already exists in this column
    }
  }

  // Check 3x3 box
  const boxStartRow = Math.floor(row / 3) * 3;
  const boxStartCol = Math.floor(col / 3) * 3;

  for (let r = boxStartRow; r < boxStartRow + 3; r++) {
    for (let c = boxStartCol; c < boxStartCol + 3; c++) {
      if ((r !== row || c !== col) && board[r][c] === value) {
        return false; // Value already exists in this 3x3 box
      }
    }
  }

  // The move is valid
  return true;
};
