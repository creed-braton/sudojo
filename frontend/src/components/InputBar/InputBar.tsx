import type { ReactElement } from "react";
import NumericButton from "../NumericButton/NumericButton";
import styles from "./styles.module.css";

type InputBarProps = {
  onNumberClick?: (number: number) => void;
  onClearClick?: () => void;
};

const InputBar = ({
  onNumberClick = () => {},
  onClearClick = () => {},
}: InputBarProps): ReactElement => {
  // Create an array of numbers 1-9 for the Sudoku buttons
  const numbers = Array.from({ length: 9 }, (_, i) => i + 1);

  // Split numbers into two rows
  const firstRowNumbers = numbers.slice(0, 5); // Numbers 1-5
  const secondRowNumbers = numbers.slice(5); // Numbers 6-9

  return (
    <div className={`${styles.inputBar} glassmorphism`}>
      <div className={styles.rowsContainer}>
        <div className={styles.buttonRow}>
          {firstRowNumbers.map((number) => (
            <NumericButton
              key={number}
              value={number}
              onClick={() => onNumberClick(number)}
              aria-label={`Number ${number}`}
            />
          ))}
        </div>
        <div className={styles.buttonRow}>
          {secondRowNumbers.map((number) => (
            <NumericButton
              key={number}
              value={number}
              onClick={() => onNumberClick(number)}
              aria-label={`Number ${number}`}
            />
          ))}
          <NumericButton
            value="clear"
            onClick={onClearClick}
            aria-label="Clear"
          />
        </div>
      </div>
    </div>
  );
};

export default InputBar;
