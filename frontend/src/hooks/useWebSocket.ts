import { useEffect, useRef, useState } from "react";
import {
  type Conflict,
  type LobbyProps,
  type Player,
  type Position,
  type Sudoku,
} from "../types";
import { WS_URL } from "../config";

type Message = {
  type: string;
  initial_state: Sudoku | null;
  current_state: Sudoku | null;
  error: string | null;
  conflict: string | null;
  row: number | null;
  column: number | null;
  value: number | null;
  strict: boolean | null;
  max_player: number | null;
  players: Player[] | null;
};

const useWebSocket = (lobbyId: string, token: string): LobbyProps => {
  const wsRef = useRef<WebSocket | null>(null);
  const [current, setCurrent] = useState<Sudoku | null>(null);
  const [initial, setInitial] = useState<Sudoku | null>(null);
  const [conflictEvent, setConflictEvent] = useState<Conflict | null>(null);
  const [pingEvent, setPingEvent] = useState<Position | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);
  const [maxPlayer, setMaxPlayer] = useState<number>(0);
  const [strict, setStrict] = useState<boolean | null>(null);

  const connect = (lobbyId: string, token: string): void => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws: WebSocket = new WebSocket(
        `${WS_URL}/lobbies/${lobbyId}?token=${token}`,
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
        }
      };

      ws.onerror = (_: Event): void => {
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
      };

      ws.onmessage = (event: MessageEvent): void => {
        const message = JSON.parse(event.data) as Message;
        if (message.error) {
          console.error(message.error);
          return;
        }
        if (message.type === "ping") {
          if (message.row !== null && message.column !== null) {
            setPingEvent({
              row: message.row,
              column: message.column,
              timestamp: Date.now(),
            });
          } else {
            console.error("missing row and column in ping event");
          }
          return;
        }
        if (
          message.conflict &&
          message.row !== null &&
          message.column != null
        ) {
          setConflictEvent({
            message: message.conflict,
            row: message.row,
            column: message.column,
            timestamp: Date.now(),
          });
          return;
        }
        message.initial_state && setInitial(message.initial_state);
        message.current_state && setCurrent(message.current_state);
        message.players && setPlayers(message.players);
        message.max_player && setMaxPlayer(message.max_player);
        message.strict !== null && setStrict(message.strict);
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
    console.log("test");
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
    if (lobbyId.length > 0 && token.length > 0) {
      connect(lobbyId, token);
    }
  }, [lobbyId, token]);

  return {
    current,
    initial,
    insert,
    ping,
    conflictEvent,
    pingEvent,
    players,
    maxPlayer,
    strict,
  };
};

export default useWebSocket;
