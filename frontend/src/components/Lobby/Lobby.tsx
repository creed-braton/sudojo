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
import useSudoku, { type SudokuProps } from "../../hooks/useSudoku";
import NameInput from "../NameInput/NameInput";
import useWebSocket from "../../hooks/useWebSocket";
import { type LobbyProps } from "../../types";

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
  const client: LobbyProps = useWebSocket(id, token);
  const sudoku: SudokuProps = useSudoku(client);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    setId(id);
    const token: string | undefined = getToken(id);
    !token && setNameInput(true);
    setToken(token || "");
  }, [location.pathname]);

  useEffect((): void => {
    if (!sudoku.current) return;
    for (let i: number = 0; i < sudoku.current.length; i++) {
      for (let j: number = 0; j < sudoku.current[i].length; j++) {
        if (sudoku.current[i][j] === 0) return;
      }
    }
    navigate(`/s/${id}`);
  }, [sudoku.current]);

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
      {sudoku.initial && sudoku.current && (
        <>
          <Board
            selected={sudoku.selected}
            select={sudoku.select}
            initialBoard={sudoku.initial}
            currentBoard={sudoku.current}
            notes={sudoku.notes}
            conflictEvent={sudoku.conflictEvent}
            pingEvent={sudoku.pingEvent}
          />
          <InputBar
            input={sudoku.input}
            pencilMode={sudoku.pencilMode}
            pingMode={sudoku.pingMode}
            togglePencil={sudoku.togglePencil}
            togglePing={sudoku.togglePing}
          />
        </>
      )}
    </div>
  );
};

export default Lobby;
