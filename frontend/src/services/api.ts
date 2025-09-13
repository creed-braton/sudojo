/**
 * API client for backend communication
 * Handles both HTTP requests and WebSocket connections
 */

// Types
interface LobbyResponse {
  id: string;
}

type MessageCallback = (message: unknown) => void;

interface LobbyConnection {
  ws: WebSocket | null;
  lobbyId: string;
  isConnected: boolean;
  messageCallbacks: Set<MessageCallback>;
}

// API base URL - can be configured based on environment
const API_BASE_URL = "http://localhost:8080"; // Replace with actual API URL

/**
 * Lobby API service
 * Manages both REST and WebSocket communication with the lobby backend
 */
class LobbyApi {
  private connection: LobbyConnection = {
    ws: null,
    lobbyId: "",
    isConnected: false,
    messageCallbacks: new Set(),
  };

  /**
   * Creates a new lobby
   * @returns Promise with the lobby ID
   */
  async createLobby(): Promise<string> {
    try {
      const response = await fetch(`${API_BASE_URL}/lobby`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to create lobby: ${response.status}`);
      }

      const data = (await response.json()) as LobbyResponse;
      return data.id;
    } catch (error) {
      console.error("Error creating lobby:", error);
      throw error;
    }
  }

  /**
   * Joins an existing lobby via WebSocket
   * @param lobbyId The ID of the lobby to join
   * @returns Promise that resolves when connection is established
   */
  joinLobby(lobbyId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      // Close existing connection if there is one
      this.closeConnection();

      // Create new WebSocket connection
      const ws = new WebSocket(
        `${API_BASE_URL.replace("http", "ws")}/lobby?id=${lobbyId}`
      );

      // Store connection info
      this.connection.ws = ws;
      this.connection.lobbyId = lobbyId;

      // Handle WebSocket events
      ws.onopen = () => {
        this.connection.isConnected = true;
        console.log(`Connected to lobby: ${lobbyId}`);
        resolve();
      };

      ws.onclose = () => {
        this.connection.isConnected = false;
        console.log(`Disconnected from lobby: ${lobbyId}`);
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        reject(error);
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.notifyCallbacks(message);
        } catch (error) {
          console.error("Error parsing message:", error);
        }
      };
    });
  }

  /**
   * Sends a message to the current lobby
   * @param message The message to send
   * @returns Boolean indicating if message was sent
   */
  sendMessage(message: unknown): boolean {
    if (!this.connection.isConnected || !this.connection.ws) {
      console.error("Cannot send message: Not connected to a lobby");
      return false;
    }

    try {
      this.connection.ws.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error("Error sending message:", error);
      return false;
    }
  }

  /**
   * Registers a callback for received messages
   * @param callback Function to call when messages are received
   * @returns Function to unregister the callback
   */
  onMessage(callback: MessageCallback): () => void {
    this.connection.messageCallbacks.add(callback);

    // Return function to remove the callback
    return () => {
      this.connection.messageCallbacks.delete(callback);
    };
  }

  /**
   * Closes the current WebSocket connection
   */
  closeConnection(): void {
    if (this.connection.ws) {
      this.connection.ws.close();
      this.connection.ws = null;
      this.connection.isConnected = false;
      this.connection.lobbyId = "";
    }
  }

  /**
   * Gets the current connection status
   * @returns Boolean indicating if connected to a lobby
   */
  isConnected(): boolean {
    return this.connection.isConnected;
  }

  /**
   * Gets the current lobby ID
   * @returns The current lobby ID or empty string if not connected
   */
  getCurrentLobbyId(): string {
    return this.connection.lobbyId;
  }

  /**
   * Notifies all registered callbacks with a message
   * @param message The message to pass to callbacks
   */
  private notifyCallbacks(message: unknown): void {
    this.connection.messageCallbacks.forEach((callback) => {
      try {
        callback(message);
      } catch (error) {
        console.error("Error in message callback:", error);
      }
    });
  }
}

// Export a singleton instance
export const lobbyApi = new LobbyApi();
