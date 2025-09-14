import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import Sudoku from "./Sudoku";

describe("Sudoku", () => {
  it("renders empty board correctly", () => {
    const emptyBoard = Array(9).fill(Array(9).fill(0));
    const { container } = render(<Sudoku board={emptyBoard} />);

    const sudokuElement = container.querySelector('[class*="sudoku"]');
    expect(sudokuElement).not.toBeNull();

    // All cells should be empty
    const cells = container.querySelectorAll('[class*="cell"]');
    expect(cells.length).toBe(81); // 9x9 grid
    cells.forEach((cell) => {
      expect(cell.textContent).toBe("");
    });
  });

  it("renders numbers correctly", () => {
    const testBoard = [
      [1, 2, 3, 4, 5, 6, 7, 8, 9],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
      [0, 0, 0, 0, 0, 0, 0, 0, 0],
    ];

    const { container } = render(<Sudoku board={testBoard} />);

    // First row should have numbers 1-9
    const rows = container.querySelectorAll('[class*="row"]');
    const firstRow = rows[0];
    const firstRowCells = firstRow.querySelectorAll('[class*="cell"]');
    expect(firstRowCells.length).toBe(9);
    firstRowCells.forEach((cell, index) => {
      expect(cell.textContent).toBe((index + 1).toString());
    });
  });
});
