export type Sudoku = number[][];

export type PencilMarks = {
  [key: string]: Set<number>; // key format: "row-col"
};
