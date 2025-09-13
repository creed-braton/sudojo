import { useCallback, useEffect, useRef, useState } from "react";

// =============================================================================
// Types
// =============================================================================

// HTTP API Response Types
type CreateLobbyResponse = {
  id: string;
};

// WebSocket Message Types (Outgoing - Client to Server)
type MakeMoveMessage = {
  type: "move";
  row: number;
  col: number;
  value: number;
};

type ClearCellMessage = {
  type: "clear";
  row: number;
  col: number;
};

type RequestGameStateMessage = {
  type: "request_state";
};

type OutgoingMessage =
  | MakeMoveMessage
  | ClearCellMessage
  | RequestGameStateMessage;

// WebSocket Message Types (Incoming - Server to Client)
type MoveSuccessMessage = {
  type: "success";
  row: number;
  col: number;
  value: number;
  success: true;
};

type MoveErrorMessage = {
  type: "error";
  success: false;
  error: string;
};

type GameStateMessage = {
  type: "state";
  board: number[][];
  initialBoard: number[][];
};

type BoardUpdateMessage = {
  type: "board_update";
  board: number[][];
};

type MoveNotificationMessage = {
  type: "move_notification";
  row: number;
  col: number;
  value: number;
  board: number[][];
};

type IncomingMessage =
  | MoveSuccessMessage
  | MoveErrorMessage
  | GameStateMessage
  | BoardUpdateMessage
  | MoveNotificationMessage;

// Connection State Type
type ConnectionState = "disconnected" | "connecting" | "connected" | "error";

// Board Type
type Board = number[][];

// Hook Parameters
type UseApiParams = {
  /**
   * The base URL for the API and WebSocket server (e.g., "localhost:8080")
   * @default "localhost:8080"
   */
  baseUrl?: string;

  /**
   * Whether to use secure connections (HTTPS/WSS) instead of HTTP/WS
   * @default false
   */
  secure?: boolean;
};

// Hook Return Type
type UseApiReturn = {
  // Connection state
  connectionState: ConnectionState;
  connectionError: string | null;

  // HTTP API methods
  createLobby: () => Promise<string>;

  // WebSocket connection method
  joinLobby: (lobbyId: string) => void;

  // Game state
  board: Board | null;
  initialBoard: Board | null;

  // WebSocket methods
  makeMove: (row: number, col: number, value: number) => void;
  clearCell: (row: number, col: number) => void;
  requestGameState: () => void;

  // Last message information
  lastError: string | null;
  lastMoveSuccess: boolean | null;
};

// =============================================================================
// Hook Implementation
// =============================================================================

/**
 * Custom hook for handling all API communication with the backend
 *
 * This hook encapsulates:
 * 1. HTTP API calls (createLobby)
 * 2. WebSocket connection management
 * 3. WebSocket message sending and handling
 * 4. Game state management
 *
 * @param {UseApiParams} params - Configuration parameters for the API
 * @param {string} [params.baseUrl="localhost:8080"] - Base URL for the API and WebSocket server
 * @param {boolean} [params.secure=false] - Whether to use secure connections (HTTPS/WSS)
 * @returns {UseApiReturn} API methods and state
 */
export const useApi = ({
  baseUrl = "localhost:8080",
  secure = false,
}: UseApiParams = {}): UseApiReturn => {
  // WebSocket reference
  const webSocketRef = useRef<WebSocket | null>(null);

  // State for connection status
  const [connectionState, setConnectionState] =
    useState<ConnectionState>("disconnected");
  const [connectionError, setConnectionError] = useState<string | null>(null);

  // Game state
  const [board, setBoard] = useState<Board | null>(null);
  const [initialBoard, setInitialBoard] = useState<Board | null>(null);

  // Last message status
  const [lastError, setLastError] = useState<string | null>(null);
  const [lastMoveSuccess, setLastMoveSuccess] = useState<boolean | null>(null);

  // Construct URLs based on parameters
  const httpProtocol = secure ? "https" : "http";
  const wsProtocol = secure ? "wss" : "ws";
  const API_BASE_URL = `${httpProtocol}://${baseUrl}`;
  const WS_BASE_URL = `${wsProtocol}://${baseUrl}`;

  // Debug flag to track if we're the lobby creator
  const [_, setIsLobbyCreator] = useState<boolean>(false);

  // =============================================================================
  // HTTP API Methods
  // =============================================================================

  /**
   * Creates a new game lobby via HTTP POST request
   *
   * @returns {Promise<string>} The created lobby ID
   */
  const createLobby = useCallback(async (): Promise<string> => {
    try {
      const response = await fetch(`${API_BASE_URL}/lobby`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error(
          `Failed to create lobby: ${response.status} ${response.statusText}`
        );
      }

      const data = (await response.json()) as CreateLobbyResponse;
      // Mark this client as the lobby creator
      setIsLobbyCreator(true);
      return data.id;
    } catch (error) {
      setConnectionError((error as Error).message);
      throw error;
    }
  }, [API_BASE_URL]);

  // =============================================================================
  // WebSocket Connection Management
  // =============================================================================

  /**
   * Joins an existing lobby by establishing a WebSocket connection
   *
   * @param {string} lobbyId - The ID of the lobby to join
   */
  const joinLobby = useCallback(
    (lobbyId: string): void => {
      // Close existing connection if any
      if (webSocketRef.current) {
        webSocketRef.current.close();
      }

      try {
        setConnectionState("connecting");
        console.log(
          `[useApi] Connecting to WebSocket at ${WS_BASE_URL}/lobby?id=${lobbyId}`
        );

        // Don't reset isLobbyCreator here - it should remain true if this client created the lobby
        // Only set to false if we know this is a different lobby than the one we created

        // Create new WebSocket connection
        const ws = new WebSocket(`${WS_BASE_URL}/lobby?id=${lobbyId}`);
        webSocketRef.current = ws;

        // Setup WebSocket event handlers
        ws.onopen = () => {
          console.log(`[useApi] WebSocket connection opened`);
          setConnectionState("connected");
          setConnectionError(null);

          // Request initial game state when connected
          // Using setTimeout to ensure this runs after the state is updated
          setTimeout(() => {
            console.log(
              `[useApi] Requesting initial game state after connection`
            );
            if (ws.readyState === WebSocket.OPEN) {
              ws.send(JSON.stringify({ type: "request_state" }));
            }
          }, 100);
        };

        ws.onclose = (event) => {
          console.log(
            `[useApi] WebSocket connection closed: ${event.code} ${event.reason}`
          );
          setConnectionState("disconnected");

          // Clear the WebSocket reference if it's the current one
          if (webSocketRef.current === ws) {
            webSocketRef.current = null;
          }
        };

        ws.onerror = (event) => {
          console.error("[useApi] WebSocket error:", event);
          setConnectionState("error");
          setConnectionError("WebSocket connection error");

          // Clear the WebSocket reference if it's the current one
          if (webSocketRef.current === ws) {
            webSocketRef.current = null;
          }
        };

        // Handle incoming messages
        ws.onmessage = (event) => {
          console.log(`[useApi] Received message:`, event.data);
          try {
            const message = JSON.parse(event.data) as IncomingMessage;
            handleIncomingMessage(message);
          } catch (error) {
            console.error("[useApi] Error parsing WebSocket message:", error);
            setLastError("Invalid message format received");
          }
        };
      } catch (error) {
        console.error(
          "[useApi] Error establishing WebSocket connection:",
          error
        );
        setConnectionState("error");
        setConnectionError((error as Error).message);
      }
    },
    [WS_BASE_URL]
  );

  // Cleanup WebSocket on unmount
  useEffect(() => {
    return () => {
      if (webSocketRef.current) {
        webSocketRef.current.close();
      }
    };
  }, []);

  // =============================================================================
  // WebSocket Message Sending
  // =============================================================================

  /**
   * Sends a message through the WebSocket connection
   *
   * @param {OutgoingMessage} message - The message to send
   */
  const sendMessage = useCallback(
    (message: OutgoingMessage): void => {
      if (webSocketRef.current && connectionState === "connected") {
        console.log(`[useApi] Sending message: ${JSON.stringify(message)}`);
        webSocketRef.current.send(JSON.stringify(message));
      } else {
        console.warn(
          `[useApi] Cannot send message: WebSocket not connected (state: ${connectionState})`
        );
        setLastError("Cannot send message: WebSocket not connected");
      }
    },
    [connectionState]
  );

  /**
   * Requests the current game state from the server
   */
  const requestGameState = useCallback((): void => {
    console.log("[useApi] Explicitly requesting game state");
    sendMessage({
      type: "request_state",
    });
  }, [sendMessage]);

  /**
   * Makes a move in the game
   *
   * @param {number} row - Row index (0-based)
   * @param {number} col - Column index (0-based)
   * @param {number} value - The value to place (1-9)
   */
  const makeMove = useCallback(
    (row: number, col: number, value: number): void => {
      // Send move to server
      sendMessage({
        type: "move",
        row,
        col,
        value,
      });
    },
    [sendMessage]
  );

  /**
   * Clears a cell in the game
   *
   * @param {number} row - Row index (0-based)
   * @param {number} col - Column index (0-based)
   */
  const clearCell = useCallback(
    (row: number, col: number): void => {
      // Send clear message to server
      sendMessage({
        type: "clear",
        row,
        col,
      });
    },
    [sendMessage]
  );

  // =============================================================================
  // WebSocket Message Handling
  // =============================================================================

  /**
   * Handles incoming WebSocket messages based on their type
   *
   * @param {IncomingMessage} message - The parsed message from the server
   */
  const handleIncomingMessage = useCallback(
    (message: IncomingMessage): void => {
      console.log(`[useApi] Handling message of type: ${message.type}`);
      switch (message.type) {
        case "success":
          console.log(
            `[useApi] Move success: row=${message.row}, col=${message.col}, value=${message.value}`
          );
          setLastMoveSuccess(true);
          setLastError(null);

          // The server will automatically broadcast the updated board state to all clients
          // after sending the success message, so we don't need to request it manually
          break;

        case "error":
          console.log(`[useApi] Move error: ${message.error}`);
          setLastMoveSuccess(false);
          setLastError(message.error);
          break;

        case "state":
          console.log(
            `[useApi] Received game state, board has ${message.board.length} rows`
          );
          setBoard(message.board);

          // Handle the initial board if it's included in the message
          if ("initialBoard" in message && message.initialBoard) {
            console.log(
              `[useApi] Received initial board, has ${message.initialBoard.length} rows`
            );
            setInitialBoard(message.initialBoard);
          }
          break;

        case "board_update":
          console.log(
            `[useApi] Received board update, board has ${message.board.length} rows`
          );
          setBoard(message.board);
          break;

        case "move_notification":
          console.log(
            `[useApi] Received move notification: row=${message.row}, col=${message.col}, value=${message.value}`
          );
          setBoard(message.board);
          break;

        default:
          console.warn("[useApi] Unknown message type received:", message);
      }
    },
    [] // Remove dependencies to prevent unnecessary re-creation
  );

  // Return the hook's API
  return {
    // Connection state
    connectionState,
    connectionError,

    // HTTP API methods
    createLobby,

    // WebSocket connection method
    joinLobby,

    // Game state
    board,
    initialBoard,

    // WebSocket methods
    makeMove,
    clearCell,
    requestGameState,

    // Last message information
    lastError,
    lastMoveSuccess,
  };
};
