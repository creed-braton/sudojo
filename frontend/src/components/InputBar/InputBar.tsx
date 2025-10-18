import { type ReactElement } from "react";
import styles from "./InputBar.module.css";
import type { Cell } from "../../types";
import DeleteIcon from "../../icons/DeleteIcon";
import PencilIcon from "../../icons/PencilIcon";

const InputBar = ({
  position,
  input,
  pencilMode,
  toggleMode,
}: {
  position: Cell | undefined;
  input: (row: number, column: number, value: number) => void;
  pencilMode: boolean;
  toggleMode: () => void;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.inputbar}`}>
      <div className={styles.row}>
        {[1, 2, 3, 4, 5, 6].map((num: number) => (
          <button
            key={num}
            className={styles.button}
            onClick={() => {
              position && input(position.row, position.column, num);
            }}
          >
            {num}
          </button>
        ))}
      </div>
      <div className={styles.row}>
        {[7, 8, 9, 0].map((num: number) => (
          <button
            key={num}
            className={styles.button}
            onClick={() => {
              position && input(position.row, position.column, num);
            }}
          >
            {num !== 0 ? num : <DeleteIcon className={styles.icon} />}
          </button>
        ))}
        <button
          className={`${styles.button} ${pencilMode ? styles.active : ""}`}
          onClick={toggleMode}
          title={pencilMode ? "Switch to normal mode" : "Switch to pencil mode"}
        >
          <PencilIcon className={styles.icon} />
        </button>
      </div>
    </div>
  );
};

export default InputBar;
