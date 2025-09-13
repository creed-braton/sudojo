import { useState, useEffect, type ReactElement } from "react";
import { lobbyApi } from "../../services/api";
import styles from "./styles.module.css";

// Constants for Sudoku
const BOARD_SIZE = 9;
const EMPTY_CELL = 0;

// Types for Sudoku messages
interface SudokuMoveMessage {
  type: "move";
  row: number;
  col: number;
  value: number;
}

interface SudokuSuccessMessage {
  type: "success";
  row: number;
  col: number;
  value: number;
  success: true;
}

interface SudokuErrorMessage {
  type: "error";
  success: false;
  error: string;
}

interface SudokuStateMessage {
  type: "state";
  board: number[][];
}

// Union type for all possible server messages
type ServerMessage =
  | SudokuSuccessMessage
  | SudokuErrorMessage
  | SudokuStateMessage;

interface SudokuGameProps {
  onBack: () => void;
  lobbyId: string;
}

const SudokuGame = ({ onBack, lobbyId }: SudokuGameProps): ReactElement => {
  // Initialize an empty 9x9 board filled with zeros
  const [board, setBoard] = useState<number[][]>(
    Array(BOARD_SIZE)
      .fill(0)
      .map(() => Array(BOARD_SIZE).fill(EMPTY_CELL))
  );

  // Track initial cells (those that were pre-filled by the backend)
  const [initialCells, setInitialCells] = useState<boolean[][]>(
    Array(BOARD_SIZE)
      .fill(false)
      .map(() => Array(BOARD_SIZE).fill(false))
  );

  const [selectedCell, setSelectedCell] = useState<[number, number] | null>(
    null
  );
  const [error, setError] = useState<string>("");
  const [status, setStatus] = useState<string>(
    `Connecting to lobby: ${lobbyId}...`
  );
  const [isConnected, setIsConnected] = useState<boolean>(
    lobbyApi.isConnected()
  );
  const [boardRequested, setBoardRequested] = useState<boolean>(false);

  // Check connection status and request initial board state
  useEffect(() => {
    const checkConnectionAndRequestState = () => {
      const connected = lobbyApi.isConnected();
      setIsConnected(connected);

      if (connected && !boardRequested) {
        console.log("Connection confirmed, requesting board state");
        setStatus(`Connected to lobby: ${lobbyId}. Requesting game state...`);

        const requestStateMessage = {
          type: "request_state",
        };

        const success = lobbyApi.sendMessage(requestStateMessage);
        console.log("Request state message sent successfully:", success);

        if (success) {
          setBoardRequested(true);
        }
      } else if (!connected) {
        setError("Not connected to lobby. Please go back and try again.");
        setStatus(`Connection to lobby failed. Please try again.`);
      }
    };

    // Initial check
    checkConnectionAndRequestState();

    // Set up a periodic check for connection status
    const connectionCheckInterval = setInterval(() => {
      checkConnectionAndRequestState();
    }, 1000);

    // Set up message listener
    const unsubscribe = lobbyApi.onMessage((message) => {
      console.log("Received message:", message);
      handleServerMessage(message as ServerMessage);
    });

    // Clean up listener and interval when component unmounts
    return () => {
      unsubscribe();
      clearInterval(connectionCheckInterval);
    };
  }, [lobbyId, boardRequested]);

  // Handle incoming messages from the server
  const handleServerMessage = (message: ServerMessage) => {
    console.log("Processing message of type:", message.type);

    switch (message.type) {
      case "success":
        // Update the board with the new value
        setBoard((prevBoard) => {
          const newBoard = [...prevBoard.map((row) => [...row])];
          newBoard[message.row][message.col] = message.value;
          return newBoard;
        });

        // Clear any previous errors
        setError("");
        break;

      case "error":
        // Display error message
        setError(message.error || "An error occurred");
        break;

      case "state":
        console.log("Received state message with board:", message.board);

        if (!message.board || !Array.isArray(message.board)) {
          setError("Received invalid board data");
          console.error("Invalid board data:", message.board);
          return;
        }

        // Set the initial board state
        setBoard(message.board);

        // Mark initial cells (non-empty cells)
        const initialCellsMap = message.board.map((row) =>
          row.map((cell) => cell !== EMPTY_CELL)
        );
        setInitialCells(initialCellsMap);

        setStatus("Game started. Your move!");
        setError("");
        break;

      default:
        console.log("Unhandled message type:", (message as any).type);
        break;
    }
  };

  // Handle cell selection
  const handleCellClick = (row: number, col: number) => {
    // Don't allow selection of initial cells
    if (initialCells[row][col]) {
      return;
    }

    setSelectedCell([row, col]);
    setError(""); // Clear any previous errors
  };

  // Handle number input
  const handleNumberClick = (value: number) => {
    if (!selectedCell) {
      setError("Please select a cell first");
      return;
    }

    const [row, col] = selectedCell;

    // Don't allow changing initial cells
    if (initialCells[row][col]) {
      setError("Cannot change initial numbers");
      return;
    }

    // Send move to server
    const moveMessage: SudokuMoveMessage = {
      type: "move",
      row,
      col,
      value,
    };

    lobbyApi.sendMessage(moveMessage);
  };

  // Handle clearing a cell (sending value 0)
  const handleClearCell = () => {
    if (!selectedCell) {
      setError("Please select a cell first");
      return;
    }

    const [row, col] = selectedCell;

    // Don't allow clearing initial cells
    if (initialCells[row][col]) {
      setError("Cannot clear initial numbers");
      return;
    }

    // Send clear move to server (value 0)
    const moveMessage: SudokuMoveMessage = {
      type: "move",
      row,
      col,
      value: EMPTY_CELL,
    };

    lobbyApi.sendMessage(moveMessage);
  };

  // Render the Sudoku board
  const renderBoard = () => {
    const cells = [];

    for (let row = 0; row < BOARD_SIZE; row++) {
      for (let col = 0; col < BOARD_SIZE; col++) {
        const value = board[row][col];
        const isInitial = initialCells[row][col];
        const isSelected =
          selectedCell && selectedCell[0] === row && selectedCell[1] === col;

        cells.push(
          <div
            key={`${row}-${col}`}
            className={`${styles.cell} ${isInitial ? styles.initialCell : ""} ${
              isSelected ? styles.selected : ""
            }`}
            onClick={() => handleCellClick(row, col)}
          >
            {value !== EMPTY_CELL ? value : ""}
          </div>
        );
      }
    }

    return cells;
  };

  // Render number buttons
  const renderNumberButtons = () => {
    const buttons = [];

    for (let i = 1; i <= 9; i++) {
      buttons.push(
        <button
          key={i}
          className={styles.numberButton}
          onClick={() => handleNumberClick(i)}
        >
          {i}
        </button>
      );
    }

    return buttons;
  };

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Sudoku Game</h1>
      <p className={styles.status}>{status}</p>

      {error && <p className={styles.error}>{error}</p>}

      {isConnected ? (
        <>
          <div className={styles.board}>{renderBoard()}</div>

          <div className={styles.controls}>
            <div className={styles.numberButtons}>
              {renderNumberButtons()}
              <button
                className={styles.clearButton}
                onClick={handleClearCell}
                disabled={!isConnected}
              >
                Clear
              </button>
            </div>
          </div>
        </>
      ) : (
        <div className={styles.connectionError}>
          <p>
            Connection to the game server has been lost or could not be
            established.
          </p>
          <p>Please return to the home screen and try again.</p>
        </div>
      )}

      <button
        className={styles.backButton}
        onClick={() => {
          lobbyApi.closeConnection();
          onBack();
        }}
      >
        Leave Game
      </button>

      {!isConnected && boardRequested && (
        <button
          className={styles.retryButton}
          onClick={() => {
            // Reset state and try to reconnect
            setBoardRequested(false);
            setStatus(`Reconnecting to lobby: ${lobbyId}...`);
            lobbyApi.closeConnection();

            // Try to rejoin the lobby
            lobbyApi
              .joinLobby(lobbyId)
              .then(() => {
                setIsConnected(true);
                setStatus(`Reconnected to lobby: ${lobbyId}`);
                setError("");
              })
              .catch((err) => {
                console.error("Failed to reconnect:", err);
                setError("Failed to reconnect to the lobby");
              });
          }}
        >
          Retry Connection
        </button>
      )}
    </div>
  );
};

export default SudokuGame;
