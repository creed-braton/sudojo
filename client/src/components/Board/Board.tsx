import { type ReactElement, type RefObject } from "react";
import type { Cell, Position } from "../../providers/board";
import style from "./Board.module.css";

type BoardProps = {
  board: Cell[][];
  cursor: Position | null;
  setCursor: (row: number, column: number) => void;
  ref: RefObject<HTMLTableElement | null>;
};

const Board = ({ board, cursor, setCursor, ref }: BoardProps): ReactElement => {
  return (
    <table ref={ref} className={style.board} role="grid">
      <tbody className={style.body}>
        {board.map((row: Cell[], i: number) => (
          <tr className={style.row} key={i}>
            {row.map((cell: Cell, j: number) => (
              <td
                className={style.cell}
                aria-selected={
                  cell.animation === null &&
                  cursor?.row === i &&
                  cursor?.column === j
                }
                data-animation={cell.animation?.type}
                key={j}
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
