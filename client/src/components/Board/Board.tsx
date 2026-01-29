import {
  useEffect,
  useRef,
  useState,
  type ReactElement,
  type RefObject,
} from "react";
import type { Cell, Position } from "../../providers/board";
import style from "./Board.module.css";

type BoardProps = {
  board: Cell[][];
  cursor: Position | null;
  setCursor: (row: number, column: number) => void;
};

const Board = ({ board, cursor, setCursor }: BoardProps): ReactElement => {
  const ref: RefObject<HTMLTableElement | null> = useRef(null);
  const [compact, setCompact] = useState<boolean>(true);

  useEffect((): (() => void) | void => {
    const table: HTMLTableElement | null = ref.current;
    if (table === null) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setCompact(entry.contentRect.width < 450);
      }
    });

    observer.observe(table);
    return () => observer.disconnect();
  }, []);

  return (
    <table ref={ref} className={style.board} data-compact={compact} role="grid">
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
