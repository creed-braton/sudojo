import {
  useEffect,
  useRef,
  useState,
  type ReactElement,
  type RefObject,
} from "react";
import Board from "../Board/Board";
import Input from "../Input/Input";
import type { Cell, Position } from "../../providers/board";
import type { InputMode } from "../../providers/input";
import style from "./Game.module.css";

type GameProps = {
  board: Cell[][];
  cursor: Position | null;
  setCursor: (row: number, column: number) => void;
  mode: InputMode;
  togglePing: () => void;
  toggleNotes: () => void;
  input: (value: number) => void;
};

const Game = ({
  board,
  cursor,
  setCursor,
  mode,
  togglePing,
  toggleNotes,
  input,
}: GameProps): ReactElement => {
  const boardRef: RefObject<HTMLTableElement | null> =
    useRef<HTMLTableElement>(null);
  const [boardWidth, setBoardWidth] = useState<number>(0);

  useEffect(() => {
    if (!boardRef.current) return;

    const observer: ResizeObserver = new ResizeObserver((entries) => {
      setBoardWidth(entries[0].contentRect.width);
    });

    observer.observe(boardRef.current);
    return () => observer.disconnect();
  }, [board]);

  return (
    <div className={style.game}>
      <div className={style.board}>
        <Board
          board={board}
          cursor={cursor}
          setCursor={setCursor}
          ref={boardRef}
        />
      </div>
      <div style={{ maxWidth: `${boardWidth}px` }}>
        <Input
          mode={mode}
          togglePing={togglePing}
          toggleNotes={toggleNotes}
          input={input}
        />
      </div>
    </div>
  );
};

export default Game;
