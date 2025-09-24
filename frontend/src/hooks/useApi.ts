import { useRef, useState } from "react";
import { type Sudoku } from "../types";

const DOMAIN: string = import.meta.env.VITE_DOMAIN;
const DEV: boolean = import.meta.env.VITE_DEV === "true";
const API_URL: string = `${DEV ? "http" : "https"}://${DOMAIN}/api`;
const WS_URL: string = `${DEV ? "ws" : "wss"}://${DOMAIN}/api`;

class ApiError extends Error {
  status: number;
  message: string;
  constructor(status: number, message: string) {
    super();
    this.status = status;
    this.message = message;
  }
}

const postLobby = async (): Promise<string> => {
  const response: Response = await fetch(API_URL + "/lobbies", {
    method: "POST",
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.text();
};

type ApiProps = {
  lobbyId: string;
  initialState: Sudoku | undefined;
  currentState: Sudoku | undefined;
  joinLobby: (id: string) => void;
  createLobby: () => void;
  sendMove: (row: number, column: number, value: number) => void;
};

type Message = {
  initial_state: Sudoku | null;
  current_state: Sudoku | null;
  error: string | null;
};

const useApi = (): ApiProps => {
  const wsRef = useRef<WebSocket | null>(null);
  const [lobbyId, setLobbyId] = useState<string | null>(null);
  const [initialState, setInitialState] = useState<Sudoku | undefined>(
    undefined,
  );
  const [currentState, setCurrentState] = useState<Sudoku | undefined>(
    undefined,
  );

  const joinLobby = (id: string): void => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    try {
      const ws: WebSocket = new WebSocket(`${WS_URL}/lobbies/${id}`);
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
        message.initial_state && setInitialState(message.initial_state);
        message.current_state && setCurrentState(message.current_state);
      };
    } catch (error) {
      console.error(error);
    }
  };

  const createLobby = (): void => {
    postLobby()
      .then((id: string) => setLobbyId(id))
      .catch((error: Error) => console.error(error));
  };

  const sendMove = (row: number, column: number, value: number): void => {
    wsRef.current &&
      wsRef.current.send(
        JSON.stringify({
          type: "move",
          row: row,
          column: column,
          value: value,
        }),
      );
  };

  return {
    lobbyId,
    initialState,
    currentState,
    joinLobby,
    createLobby,
    sendMove,
  } as ApiProps;
};

export default useApi;
export type { ApiProps };
