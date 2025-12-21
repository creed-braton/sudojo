import { useCallback, useEffect, useRef, useState } from "react";
import {
  isInsertMessage,
  isJoinMessage,
  isLeaveMessage,
  isPingMessage,
  isStateMessage,
  isSystemMessage,
  type Player,
  type Sudoku,
} from "./types";
import { WS_URL } from "./config";

export type WebSocketProps = {
  current: Sudoku | undefined;
  initial: Sudoku | undefined;
  players: Player[];
  maxPlayer: number;
  strict: boolean | undefined;
  insert: (row: number, column: number, value: number) => void;
  ping: (row: number, column: number) => void;
  animations: Map<string, Animation>;
  idle: boolean;
};

export type AnimationType = "conflict" | "ping";

export type Animation = {
  type: AnimationType;
  key: number;
};

const useWebSocket = (id: string, token: string): WebSocketProps => {
  const wsRef = useRef<WebSocket | null>(null);
  const [current, setCurrent] = useState<Sudoku | undefined>(undefined);
  const [initial, setInitial] = useState<Sudoku | undefined>(undefined);
  const [players, setPlayers] = useState<Player[]>([]);
  const [maxPlayer, setMaxPlayer] = useState<number>(0);
  const [strict, setStrict] = useState<boolean | undefined>(undefined);
  const [animations, setAnimations] = useState<Map<string, Animation>>(
    new Map(),
  );
  const [idle, setIdle] = useState<boolean>(false);

  const triggerAnimation = useCallback(
    (row: number, column: number, type: AnimationType) => {
      const key: string = `${row}-${column}`;

      setAnimations((prev: Map<string, Animation>) => {
        const next = new Map(prev);
        next.set(key, { type, key: Date.now() });
        return next;
      });

      setTimeout(() => {
        setAnimations((prev: Map<string, Animation>) => {
          const next: Map<string, Animation> = new Map(prev);
          next.delete(key);
          return next;
        });
      }, 600);
    },
    [],
  );

  const connect = (id: string, token: string): void => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws: WebSocket = new WebSocket(
        `${WS_URL}/lobbies/${id}/ws?token=${token}`,
      );
      wsRef.current = ws;

      ws.onopen = (): void => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: "state" }));
        }
      };

      ws.onclose = (event: CloseEvent): void => {
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
        if (event.code !== 1000 && event.code !== 1001) {
          console.error("receive closure code: ", event.code);
          connect(id, token);
        } else {
          setIdle(true);
        }
      };

      ws.onerror = (_: Event): void => {
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
          setMaxPlayer(message.max_player);
          setStrict(message.strict);
        } else if (isInsertMessage(message)) {
          message.current_board && setCurrent(message.current_board);
          if (
            message.conflict !== undefined &&
            message.row !== undefined &&
            message.column !== undefined
          ) {
            triggerAnimation(message.row, message.column, "conflict");
          }
          message.error && console.error(message.error);
        } else if (isPingMessage(message)) {
          if (message.row !== undefined && message.column !== undefined) {
            triggerAnimation(message.row, message.column, "ping");
          } else if (message.error !== undefined) {
            console.error(message.error);
          }
        } else if (isSystemMessage(message)) {
          console.error(message.error);
        }
      };
    } catch (error) {
      console.error(error);
    }
  };

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

  useEffect((): void => {
    if (id.length > 0 && token.length > 0) {
      connect(id, token);
    }
  }, [id, token]);

  return {
    current,
    initial,
    players,
    maxPlayer,
    strict,
    insert,
    ping,
    animations,
    idle,
  };
};

export default useWebSocket;
