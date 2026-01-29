import { useEffect, useState, type ReactElement } from "react";
import { getPlayer, HttpError, postPlayer } from "../../api/api";
import { useSocket, type SocketContextProps } from "../../providers/socket";
import Game from "../Game/Game";

const Lobby = (): ReactElement => {
  const socket: SocketContextProps = useSocket();
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
      <Game />
    </div>
  );
};

export default Lobby;
