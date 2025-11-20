import { type ReactElement } from "react";
import styles from "./InputBar.module.css";
import DeleteIcon from "../../icons/DeleteIcon";
import PencilIcon from "../../icons/PencilIcon";
import PingIcon from "../../icons/PingIcon";

const InputBar = ({
  input,
  pencilMode,
  pingMode,
  togglePencil,
  togglePing,
}: {
  input: (value: number) => void;
  pencilMode: boolean;
  pingMode: boolean;
  togglePencil: () => void;
  togglePing: () => void;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.inputbar}`}>
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
          className={`${styles.button} ${pingMode ? styles.active : ""}`}
          onClick={togglePing}
          title={pingMode ? "Switch to normal mode" : "Switch to ping mode"}
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
          className={`${styles.button} ${pencilMode ? styles.active : ""}`}
          onClick={togglePencil}
          title={pencilMode ? "Switch to normal mode" : "Switch to pencil mode"}
        >
          <PencilIcon className={styles.icon} />
        </button>
      </div>
    </div>
  );
};

export default InputBar;
