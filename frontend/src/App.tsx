import { useEffect, useState, type ReactElement } from "react";
import Layout from "./components/Layout/Layout";
import Board from "./components/Board/Board";
import InputBar from "./components/InputBar/InputBar";
import Welcome from "./components/Welcome/Welcome";
import { useApi } from "./hooks/useApi";
import { isValidSudokuMove } from "./utils/sudokuValidator";
import {
  useNavigate,
  useParams,
  BrowserRouter,
  Routes,
  Route,
} from "react-router-dom";

const GameScreen = (): ReactElement => {
  const { lobbyId } = useParams<{ lobbyId: string }>();
  const {
    joinLobby,
    board,
    initialBoard: serverInitialBoard,
    makeMove,
    clearCell,
    lastMoveSuccess,
    lastError,
  } = useApi();
  const [selectedCell, setSelectedCell] = useState<{
    row: number;
    col: number;
  } | null>(null);

  // Add state for tracking invalid move for visual feedback
  const [invalidCell, setInvalidCell] = useState<{
    row: number;
    col: number;
  } | null>(null);

  // Create a local copy of the board for immediate feedback
  const [localBoard, setLocalBoard] = useState<number[][]>([[]]);

  // Track initial board state to prevent modification of initial numbers
  const [initialBoard, setInitialBoard] = useState<number[][]>([[]]);

  // Flag to lock the entire board during validation
  const [isValidating, setIsValidating] = useState<boolean>(false);

  // Keep a copy of the last valid board state (from server)
  const [lastValidBoard, setLastValidBoard] = useState<number[][]>([[]]);

  // Update local board and initial board when server board changes
  useEffect(() => {
    if (board) {
      // The server's board is the authoritative state
      const boardCopy = JSON.parse(JSON.stringify(board));

      // Always update local and last valid board with server state
      setLocalBoard(boardCopy);
      setLastValidBoard(boardCopy);

      // Unlock the board after receiving server state
      setIsValidating(false);
    }
  }, [board]);

  // Update initial board when server initial board changes
  useEffect(() => {
    if (serverInitialBoard && serverInitialBoard.length > 0) {
      setInitialBoard(JSON.parse(JSON.stringify(serverInitialBoard)));
    }
  }, [serverInitialBoard]);

  useEffect(() => {
    if (lobbyId) {
      joinLobby(lobbyId);
    }
  }, [lobbyId, joinLobby]);

  // Handle server move responses
  useEffect(() => {
    // Handle move success
    if (lastMoveSuccess === true) {
      // Update last valid board to match the current local board
      // which now has the validated move
      setLastValidBoard(JSON.parse(JSON.stringify(localBoard)));

      // Unlock the board
      setIsValidating(false);
    }
    // Handle move rejection
    else if (lastMoveSuccess === false) {
      // Revert to last valid board state from server
      setLocalBoard(JSON.parse(JSON.stringify(lastValidBoard)));

      console.log("Move rejected:", lastError);

      // Unlock the board after a short delay to prevent spamming
      setTimeout(() => {
        setIsValidating(false);
      }, 300);
    }
  }, [lastMoveSuccess, lastError, selectedCell, lastValidBoard, localBoard]);

  const handleCellSelect = (row: number, col: number) => {
    // If board is locked during validation, ignore selection
    if (isValidating) {
      return;
    }

    // Prevent selecting cells with initial values
    if (
      initialBoard.length > row &&
      initialBoard[row].length > col &&
      initialBoard[row][col] !== 0
    ) {
      return;
    }

    // No need to clear invalid cell anymore

    setSelectedCell({ row, col });
  };

  const handleNumberClick = (number: number) => {
    // If board is locked during validation, ignore input
    if (isValidating) {
      console.log("Board is locked during validation, please wait");
      return;
    }

    if (selectedCell && localBoard.length > 0) {
      const { row, col } = selectedCell;

      // Check if this is an initial board value
      if (
        initialBoard.length > row &&
        initialBoard[row].length > col &&
        initialBoard[row][col] !== 0
      ) {
        // Can't modify initial board numbers
        return;
      }

      // First validate move locally
      if (!isValidSudokuMove(localBoard, row, col, number)) {
        console.log("Invalid move detected by client-side validation");

        // Set invalid cell for visual feedback
        setInvalidCell({ row, col });

        // Clear invalid cell state after animation completes
        setTimeout(() => {
          setInvalidCell(null);
        }, 600); // Slightly longer than animation duration (500ms)

        return; // Don't update local board or send to server for invalid moves
      }

      // Lock the board during validation
      setIsValidating(true);

      // Update local board immediately for better user experience
      const newBoard = JSON.parse(JSON.stringify(localBoard));
      newBoard[row][col] = number;
      setLocalBoard(newBoard);

      // Send move to server for final validation
      makeMove(row, col, number);
    }
  };

  const handleClearClick = () => {
    // If board is locked during validation, ignore input
    if (isValidating) {
      console.log("Board is locked during validation, please wait");
      return;
    }

    if (selectedCell && localBoard.length > 0) {
      const { row, col } = selectedCell;

      // Check if this is an initial board value
      if (
        initialBoard.length > row &&
        initialBoard[row].length > col &&
        initialBoard[row][col] !== 0
      ) {
        // Can't clear initial board numbers
        return;
      }

      // Clearing a cell is always valid in Sudoku
      // Lock the board during validation
      setIsValidating(true);

      // Update local board immediately
      const newBoard = JSON.parse(JSON.stringify(localBoard));
      newBoard[row][col] = 0;
      setLocalBoard(newBoard);

      // Use the dedicated clear function
      clearCell(row, col);
    }
  };

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "4px",
        alignItems: "center",
      }}
    >
      <Board
        board={localBoard}
        initialBoard={initialBoard}
        onCellSelect={handleCellSelect}
        selectedCell={selectedCell}
        invalidCell={invalidCell}
      />
      <InputBar
        onNumberClick={handleNumberClick}
        onClearClick={handleClearClick}
      />
    </div>
  );
};

const HomeScreen = (): ReactElement => {
  const navigate = useNavigate();
  const { createLobby } = useApi();

  const handlePlay = async (): Promise<void> => {
    try {
      const lobbyId = await createLobby();
      navigate(`/l/${lobbyId}`);
    } catch (error) {
      console.error("Failed to create lobby:", error);
      // Could add error handling UI here
    }
  };

  return <Welcome onPlay={handlePlay} />;
};

const App = (): ReactElement => {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<HomeScreen />} />
          <Route path="/l/:lobbyId" element={<GameScreen />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
};

export default App;
