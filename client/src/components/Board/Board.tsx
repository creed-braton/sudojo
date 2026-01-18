import { type ReactElement } from "react";
import type { Cell, Position } from "../../providers/board";
import style from "./Board.module.css";

type BoardProps = {
  board: Cell[][];
  cursor: Position | null;
  setCursor: (row: number, column: number) => void;
};

const Board = ({ board, cursor, setCursor }: BoardProps): ReactElement => {
  return (
    <table className={style.board} role="grid">
      <tbody className={style.body}>
        {board.map((row: Cell[], i: number) => (
          <tr className={style.row} key={`row-${i}`}>
            {row.map((cell: Cell, j: number) => (
              <td
                className={style.cell}
                aria-selected={cursor?.row === i && cursor?.column === j}
                key={`cell-${i}-${j}`}
              >
                <button
                  className={style.button}
                  onClick={() => setCursor(i, j)}
                  type="button"
                >
                  {cell.value !== 0 ? (
                    <span data-initial={cell.initial}>{cell.value}</span>
                  ) : (
                    cell.notes !== undefined && (
                      <div className={style.notes}>
                        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((n) => (
                          <span key={n} className={style.note}>
                            {cell.notes.has(n) ? n : ""}
                          </span>
                        ))}
                      </div>
                    )
                  )}
                </button>
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
};

export default Board;
