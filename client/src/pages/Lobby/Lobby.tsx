import { useEffect, useState, type ReactElement } from "react";
import { getPlayer, HttpError, postPlayer } from "../../api/api";
import { useSocket, type SocketContextProps } from "../../providers/socket";
import { useBoard, type BoardContextProps } from "../../providers/board";
import { useInput, type InputContextProps } from "../../providers/input";
import Game from "../../components/Game/Game";

const Lobby = (): ReactElement => {
  const socket: SocketContextProps = useSocket();
  const board: BoardContextProps = useBoard();
  const input: InputContextProps = useInput();
  const [_, setError] = useState<string | undefined>(undefined);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    getPlayer(id)
      .then((): void => socket.setId(id))
      .catch((error: Error) => {
        if (error instanceof HttpError && error.status == 423) {
        } else if (error instanceof HttpError && error.status == 401) {
          postPlayer(id)
            .then(() => socket.setId(id))
            .catch((error: Error) => setError(error.message));
        } else {
          setError(error.message);
        }
      });
  }, [location.pathname]);

  return (
    <div style={{ width: "100%", height: "100%" }}>
      {board.board !== null && (
        <Game
          board={board.board}
          cursor={board.cursor}
          setCursor={input.setCursor}
          mode={input.mode}
          togglePing={input.togglePing}
          toggleNotes={input.toggleNotes}
          input={input.input}
        />
      )}
    </div>
  );
};

export default Lobby;
