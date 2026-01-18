import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
  type RefObject,
} from "react";
import {
  isInsertMessage,
  isJoinMessage,
  isLeaveMessage,
  isPingMessage,
  isStateMessage,
  type Sudoku,
  type Config,
  type Player,
  isUUID,
} from "../api/types";

export type SocketContextProps = {
  id: string;
  setId: (state: string) => void;
  initial: Sudoku | null;
  current: Sudoku | null;
  players: Player[];
  config: Config | null;
  insert: (row: number, column: number, value: number) => void;
  ping: (row: number, column: number) => void;
  setOnInsert: (
    callback: (row: number, column: number, value: number) => void,
  ) => void;
  setOnConflict: (callback: (row: number, column: number) => void) => void;
  setOnPing: (callback: (row: number, column: number) => void) => void;
};

const SocketContext = createContext<SocketContextProps | null>(null);

type SocketProviderProps = {
  url: string;
  children: ReactNode;
};

export const SocketProvider = ({ url, children }: SocketProviderProps) => {
  const [id, setId] = useState<string>("");
  const wsRef: RefObject<WebSocket | null> = useRef<WebSocket | null>(null);

  const [initial, setInitial] = useState<Sudoku | null>(null);
  const [current, setCurrent] = useState<Sudoku | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);
  const [config, setConfig] = useState<Config | null>(null);

  const onInsertRef: RefObject<
    (row: number, column: number, value: number) => void
  > = useRef<(row: number, column: number, value: number) => void>(() => {});
  const onConflictRef: RefObject<(row: number, column: number) => void> =
    useRef<(row: number, column: number) => void>(() => {});
  const onPingRef: RefObject<(row: number, column: number) => void> = useRef<
    (row: number, column: number) => void
  >(() => {});

  const setOnInsert = (
    callback: (row: number, column: number, value: number) => void,
  ): void => {
    onInsertRef.current = callback;
  };
  const setOnConflict = (
    callback: (row: number, column: number) => void,
  ): void => {
    onConflictRef.current = callback;
  };
  const setOnPing = (callback: (row: number, column: number) => void): void => {
    onPingRef.current = callback;
  };

  useEffect((): void => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    if (!isUUID(id)) return;

    try {
      const ws: WebSocket = new WebSocket(`${url}/lobbies/${id}/ws`);
      wsRef.current = ws;

      ws.onopen = (): void => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: "state" }));
        }
      };

      ws.onclose = (_: CloseEvent): void => {
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
      };

      ws.onerror = (event: Event): void => {
        console.error(event);
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
      };

      ws.onmessage = (event: MessageEvent): void => {
        const message: unknown = JSON.parse(event.data);
        if (isJoinMessage(message)) {
          setPlayers(message.players);
        } else if (isLeaveMessage(message)) {
          setPlayers(message.players);
        } else if (isStateMessage(message)) {
          setCurrent(message.current_board);
          setInitial(message.initial_board);
          setPlayers(message.players);
          setConfig(message.config);
        } else if (isInsertMessage(message)) {
          message.current_board && setCurrent(message.current_board);
          if (
            message.row !== undefined &&
            message.column !== undefined &&
            message.current_board !== undefined
          ) {
            onInsertRef.current(
              message.row,
              message.column,
              message.current_board[message.row][message.column],
            );
          }
          if (
            message.row !== undefined &&
            message.column !== undefined &&
            message.conflict !== undefined
          ) {
            onConflictRef.current(message.row, message.column);
          }
          message.error && console.error(message.error);
        } else if (isPingMessage(message)) {
          if (message.row !== undefined && message.column !== undefined) {
            onPingRef.current(message.row, message.column);
          } else if (message.error !== undefined) {
            console.error(message.error);
          }
        }
      };
    } catch (error) {
      console.error(error);
    }
  }, [id]);

  const insert = (row: number, column: number, value: number): void => {
    wsRef.current &&
      wsRef.current.send(
        JSON.stringify({
          type: "insert",
          row: row,
          column: column,
          value: value,
        }),
      );
  };

  const ping = (row: number, column: number): void => {
    wsRef.current &&
      wsRef.current.send(
        JSON.stringify({
          type: "ping",
          row: row,
          column: column,
        }),
      );
  };

  const value: SocketContextProps = {
    id,
    setId,
    initial,
    current,
    players,
    config,
    insert,
    ping,
    setOnInsert,
    setOnConflict,
    setOnPing,
  };

  return (
    <SocketContext.Provider value={value}>{children}</SocketContext.Provider>
  );
};

export const useSocket = (): SocketContextProps => {
  const context = useContext(SocketContext);
  if (!context) {
    throw new Error("useSocket must be used within a SocketProvider");
  }
  return context;
};
