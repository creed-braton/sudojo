import type { ReactElement } from "react";
import Player from "../Player/Player";
import style from "./PlayerList.module.css";

type PlayerListProps = {
  players: Map<string, string>;
  maxPlayers: number;
};

const PlayerList = ({ players, maxPlayers }: PlayerListProps): ReactElement => {
  return (
    <div className={style.list}>
      <span className={style.headline}>
        Players {players.size}/{maxPlayers}
      </span>
      <div className={style.entries}>
        {Array.from(players.entries()).map(([color, name]) => (
          <div className={style.entry} key={color} title={name}>
            <Player color={color} name={name} />
          </div>
        ))}
      </div>
    </div>
  );
};

export default PlayerList;
