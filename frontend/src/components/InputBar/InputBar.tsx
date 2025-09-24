import type { ReactElement } from "react";
import styles from "./InputBar.module.css";
import type { Position } from "../Lobby/Lobby";

const InputBar = ({
  position,
  sendMove,
}: {
  position: Position | undefined;
  sendMove: (row: number, column: number, value: number) => void;
}): ReactElement => {
  return (
    <div className={`glassmorphism ${styles.inputbar}`}>
      <div className={styles.row}>
        {[1, 2, 3, 4, 5].map((num: number) => (
          <button
            className={styles.button}
            onClick={() => {
              position && sendMove(position.row, position.column, num);
            }}
          >
            {num}
          </button>
        ))}
      </div>
      <div className={styles.row}>
        {[6, 7, 8, 9, 0].map((num: number) => (
          <button
            className={styles.button}
            onClick={() => {
              position && sendMove(position.row, position.column, num);
            }}
          >
            {num !== 0 ? (
              num
            ) : (
              <div className={styles.deleteIcon}>
                <svg
                  stroke="currentColor"
                  fill="currentColor"
                  strokeWidth="0"
                  viewBox="0 0 24 24"
                  height="1em"
                  width="1em"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path fill="none" d="M0 0h24v24H0V0z"></path>
                  <path d="M14.12 10.47 12 12.59l-2.13-2.12-1.41 1.41L10.59 14l-2.12 2.12 1.41 1.41L12 15.41l2.12 2.12 1.41-1.41L13.41 14l2.12-2.12zM15.5 4l-1-1h-5l-1 1H5v2h14V4zM6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM8 9h8v10H8V9z"></path>
                </svg>
              </div>
            )}
          </button>
        ))}
      </div>
    </div>
  );
};

export default InputBar;
