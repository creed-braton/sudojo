import { useEffect, useState, type ReactElement } from "react";
import {
  useLocation,
  useNavigate,
  type Location,
  type NavigateFunction,
} from "react-router-dom";
import Board from "../Board/Board";
import styles from "./Lobby.module.css";
import InputBar from "../InputBar/InputBar";
import NameInput from "../NameInput/NameInput";
import LobbyInfo from "../LobbyInfo/LobbyInfo";
import type { SudokuProps } from "../../hooks/sudoku";
import useWebSocket, { type WebSocketProps } from "../../api/socket";
import useSudoku from "../../hooks/sudoku";

const Lobby = ({
  joinLobby,
  getToken,
}: {
  joinLobby: (id: string, name: string) => Promise<string>;
  getToken: (id: string) => string | undefined;
}): ReactElement => {
  const [id, setId] = useState<string>("");
  const [nameInput, setNameInput] = useState<boolean>(false);
  const [token, setToken] = useState<string>("");
  const location: Location = useLocation();
  const navigate: NavigateFunction = useNavigate();
  const socket: WebSocketProps = useWebSocket(id, token);
  const sudoku: SudokuProps = useSudoku(
    socket.initial,
    socket.current,
    socket.insert,
    socket.ping,
  );

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    setId(id);
    const token: string | undefined = getToken(id);
    !token && setNameInput(true);
    setToken(token || "");
  }, [location.pathname]);

  useEffect((): void => {
    sudoku.finished && navigate(`/s/${id}`);
  }, [sudoku.finished]);

  return (
    <div className={styles.lobby}>
      {nameInput && (
        <NameInput
          lobbyId={id}
          joinLobby={joinLobby}
          setToken={setToken}
          close={() => setNameInput(false)}
        />
      )}
      {socket.initial && socket.current && (
        <>
          <LobbyInfo
            players={socket.players}
            maxPlayers={socket.maxPlayer}
            strict={socket.strict}
          />
          <Board
            cursor={sudoku.cursor}
            select={sudoku.select}
            initial={socket.initial}
            current={socket.current}
            notes={sudoku.notes}
            animations={socket.animations}
          />
          <InputBar
            input={sudoku.input}
            mode={sudoku.mode}
            togglePencil={sudoku.togglePencil}
            togglePing={sudoku.togglePing}
          />
        </>
      )}
    </div>
  );
};

export default Lobby;
