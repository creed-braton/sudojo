import { type ReactElement } from "react";
import styles from "./styles.module.css";

interface HomeProps {
  onCreateLobby: () => void;
  onJoinLobby: () => void;
}

const Home = ({ onCreateLobby, onJoinLobby }: HomeProps): ReactElement => {
  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Sudojo</h1>
      <div className={styles.buttonContainer}>
        <button className={styles.button} onClick={onCreateLobby}>
          Create Lobby
        </button>
        <button className={styles.button} onClick={onJoinLobby}>
          Join Lobby
        </button>
      </div>
    </div>
  );
};

export default Home;
