import { useEffect, useMemo, useState, type ReactElement } from "react";
import { createPortal } from "react-dom";
import { getLobby } from "../../api/api";
import type { Lobby } from "../../api/types";
import Board, { type Insertion } from "../Board/Board";
import Plot from "../Plot/Plot";
import CrownIcon from "../../icons/CrownIcon";
import styles from "./Statistic.module.css";

const emptyNotes = new Map<string, Set<number>>();
const emptyAnimations = new Map();
const noop = () => {};

const formatTimestamp = (timestampNanos: number): string => {
  const date = new Date(timestampNanos / 1_000_000);
  const day = date.getDate().toString().padStart(2, "0");
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const year = date.getFullYear();
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const seconds = date.getSeconds().toString().padStart(2, "0");
  return `${day}/${month}/${year} ${hours}:${minutes}:${seconds}`;
};

const formatDuration = (startNanos: number, endNanos: number): string => {
  const durationMs = (endNanos - startNanos) / 1_000_000;
  const totalSeconds = Math.floor(durationMs / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  const parts: string[] = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);
  parts.push(`${seconds}s`);

  return parts.join(" ");
};

export const PLAYER_COLORS = [
  "#F56565", // Red
  "#9F7AEA", // Purple
  "#ED64A6", // Pink
  "#ED8936", // Orange
  "#48BB78", // Green
  "#63B3ED", // Blue
  "#ECC94B", // Yellow
  "#A78BFA", // Lavender
];

const Statistic = (): ReactElement => {
  const [id, setId] = useState<string>("");
  const [lobby, setLobby] = useState<Lobby | undefined>(undefined);
  const [tooltip, setTooltip] = useState<{
    text: string;
    x: number;
    y: number;
  } | null>(null);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    setId(id);
  }, [location.pathname]);

  useEffect((): void => {
    id.length > 0 && getLobby(id).then((lobby: Lobby) => setLobby(lobby));
  }, [id]);

  const insertions = useMemo((): Insertion[] => {
    if (!lobby) {
      return [];
    }

    const players = lobby.history.map((h, index) => ({
      name:
        h.player_name && h.player_name.length > 0 ? h.player_name : "<anonym>",
      color: PLAYER_COLORS[index % PLAYER_COLORS.length],
      artifacts: h.artifacts ?? [],
    }));

    const allArtifacts = players.flatMap((player) =>
      player.artifacts.map((artifact) => ({
        ...artifact,
        playerName: player.name,
        playerColor: player.color,
      })),
    );

    allArtifacts.sort((a, b) => a.timestamp - b.timestamp);

    const correctInsertions: Insertion[] = [];
    const claimed = new Set<string>();

    for (const artifact of allArtifacts) {
      const cellKey = `${artifact.row}-${artifact.column}`;

      // Skip if already claimed by an earlier insertion
      if (claimed.has(cellKey)) continue;

      // Skip if this cell was pre-filled in the initial board
      if (lobby.initial_board[artifact.row][artifact.column] !== 0) continue;

      // Skip if this cell is empty in the current board
      const currentValue = lobby.current_board[artifact.row][artifact.column];
      if (currentValue === 0) continue;

      // In strict mode, only count insertion if the value matches the current board
      if (lobby.strict && artifact.value !== currentValue) continue;

      // First insertion for this cell - claim it for this player
      claimed.add(cellKey);
      correctInsertions.push({
        row: artifact.row,
        column: artifact.column,
        value: currentValue,
        playerName: artifact.playerName,
        playerColor: artifact.playerColor,
        timestamp: artifact.timestamp,
      });
    }

    return correctInsertions;
  }, [lobby]);

  const playerNames = lobby?.history.map((h) => h.player_name ?? "") ?? [];

  const playersData = useMemo(() => {
    return playerNames.map((name, index) => ({
      name: name.length > 0 ? name : "<anonym>",
      color: PLAYER_COLORS[index % PLAYER_COLORS.length],
    }));
  }, [playerNames]);

  const leaderIndex = useMemo((): number | null => {
    if (playerNames.length < 2) return null;

    const insertionCounts = new Map<number, number>();
    for (let i = 0; i < playerNames.length; i++) {
      insertionCounts.set(i, 0);
    }

    for (const insertion of insertions) {
      const playerIndex = playersData.findIndex(
        (p) => p.name === insertion.playerName && p.color === insertion.playerColor
      );
      if (playerIndex !== -1) {
        insertionCounts.set(playerIndex, (insertionCounts.get(playerIndex) ?? 0) + 1);
      }
    }

    let maxCount = 0;
    let leader: number | null = null;
    let isTie = false;

    for (const [index, count] of insertionCounts) {
      if (count > maxCount) {
        maxCount = count;
        leader = index;
        isTie = false;
      } else if (count === maxCount && count > 0) {
        isTie = true;
      }
    }

    return isTie || maxCount === 0 ? null : leader;
  }, [playerNames, insertions, playersData]);

  return (
    <div className={styles.wrapper}>
      {lobby && (
        <div className={styles.content}>
          <div className={styles.topSection}>
            <Board
              cursor={null}
              select={noop}
              initial={lobby.initial_board}
              current={lobby.current_board}
              notes={emptyNotes}
              animations={emptyAnimations}
              insertions={insertions}
            />
            <div className={styles.sidebar}>
              <div className={`${styles.container} glass-container`}>
                <span className={styles.mode}>
                  {lobby.strict ? "Strict Mode" : "Lax Mode"}
                </span>
                <div className={styles.divider} />
                <span className={styles.timestamp}>
                  <span className={styles.timestampLabel}>Start</span>
                  <span className={styles.timestampValue}>
                    {formatTimestamp(lobby.started_at)}
                  </span>
                </span>
                <span className={styles.timestamp}>
                  <span className={styles.timestampLabel}>Finish</span>
                  <span className={styles.timestampValue}>
                    {lobby.finished_at !== null
                      ? formatTimestamp(lobby.finished_at)
                      : "<null>"}
                  </span>
                </span>
                <span className={styles.timestamp}>
                  <span className={styles.timestampLabel}>Duration</span>
                  <span className={styles.timestampValue}>
                    {lobby.finished_at !== null
                      ? formatDuration(lobby.started_at, lobby.finished_at)
                      : "<null>"}
                  </span>
                </span>
              </div>
              <div className={`${styles.container} glass-container`}>
                <div className={styles.header}>
                  <span className={styles.title}>Players</span>
                  <span className={styles.count}>
                    {playerNames.length}/{lobby.max_player}
                  </span>
                </div>
                <div className={styles.list}>
                  {playerNames.map((name, index) => {
                    const displayName = name.length > 0 ? name : "<anonym>";
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
                          className={styles.colorIndicator}
                          style={{
                            backgroundColor:
                              PLAYER_COLORS[index % PLAYER_COLORS.length],
                            boxShadow: `0 0 6px ${PLAYER_COLORS[index % PLAYER_COLORS.length]}80`,
                          }}
                        />
                        <span className={styles.name}>{displayName}</span>
                        {leaderIndex === index && (
                          <CrownIcon className={styles.crown} />
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          </div>
          <div className={`${styles.plotContainer} glass-container`}>
            <Plot
              insertions={insertions}
              players={playersData}
              startTimestamp={lobby.started_at}
            />
          </div>
        </div>
      )}
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

export default Statistic;
