import { type ReactElement } from "react";
import style from "./Game.module.css";
import Info from "../../components/Info/Info";
import Board from "../../components/Board/Board";
import Input from "../../components/Input/Input";
import { useBoard, type BoardContextProps } from "../../providers/board";
import { useInput, type InputContextProps } from "../../providers/input";

const Game = (): ReactElement => {
  const board: BoardContextProps = useBoard();
  const input: InputContextProps = useInput();

  return (
    <div className={style.game}>
      {board.board !== null && (
        <>
          <div className={style.info}>
            <Info />
          </div>
          <div className={style.board}>
            <Board
              board={board.board}
              cursor={board.cursor}
              setCursor={input.setCursor}
            />
          </div>
          <div className={style.input}>
            <Input
              mode={input.mode}
              togglePing={input.togglePing}
              toggleNotes={input.toggleNotes}
              input={input.input}
            />
          </div>
        </>
      )}
    </div>
  );
};

export default Game;
