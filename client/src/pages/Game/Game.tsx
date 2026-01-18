import { useEffect, type ReactElement } from "react";
import { useBoard, type BoardContextProps } from "../../providers/board";
import { useInput, type InputContextProps } from "../../providers/input";
import { useSocket, type SocketContextProps } from "../../providers/socket";
import Board from "../../components/Board/Board";

const Game = (): ReactElement => {
  const socket: SocketContextProps = useSocket();
  const board: BoardContextProps = useBoard();
  const input: InputContextProps = useInput();

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    socket.setId(id);
  }, [location.pathname]);

  return (
    <div
      style={{
        width: "50%",
        height: "50%",
      }}
    >
      {board.board !== null && (
        <Board
          board={board.board}
          cursor={board.cursor}
          setCursor={input.setCursor}
        />
      )}
    </div>
  );
};

export default Game;
