import { useEffect, useState, type ReactElement } from "react";
import { useLocation, type Location } from "react-router-dom";
import type { Cell } from "../../types";
import Board from "../Board/Board";
import styles from "./Lobby.module.css";
import InputBar from "../InputBar/InputBar";
import useSudoku, { type SudokuProps } from "../../hooks/useSudoku";
import NameInput from "../NameInput/NameInput";

const Lobby = ({
  joinLobby,
  getToken,
}: {
  joinLobby: (id: string, name: string) => Promise<string>;
  getToken: (id: string) => string | undefined;
}): ReactElement => {
  const [position, setPosition] = useState<Cell | undefined>(undefined);
  const [nameInput, setNameInput] = useState<boolean>(false);
  const [id, setId] = useState<string>("");
  const [token, setToken] = useState<string | undefined>(undefined);
  const location: Location = useLocation();
  const sudoku: SudokuProps = useSudoku();

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    setId(id);
    const token: string | undefined = getToken(id);
    !token && setNameInput(true);
    setToken(token);
  }, [location.pathname]);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    token && sudoku.connect(id, token);
  }, [token]);

  useEffect((): void => {});

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
      {sudoku.initialBoard && sudoku.currentBoard && (
        <>
          <Board
            position={position}
            setPosition={setPosition}
            initialBoard={sudoku.initialBoard}
            currentBoard={sudoku.currentBoard}
            notes={sudoku.notes}
            conflictEvent={sudoku.conflictEvent}
          />
          <InputBar
            position={position}
            input={sudoku.input}
            pencilMode={sudoku.pencilMode}
            toggleMode={sudoku.toggleMode}
          />
        </>
      )}
    </div>
  );
};

export default Lobby;
