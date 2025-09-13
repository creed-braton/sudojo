import { useState, type ReactElement, useEffect } from "react";
import { lobbyApi } from "../../services/api";
import styles from "./styles.module.css";

interface JoinLobbyProps {
  onBack: () => void;
  onJoinSuccess: (lobbyId: string) => void;
}

const JoinLobby = ({ onBack, onJoinSuccess }: JoinLobbyProps): ReactElement => {
  const [lobbyId, setLobbyId] = useState<string>("");
  const [isConnecting, setIsConnecting] = useState<boolean>(false);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const [error, setError] = useState<string>("");
  const [messages, setMessages] = useState<unknown[]>([]);

  // Clean up WebSocket connection on unmount
  useEffect(() => {
    return () => {
      if (isConnected) {
        lobbyApi.closeConnection();
      }
    };
  }, [isConnected]);

  // Set up message listener when connected
  useEffect(() => {
    if (isConnected) {
      const unsubscribe = lobbyApi.onMessage((message) => {
        setMessages((prev) => [...prev, message]);
      });

      // Clean up listener when component unmounts or disconnects
      return () => {
        unsubscribe();
      };
    }
  }, [isConnected]);

  const handleJoinLobby = async (): Promise<void> => {
    if (!lobbyId.trim()) {
      setError("Please enter a lobby ID");
      return;
    }

    setIsConnecting(true);
    setError("");

    try {
      await lobbyApi.joinLobby(lobbyId);
      setIsConnected(true);
      onJoinSuccess(lobbyId);
    } catch (err) {
      setError("Failed to join lobby. Please check the ID and try again.");
      console.error(err);
    } finally {
      setIsConnecting(false);
    }
  };

  const handleSendMessage = (): void => {
    if (isConnected) {
      const message = {
        type: "CHAT",
        content: "Hello from a new user!",
        timestamp: new Date().toISOString(),
      };

      lobbyApi.sendMessage(message);
    }
  };

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Join Lobby</h1>

      {!isConnected ? (
        <div className={styles.joinForm}>
          <label htmlFor="lobbyId" className={styles.label}>
            Enter Lobby ID:
          </label>
          <input
            id="lobbyId"
            type="text"
            className={styles.input}
            value={lobbyId}
            onChange={(e) => setLobbyId(e.target.value)}
            placeholder="Paste lobby ID here"
            disabled={isConnecting}
          />

          {error && <p className={styles.error}>{error}</p>}

          <button
            className={styles.joinButton}
            onClick={handleJoinLobby}
            disabled={isConnecting || !lobbyId.trim()}
          >
            {isConnecting ? "Connecting..." : "Join Lobby"}
          </button>
        </div>
      ) : (
        <div className={styles.lobbyContent}>
          <p className={styles.connectionStatus}>
            Connected to lobby:{" "}
            <span className={styles.lobbyIdText}>{lobbyId}</span>
          </p>

          <div className={styles.messagesContainer}>
            {messages.length > 0 ? (
              messages.map((msg, index) => (
                <div key={index} className={styles.message}>
                  <pre>{JSON.stringify(msg, null, 2)}</pre>
                </div>
              ))
            ) : (
              <p className={styles.noMessages}>No messages yet</p>
            )}
          </div>

          <button className={styles.sendButton} onClick={handleSendMessage}>
            Send Test Message
          </button>
        </div>
      )}

      <button
        className={styles.backButton}
        onClick={() => {
          if (isConnected) {
            lobbyApi.closeConnection();
          }
          onBack();
        }}
      >
        Back to Home
      </button>
    </div>
  );
};

export default JoinLobby;
