import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { isUUID, type History, type Lobby } from "../api/types";
import { getLobby } from "../api/api";
import { useSocket, type SocketContextProps } from "./socket";
import type { Cell } from "./board";

export type AnalysisContextProps = {
  lobbyId: string;
  setLobbyId: (state: string) => void;
  error: Error | null;
  loading: boolean;
  board: Cell[][] | null;
  players: Map<string, string>;
};

const AnalysisContext = createContext<AnalysisContextProps | null>(null);

type AnalysisProviderProps = {
  children: ReactNode;
};

const PLAYER_COLORS = [
  "#f56565", // red
  "#9f7aea", // purple
  "#ed64a6", // pink
  "#ed8936", // orange
  "#48bb78", // green
  "#63b3ed", // blue
  "#ecc94b", // yellow
  "#a78bfa", // lavender
];

export const AnalysisProvider = ({ children }: AnalysisProviderProps) => {
  const [lobbyId, setLobbyId] = useState<string>("");
  const [lobby, setLobby] = useState<Lobby | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const board: Cell[][] | null = useMemo((): Cell[][] | null => {
    if (
      lobby === null ||
      lobby.current_board === null ||
      lobby.initial_board === null
    )
      return null;

    const cells: Cell[][] = lobby.current_board.map((row, r) =>
      row.map((value, c) => ({
        value,
        initial: lobby.initial_board[r][c] !== 0,
        animation: null,
        notes: new Set<number>(),
      })),
    );

    for (let i: number = 0; i < lobby.history.length; i++) {
      const { player_name, artifacts } = lobby.history[i];
      const color = PLAYER_COLORS[i % PLAYER_COLORS.length];
      for (const artifact of artifacts) {
        const cell = cells[artifact.row][artifact.column];
        cell.color = color;
        cell.tooltip = player_name.length > 0 ? player_name : "<anonym>";
        cell.mistake =
          artifact.value !== lobby.current_board[artifact.row][artifact.column];
      }
    }

    return cells;
  }, [lobby]);

  const players: Map<string, string> = useMemo((): Map<string, string> => {
    if (lobby === null) return new Map();
    return new Map(
      lobby.history.map((history: History, index: number): [string, string] => [
        PLAYER_COLORS[index % PLAYER_COLORS.length],
        history.player_name,
      ]),
    );
  }, [lobby]);

  useEffect((): (() => void) => {
    const controller = new AbortController();

    if (!isUUID(lobbyId)) {
      setLobby(null);
      return () => controller.abort();
    }

    setError(null);
    setLoading(true);
    getLobby(lobbyId, controller.signal)
      .then((lobby: Lobby): void => setLobby(lobby))
      .catch((error: Error): void => {
        if (!controller.signal.aborted) setError(error);
      })
      .finally((): void => setLoading(false));

    return () => controller.abort();
  }, [lobbyId]);

  const value: AnalysisContextProps = {
    lobbyId,
    setLobbyId,
    error,
    loading,
    board,
    players,
  };

  return (
    <AnalysisContext.Provider value={value}>
      {children}
    </AnalysisContext.Provider>
  );
};

export const useAnalysis = (): AnalysisContextProps => {
  const context = useContext(AnalysisContext);
  if (!context) {
    throw new Error("useAnalysis must be used within a AnalysisProvider");
  }
  return context;
};
