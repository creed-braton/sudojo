import { useState, type ReactElement } from "react";
import style from "./Game.module.css";
import Info from "../../components/Info/Info";
import Board from "../../components/Board/Board";
import Input from "../../components/Input/Input";
import { useBoard, type BoardContextProps } from "../../providers/board";
import { useInput, type InputContextProps } from "../../providers/input";
import { useSocket, type SocketContextProps } from "../../providers/socket";

const Game = (): ReactElement => {
  const board: BoardContextProps = useBoard();
  const input: InputContextProps = useInput();
  const socket: SocketContextProps = useSocket();
  const [copied, setCopied] = useState(false);

  const copyUrl = (): void => {
    navigator.clipboard.writeText(window.location.href);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={style.game}>
      {board.board !== null && (
        <>
          <div className={style.info}>
            <Info
              players={socket.players}
              maxPlayers={socket.config?.max_player ?? 0}
              onShare={copyUrl}
              copied={copied}
            />
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
