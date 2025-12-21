import { useState, type ReactElement } from "react";
import { createPortal } from "react-dom";
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
  const [tooltip, setTooltip] = useState<{
    text: string;
    x: number;
    y: number;
  } | null>(null);

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
        {players.map((player, index) => {
          const displayName = player.name.length > 0 ? player.name : "<anonym>";
          return (
            <div
              key={index}
              className={styles.player}
              onMouseEnter={(e) => {
                setTooltip({
                  text: displayName,
                  x: e.clientX + 10,
                  y: e.clientY - 10,
                });
              }}
              onMouseMove={(e) => {
                if (tooltip) {
                  setTooltip({
                    text: displayName,
                    x: e.clientX + 10,
                    y: e.clientY - 10,
                  });
                }
              }}
              onMouseLeave={() => setTooltip(null)}
            >
              <div
                className={`${styles.status} ${player.active ? styles.active : styles.inactive}`}
              />
              <span className={styles.name}>{displayName}</span>
            </div>
          );
        })}
      </div>
      {tooltip &&
        createPortal(
          <div
            className={styles.tooltip}
            style={{
              left: tooltip.x,
              top: tooltip.y,
            }}
          >
            {tooltip.text}
          </div>,
          document.body
        )}
    </div>
  );
};

export default LobbyInfo;
