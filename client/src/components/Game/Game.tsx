import { type ReactElement } from "react";
import Board from "../Board/Board";
import Input from "../Input/Input";
import type { Cell, Position } from "../../providers/board";
import type { InputMode } from "../../providers/input";
import style from "./Game.module.css";
import Info from "../Info/Info";

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
  return (
    <div className={style.game}>
      <div className={style.info}>
        <Info />
      </div>
      <div className={style.board}>
        <Board board={board} cursor={cursor} setCursor={setCursor} />
      </div>
      <div className={style.input}>
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
