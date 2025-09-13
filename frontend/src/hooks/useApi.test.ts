import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useApi } from "./useApi";

// Types from useApi.ts that we need for testing
type ConnectionState = "disconnected" | "connecting" | "connected" | "error";
type Board = number[][];

// Interface matching the return type of useApi hook for type safety
type UseApiReturn = {
  connectionState: ConnectionState;
  connectionError: string | null;
  board: Board | null;
  lastError: string | null;
  lastMoveSuccess: boolean | null;
  createLobby: () => Promise<string>;
  joinLobby: (lobbyId: string) => void;
  makeMove: (row: number, col: number, value: number) => void;
  requestGameState: () => void;
};

type RenderHookResult = {
  result: {
    current: UseApiReturn;
  };
  rerender: () => void;
  unmount: () => void;
};

/**
 * Waits for a condition to be true
 */
const waitForCondition = async (
  condition: () => boolean,
  options: { timeout?: number; interval?: number; errorMessage?: string } = {}
): Promise<void> => {
  const {
    timeout = 10000, // Increased timeout to 10 seconds
    interval = 100,
    errorMessage = "Condition not met within timeout period",
  } = options;

  const startTime = Date.now();
  let lastCheckTime = startTime;
  let checkCount = 0;

  while (Date.now() - startTime < timeout) {
    if (condition()) {
      console.log(
        `[Test] Condition met after ${
          Date.now() - startTime
        }ms and ${checkCount} checks`
      );
      return;
    }

    // Log progress every second
    const now = Date.now();
    if (now - lastCheckTime > 1000) {
      console.log(
        `[Test] Still waiting for condition after ${now - startTime}ms...`
      );
      lastCheckTime = now;
    }

    checkCount++;
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  throw new Error(
    `${errorMessage} (Timeout after ${timeout}ms and ${checkCount} checks)`
  );
};

// Backend configuration for tests
const TEST_API_URL = "localhost:8080";

describe("useApi - Integration Tests with Real Backend", () => {
  // Track active hooks so we can clean them up after tests
  const activeHooks: RenderHookResult[] = [];
  let testLobbyId: string | null = null;

  // Increase timeout for real network requests
  vi.setConfig({ testTimeout: 20000 }); // Double the timeout to 20 seconds

  console.log(
    `[Test] Starting integration tests with backend at ${TEST_API_URL}`
  );

  // Clean up any active hooks after each test
  afterEach(() => {
    // Unmount any hooks that were created during the test
    activeHooks.forEach((hook) => {
      try {
        hook.unmount();
      } catch (error) {
        console.warn("Error unmounting hook:", error);
      }
    });

    // Clear the active hooks array
    activeHooks.length = 0;
  });

  // Helper function to render a hook and track it for cleanup
  const renderApiHook = (): RenderHookResult => {
    const hook = renderHook(() =>
      useApi({ baseUrl: TEST_API_URL })
    ) as RenderHookResult;
    activeHooks.push(hook);
    return hook;
  };

  it("should initialize with default values", () => {
    const { result } = renderApiHook();

    expect(result.current.connectionState).toBe("disconnected");
    expect(result.current.connectionError).toBeNull();
    expect(result.current.board).toBeNull();
    expect(result.current.lastError).toBeNull();
    expect(result.current.lastMoveSuccess).toBeNull();
  });

  // HTTP API Tests
  describe("createLobby", () => {
    it("should create a lobby successfully", async () => {
      const { result } = renderApiHook();

      // Call createLobby
      const lobbyId = await result.current.createLobby();

      // Store the lobby ID for later tests
      testLobbyId = lobbyId;

      // Verify we got a valid lobby ID
      expect(lobbyId).toBeDefined();
      expect(typeof lobbyId).toBe("string");
      expect(lobbyId.length).toBeGreaterThan(0);
    });
  });

  // WebSocket Connection Tests
  describe("joinLobby", () => {
    it("should establish WebSocket connection when joining a lobby", async () => {
      // We need a lobby ID to join, so create one if we don't have one
      if (!testLobbyId) {
        const createHook = renderApiHook();
        testLobbyId = await createHook.result.current.createLobby();
      }

      const { result } = renderApiHook();

      // Initial state should be disconnected
      expect(result.current.connectionState).toBe("disconnected");

      // Join the lobby - wrap in act because it triggers state changes
      act(() => {
        result.current.joinLobby(testLobbyId as string);
      });

      // Should immediately change to connecting
      expect(result.current.connectionState).toBe("connecting");

      // Wait for connection to establish (with longer timeout for real network)
      await waitForCondition(
        () => result.current.connectionState === "connected",
        {
          timeout: 10000,
          errorMessage: "WebSocket connection not established within timeout",
        }
      );

      console.log("[Test] WebSocket connection established successfully");

      // Verify connection state
      expect(result.current.connectionError).toBeNull();
    });
  });

  // WebSocket Message Sending and Receiving Tests
  describe("game interactions", () => {
    let apiHook: RenderHookResult;

    beforeEach(async () => {
      // If we don't have a lobby ID, create one
      if (!testLobbyId) {
        const createHook = renderApiHook();
        testLobbyId = await createHook.result.current.createLobby();
      }

      // Create a new hook instance for the test
      apiHook = renderApiHook();

      // Join the lobby
      act(() => {
        apiHook.result.current.joinLobby(testLobbyId as string);
      });

      // Wait for the connection to establish (with longer timeout for real network)
      await waitForCondition(
        () => apiHook.result.current.connectionState === "connected",
        {
          timeout: 10000,
          errorMessage: "WebSocket connection not established in beforeEach",
        }
      );

      console.log("[Test] beforeEach: WebSocket connection established");
    });

    it("should request and receive game state", async () => {
      // Request the game state
      act(() => {
        apiHook.result.current.requestGameState();
      });

      console.log("[Test] Requesting game state");

      // Wait for the board to be populated (with longer timeout for real network)
      await waitForCondition(() => apiHook.result.current.board !== null, {
        timeout: 10000,
        errorMessage: "Did not receive game state within timeout",
      });

      console.log("[Test] Game state received successfully");

      // Verify that we received a valid board
      const board = apiHook.result.current.board;
      expect(Array.isArray(board)).toBe(true);
      expect(board?.length).toBe(9); // Sudoku board is 9x9

      // Verify that each row is also an array of length 9
      board?.forEach((row: number[]) => {
        expect(Array.isArray(row)).toBe(true);
        expect(row.length).toBe(9);
      });
    });

    it("should make a move and receive response", async () => {
      // First request the game state to see the current board
      act(() => {
        apiHook.result.current.requestGameState();
      });

      console.log("[Test] Requesting game state before making move");

      // Wait for the board to be populated (with longer timeout for real network)
      await waitForCondition(() => apiHook.result.current.board !== null, {
        timeout: 10000,
        errorMessage: "Did not receive game state before attempting move",
      });

      console.log("[Test] Game state received, preparing to make move");

      const board = apiHook.result.current.board;

      // Find an empty cell (with value 0)
      let emptyRow = -1;
      let emptyCol = -1;

      for (let row = 0; row < 9; row++) {
        for (let col = 0; col < 9; col++) {
          if (board && board[row][col] === 0) {
            emptyRow = row;
            emptyCol = col;
            break;
          }
        }
        if (emptyRow !== -1) break;
      }

      // If no empty cell found, the test can't proceed
      expect(emptyRow).not.toBe(-1);
      expect(emptyCol).not.toBe(-1);

      // Store the current state before making move
      const initialMoveSuccess = apiHook.result.current.lastMoveSuccess;
      console.log(
        `[Test] Found empty cell at row=${emptyRow}, col=${emptyCol}, placing value 5`
      );

      // Try placing a value in the empty cell
      act(() => {
        apiHook.result.current.makeMove(emptyRow, emptyCol, 5);
      });

      // Wait for a response (either success or error) with longer timeout for real network
      await waitForCondition(
        () => apiHook.result.current.lastMoveSuccess !== initialMoveSuccess,
        {
          timeout: 10000,
          errorMessage: "Did not receive move response within timeout",
        }
      );

      console.log(
        `[Test] Received move response: success=${apiHook.result.current.lastMoveSuccess}, error=${apiHook.result.current.lastError}`
      );

      // Verify we got some kind of response
      expect(apiHook.result.current.lastMoveSuccess).not.toBeNull();

      // If move failed, we should have an error message
      if (apiHook.result.current.lastMoveSuccess === false) {
        expect(apiHook.result.current.lastError).not.toBeNull();
      }
    });
  });

  // Multi-client integration test
  describe("multi-client real-time synchronization", () => {
    it("should synchronize moves across 5 WebSocket connections", async () => {
      console.log("[Test] Starting multi-client synchronization test");

      // Create a lobby first
      const creatorHook = renderApiHook();
      const lobbyId = await creatorHook.result.current.createLobby();
      console.log(`[Test] Created lobby: ${lobbyId}`);

      // Create 5 client connections
      const clients: RenderHookResult[] = [];
      for (let i = 0; i < 5; i++) {
        const client = renderApiHook();
        clients.push(client);
        console.log(`[Test] Created client ${i + 1}`);
      }

      // Connect all clients to the same lobby
      console.log("[Test] Connecting all clients to lobby...");
      for (let i = 0; i < clients.length; i++) {
        act(() => {
          clients[i].result.current.joinLobby(lobbyId);
        });
        console.log(`[Test] Client ${i + 1} joining lobby`);
      }

      // Wait for all clients to connect
      console.log("[Test] Waiting for all clients to connect...");
      for (let i = 0; i < clients.length; i++) {
        await waitForCondition(
          () => clients[i].result.current.connectionState === "connected",
          {
            timeout: 15000,
            errorMessage: `Client ${i + 1} failed to connect within timeout`,
          }
        );
        console.log(`[Test] Client ${i + 1} connected successfully`);
      }

      // Request initial game state for all clients
      console.log("[Test] Requesting initial game state for all clients...");
      for (let i = 0; i < clients.length; i++) {
        act(() => {
          clients[i].result.current.requestGameState();
        });
      }

      // Wait for all clients to receive the initial board
      console.log("[Test] Waiting for all clients to receive initial board...");
      for (let i = 0; i < clients.length; i++) {
        await waitForCondition(() => clients[i].result.current.board !== null, {
          timeout: 15000,
          errorMessage: `Client ${
            i + 1
          } did not receive initial board within timeout`,
        });
        console.log(`[Test] Client ${i + 1} received initial board`);
      }

      // Verify all clients have the same initial board
      const referenceBoard = clients[0].result.current.board;
      expect(referenceBoard).not.toBeNull();
      expect(Array.isArray(referenceBoard)).toBe(true);
      expect(referenceBoard?.length).toBe(9);

      for (let i = 1; i < clients.length; i++) {
        const clientBoard = clients[i].result.current.board;
        expect(clientBoard).toEqual(referenceBoard);
        console.log(`[Test] Client ${i + 1} board matches reference board`);
      }

      // Find 5 empty cells for testing
      const emptyCells: Array<{ row: number; col: number; value: number }> = [];
      if (referenceBoard) {
        for (let row = 0; row < 9 && emptyCells.length < 5; row++) {
          for (let col = 0; col < 9 && emptyCells.length < 5; col++) {
            if (referenceBoard[row][col] === 0) {
              emptyCells.push({
                row,
                col,
                value: (emptyCells.length % 9) + 1, // Use values 1-5
              });
            }
          }
        }
      }

      expect(emptyCells.length).toBeGreaterThanOrEqual(5);
      console.log(`[Test] Found ${emptyCells.length} empty cells for testing`);

      // Have each client make a move sequentially
      for (let clientIndex = 0; clientIndex < 5; clientIndex++) {
        const move = emptyCells[clientIndex];
        const client = clients[clientIndex];

        console.log(
          `[Test] Client ${clientIndex + 1} making move: row=${move.row}, col=${
            move.col
          }, value=${move.value}`
        );

        // Store the current board state before making the move
        const boardBeforeMove = clients[0].result.current.board;
        const cellValueBefore = boardBeforeMove
          ? boardBeforeMove[move.row][move.col]
          : null;

        // Make the move
        act(() => {
          client.result.current.makeMove(move.row, move.col, move.value);
        });

        // Wait for all clients to receive the board update (this is the real test)
        // If the move is successful, all clients should get the updated board
        console.log(
          `[Test] Waiting for all clients to receive board update for move by client ${
            clientIndex + 1
          }...`
        );

        for (let i = 0; i < clients.length; i++) {
          await waitForCondition(
            () => {
              const board = clients[i].result.current.board;
              // Check if the board has been updated with the move
              // Either the move succeeded (cell has the new value) or failed (cell unchanged but we got a response)
              if (board === null) return false;

              const currentCellValue = board[move.row][move.col];
              const boardChanged = currentCellValue !== cellValueBefore;
              const moveSucceeded = currentCellValue === move.value;

              // We consider the update received if either:
              // 1. The move succeeded and the cell has the new value
              // 2. The move failed but we have some indication (lastMoveSuccess is not null)
              return (
                moveSucceeded ||
                clients[i].result.current.lastMoveSuccess !== null
              );
            },
            {
              timeout: 10000,
              errorMessage: `Client ${
                i + 1
              } did not receive board update for move by client ${
                clientIndex + 1
              } within timeout`,
            }
          );

          const board = clients[i].result.current.board;
          const moveSucceeded =
            board && board[move.row][move.col] === move.value;
          console.log(
            `[Test] Client ${i + 1} received update for move by client ${
              clientIndex + 1
            }: ${moveSucceeded ? "SUCCESS" : "FAILED/ERROR"}`
          );
        }

        // Verify all clients have the same board state after the move attempt
        const referenceBoard = clients[0].result.current.board;
        for (let i = 1; i < clients.length; i++) {
          const clientBoard = clients[i].result.current.board;
          expect(clientBoard).toEqual(referenceBoard);
        }

        const moveSucceeded =
          referenceBoard && referenceBoard[move.row][move.col] === move.value;
        console.log(
          `[Test] Move by client ${clientIndex + 1} ${
            moveSucceeded ? "succeeded" : "failed"
          }, all clients synchronized`
        );

        // Small delay between moves to avoid overwhelming the server
        await new Promise((resolve) => setTimeout(resolve, 500));
      }

      // Final verification: ensure all clients still have synchronized boards
      console.log("[Test] Final synchronization check...");
      const finalReferenceBoard = clients[0].result.current.board;
      for (let i = 1; i < clients.length; i++) {
        const clientBoard = clients[i].result.current.board;
        expect(clientBoard).toEqual(finalReferenceBoard);
      }

      console.log(
        "[Test] Multi-client synchronization test completed successfully!"
      );

      // Verify that at least some moves were successful
      const successfulMoves = clients.filter(
        (c) => c.result.current.lastMoveSuccess === true
      ).length;
      console.log(`[Test] ${successfulMoves} out of 5 moves were successful`);

      // We expect at least one move to be successful in a typical Sudoku game
      expect(successfulMoves).toBeGreaterThan(0);
    });

    it("should handle concurrent moves from multiple clients", async () => {
      console.log("[Test] Starting concurrent moves test");

      // Create a lobby
      const creatorHook = renderApiHook();
      const lobbyId = await creatorHook.result.current.createLobby();
      console.log(`[Test] Created lobby: ${lobbyId}`);

      // Create 3 client connections for this test
      const clients: RenderHookResult[] = [];
      for (let i = 0; i < 3; i++) {
        const client = renderApiHook();
        clients.push(client);
      }

      // Connect all clients
      for (let i = 0; i < clients.length; i++) {
        act(() => {
          clients[i].result.current.joinLobby(lobbyId);
        });
      }

      // Wait for all clients to connect and receive initial board
      for (let i = 0; i < clients.length; i++) {
        await waitForCondition(
          () => clients[i].result.current.connectionState === "connected",
          { timeout: 15000 }
        );

        act(() => {
          clients[i].result.current.requestGameState();
        });

        await waitForCondition(() => clients[i].result.current.board !== null, {
          timeout: 15000,
        });
      }

      // Find empty cells for concurrent moves
      const board = clients[0].result.current.board;
      const emptyCells: Array<{ row: number; col: number; value: number }> = [];

      if (board) {
        for (let row = 0; row < 9 && emptyCells.length < 3; row++) {
          for (let col = 0; col < 9 && emptyCells.length < 3; col++) {
            if (board[row][col] === 0) {
              emptyCells.push({
                row,
                col,
                value: (emptyCells.length % 9) + 1,
              });
            }
          }
        }
      }

      expect(emptyCells.length).toBeGreaterThanOrEqual(3);

      // Store initial board state
      const initialBoard = clients[0].result.current.board;
      const initialCellValues = emptyCells.map((cell) =>
        initialBoard ? initialBoard[cell.row][cell.col] : 0
      );

      // Make concurrent moves
      console.log("[Test] Making concurrent moves...");
      for (let i = 0; i < 3; i++) {
        const move = emptyCells[i];
        act(() => {
          clients[i].result.current.makeMove(move.row, move.col, move.value);
        });
      }

      // Wait for all clients to receive board updates
      // We check that the board state has changed from the initial state
      for (let i = 0; i < clients.length; i++) {
        await waitForCondition(
          () => {
            const currentBoard = clients[i].result.current.board;
            if (!currentBoard || !initialBoard) return false;

            // Check if any of the target cells have been updated
            let boardChanged = false;
            for (let j = 0; j < emptyCells.length; j++) {
              const cell = emptyCells[j];
              const initialValue = initialCellValues[j];
              const currentValue = currentBoard[cell.row][cell.col];
              if (currentValue !== initialValue) {
                boardChanged = true;
                break;
              }
            }

            // Also check if we have any move response
            const hasResponse =
              clients[i].result.current.lastMoveSuccess !== null;

            return boardChanged || hasResponse;
          },
          {
            timeout: 15000,
            errorMessage: `Client ${
              i + 1
            } did not receive response to concurrent move`,
          }
        );
      }

      // Wait a bit more for all board updates to propagate
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Verify all clients have synchronized boards
      const finalBoard = clients[0].result.current.board;
      for (let i = 1; i < clients.length; i++) {
        expect(clients[i].result.current.board).toEqual(finalBoard);
      }

      console.log("[Test] Concurrent moves test completed successfully!");
    });
  });
});
