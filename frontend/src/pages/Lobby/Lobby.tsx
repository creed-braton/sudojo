import { useEffect, useState, type ReactElement } from "react";
import { getPlayer, HttpError, postPlayer } from "../../api/api";
import { useSocket, type SocketContextProps } from "../../providers/socket";
import Game from "../Game/Game";
import Username from "../../components/Username/Username";
import style from "./Lobby.module.css";

const Lobby = (): ReactElement => {
  const socket: SocketContextProps = useSocket();
  const [id, setId] = useState<string>("");
  const [username, setUsername] = useState<string>("");
  const [_, setError] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect((): void => {
    setLoading(true);

    const id: string = location.pathname.split("/")[2];
    setId(id);

    getPlayer(id)
      .then((): void => socket.setId(id))
      .catch((error: Error) => {
        if (error instanceof HttpError && error.status == 423) {
        } else if (error instanceof HttpError && error.status == 401) {
          // nothing happens user must put in name
        } else {
          setError(error.message);
        }
      })
      .finally(() => setLoading(false));
  }, [location.pathname]);

  const joinGame = (): void => {
    postPlayer(id, username)
      .then(() => socket.setId(id))
      .catch((error: Error) => setError(error.message));
  };

  return (
    <div className={style.lobby}>
      {!loading &&
        (socket.id.length < 1 ? (
          <div className={style.username}>
            <Username
              username={username}
              setUsername={setUsername}
              onClick={joinGame}
            />
          </div>
        ) : (
          <div className={style.game}>
            <Game />
          </div>
        ))}
    </div>
  );
};

export default Lobby;
