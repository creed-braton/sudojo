import { type ReactElement } from "react";
import styles from "./InputBar.module.css";
import DeleteIcon from "../../icons/DeleteIcon";
import PencilIcon from "../../icons/PencilIcon";
import PingIcon from "../../icons/PingIcon";
import type { Mode } from "../../hooks/sudoku";

const InputBar = ({
  input,
  mode,
  togglePencil,
  togglePing,
}: {
  input: (value: number) => void;
  mode: Mode;
  togglePencil: () => void;
  togglePing: () => void;
}): ReactElement => {
  return (
    <div className={`glass-container ${styles.inputbar}`}>
      <div className={styles.row}>
        {[1, 2, 3, 4, 5].map((num: number) => (
          <button
            key={num}
            className={styles.button}
            onClick={() => {
              input(num);
            }}
          >
            {num}
          </button>
        ))}
        <button
          className={`${styles.button} ${mode === "ping" ? styles.active : ""}`}
          onClick={togglePing}
          title={
            mode === "ping" ? "Switch to normal mode" : "Switch to ping mode"
          }
        >
          <PingIcon className={styles.icon} />
        </button>
      </div>
      <div className={styles.row}>
        {[6, 7, 8, 9, 0].map((num: number) => (
          <button
            key={num}
            className={styles.button}
            onClick={() => {
              input(num);
            }}
          >
            {num !== 0 ? num : <DeleteIcon className={styles.icon} />}
          </button>
        ))}
        <button
          className={`${styles.button} ${mode === "pencil" ? styles.active : ""}`}
          onClick={togglePencil}
          title={
            mode === "pencil"
              ? "Switch to normal mode"
              : "Switch to pencil mode"
          }
        >
          <PencilIcon className={styles.icon} />
        </button>
      </div>
    </div>
  );
};

export default InputBar;
