import { useEffect, useState, type ReactElement } from "react";
import { useLocation } from "react-router-dom";
import type { Sudoku } from "../../types";
import Board from "../Board/Board";
import styles from "./Lobby.module.css";
import InputBar from "../InputBar/InputBar";

type Position = {
  row: number;
  column: number;
};

const Lobby = ({
  joinLobby,
  sendMove,
  initialState,
  currentState,
}: {
  joinLobby: (id: string) => void;
  sendMove: (row: number, column: number, value: number) => void;
  initialState: Sudoku | undefined;
  currentState: Sudoku | undefined;
}): ReactElement => {
  const [position, setPosition] = useState<Position | undefined>(undefined);
  const location = useLocation();

  useEffect((): void => {
    const id = location.pathname.split("/")[2];
    joinLobby(id);
  }, [location.pathname]);

  return (
    <div className={styles.lobby}>
      {initialState && currentState && (
        <>
          <Board
            position={position}
            setPosition={setPosition}
            initialState={initialState}
            currentState={currentState}
          />
          <InputBar position={position} sendMove={sendMove} />
        </>
      )}
    </div>
  );
};

export default Lobby;
export type { Position };
