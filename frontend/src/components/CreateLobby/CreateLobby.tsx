import { useState, useEffect, type ReactElement } from "react";
import { lobbyApi } from "../../services/api";
import styles from "./styles.module.css";

interface CreateLobbyProps {
  onBack: () => void;
  onJoinSuccess?: (lobbyId: string) => void;
}

const CreateLobby = ({
  onBack,
  onJoinSuccess,
}: CreateLobbyProps): ReactElement => {
  const [lobbyId, setLobbyId] = useState<string>("");
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    const createNewLobby = async (): Promise<void> => {
      setIsLoading(true);
      setError("");

      try {
        const id = await lobbyApi.createLobby();
        setLobbyId(id);
      } catch (err) {
        setError("Failed to create lobby. Please try again.");
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    };

    createNewLobby();
  }, []);

  const handleCopyToClipboard = (): void => {
    navigator.clipboard
      .writeText(lobbyId)
      .then(() => {
        alert("Lobby ID copied to clipboard!");
      })
      .catch(() => {
        alert("Failed to copy. Please manually select and copy the ID.");
      });
  };

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Create Lobby</h1>

      {isLoading && <p className={styles.loading}>Creating lobby...</p>}

      {error && <p className={styles.error}>{error}</p>}

      {lobbyId && !isLoading && (
        <div className={styles.lobbyInfo}>
          <p className={styles.infoText}>
            Your lobby has been created! Share this ID with others to join:
          </p>
          <div className={styles.idContainer}>
            <span className={styles.lobbyId}>{lobbyId}</span>
            <button
              className={styles.copyButton}
              onClick={handleCopyToClipboard}
            >
              Copy
            </button>
          </div>
        </div>
      )}

      <div className={styles.buttonContainer}>
        {lobbyId && !isLoading && onJoinSuccess && (
          <button
            className={styles.joinButton}
            onClick={async () => {
              try {
                setIsLoading(true);
                console.log(
                  "Establishing WebSocket connection to lobby:",
                  lobbyId
                );
                // First establish the WebSocket connection
                await lobbyApi.joinLobby(lobbyId);
                console.log("WebSocket connection established successfully");
                // Then transition to the game
                onJoinSuccess(lobbyId);
              } catch (err) {
                console.error("Failed to join lobby:", err);
                setError("Failed to connect to the lobby. Please try again.");
                setIsLoading(false);
              }
            }}
          >
            Join Game
          </button>
        )}
        <button className={styles.backButton} onClick={onBack}>
          Back to Home
        </button>
      </div>
    </div>
  );
};

export default CreateLobby;
