import { useRef, useState } from "react";
import { WS_URL } from "../config";
import { ConflictEvent, type Message, type Sudoku } from "../types";

export type ClientProps = {
  connect: (id: string, token: string) => void;
  input: (row: number, column: number, value: number) => void;
  initialBoard: Sudoku | null;
  currentBoard: Sudoku | null;
  conflictEvent: ConflictEvent | undefined;
};

const useClient = (): ClientProps => {
  const wsRef = useRef<WebSocket | null>(null);
  const [initialBoard, setInitialBoard] = useState<Sudoku | null>(null);
  const [currentBoard, setCurrentBoard] = useState<Sudoku | null>(null);
  const [conflictEvent, setConflictEvent] = useState<ConflictEvent | undefined>(
    undefined,
  );

  const connect = (id: string, token: string): void => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws: WebSocket = new WebSocket(
        `${WS_URL}/lobbies/${id}?token=${token}`,
      );
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
        if (message.conflict) {
          setConflictEvent(new ConflictEvent(message.cell, message.conflict));
          return;
        }
        message.initial_state && setInitialBoard(message.initial_state);
        message.current_state && setCurrentBoard(message.current_state);
      };
    } catch (error) {
      console.error(error);
    }
  };

  const input = (row: number, column: number, value: number): void => {
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

  return {
    connect,
    input,
    initialBoard,
    currentBoard,
    conflictEvent,
  };
};

export default useClient;
