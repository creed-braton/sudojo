import type { ReactElement } from "react";
import type { Player } from "../../api/types";
import styles from "./LobbyInfo.module.css";

type LobbyInfoProps = {
  players: Player[];
  maxPlayers: number;
  strict: boolean | undefined;
};

const LobbyInfo = ({
  players,
  maxPlayers,
  strict,
}: LobbyInfoProps): ReactElement => {
  return (
    <div className={`${styles.container} glass-container`}>
      <div className={styles.info}>
        <div className={styles.header}>
          <span className={styles.title}>Players</span>
          <span className={styles.count}>
            {players.length}/{maxPlayers}
          </span>
        </div>
        <span className={styles.mode}>
          {strict ? "Strict Mode" : "Lax Mode"}
        </span>
      </div>
      <div className={styles.divider} />
      <div className={styles.list}>
        {players.map((player, index) => (
          <div key={index} className={styles.player} title={player.name}>
            <div
              className={`${styles.status} ${player.active ? styles.active : styles.inactive}`}
            />
            <span className={styles.name}>
              {player.name.length > 0 ? player.name : "<anonym>"}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default LobbyInfo;
