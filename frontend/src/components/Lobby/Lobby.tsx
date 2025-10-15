import { useEffect, useState, type ReactElement } from "react";
import { useLocation } from "react-router-dom";
import type { Cell, Player } from "../../types";
import Board from "../Board/Board";
import styles from "./Lobby.module.css";
import InputBar from "../InputBar/InputBar";
import useSudoku, { type SudokuProps } from "../../hooks/useSudoku";

const Lobby = ({
  joinLobby,
  getPlayer,
}: {
  joinLobby: (id: string, name: string) => Promise<string>;
  getPlayer: (id: string) => Player | undefined;
}): ReactElement => {
  const [position, setPosition] = useState<Cell | undefined>(undefined);
  const location = useLocation();
  const sudoku: SudokuProps = useSudoku();

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    const player: Player | undefined = getPlayer(id);
    player
      ? sudoku.connect(id, player.token)
      : joinLobby(id, "").then((token: string) => sudoku.connect(id, token));
  }, [location.pathname]);

  return (
    <div className={styles.lobby}>
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
