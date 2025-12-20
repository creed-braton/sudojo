import type { ReactElement } from "react";
import type { Player } from "../../api/types";
import styles from "./Players.module.css";

type PlayersProps = {
  players: Player[];
  maxPlayers: number;
};

const Players = ({ players, maxPlayers }: PlayersProps): ReactElement => {
  return (
    <div className={`${styles.players} glass-container`}>
      <div className={styles.header}>
        <span className={styles.title}>Players</span>
        <span className={styles.count}>
          {players.length}/{maxPlayers}
        </span>
      </div>
      <div className={styles.list}>
        {players.map((player, index) => (
          <div key={index} className={styles.player}>
            <div
              className={`${styles.status} ${player.active ? styles.active : styles.inactive}`}
            />
            <span className={styles.name}>{player.name}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Players;
