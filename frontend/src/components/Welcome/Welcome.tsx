import type { ReactElement } from "react";
import Button from "../Button/Button";
import styles from "./styles.module.css";

type WelcomeProps = {
  onPlay?: () => void;
};

const Welcome = ({ onPlay }: WelcomeProps): ReactElement => {
  return (
    <div className={`${styles.welcome} glassmorphism`}>
      <div className={styles.content}>
        <h1 className={styles.title}>Welcome</h1>
        <p className={styles.message}>have a seat, play a game!</p>
        <div className={styles.buttonContainer}>
          <Button onClick={onPlay}>Play</Button>
        </div>
      </div>
    </div>
  );
};

export default Welcome;
