import type { ReactElement } from "react";
import style from "./Input.module.css";
import type { InputMode } from "../../providers/input";
import PingIcon from "../../icons/Ping";
import CrossIcon from "../../icons/Cross";
import NotesIcon from "../../icons/Notes";

type InputProps = {
  mode: InputMode;
  togglePing: () => void;
  toggleNotes: () => void;
  input: (value: number) => void;
};

const Input = ({
  mode,
  togglePing,
  toggleNotes,
  input,
}: InputProps): ReactElement => {
  return (
    <div className={style.input}>
      <div className={style.row}>
        {[1, 2, 3, 4, 5].map((num: number) => (
          <button key={num} className={style.button} onClick={() => {}}>
            {num}
          </button>
        ))}
        <button
          className={style.button}
          aria-selected={mode === "ping"}
          onClick={togglePing}
          title={
            mode === "ping" ? "Switch to insert mode" : "Switch to ping mode"
          }
        >
          <PingIcon className={style.icon} />
        </button>
      </div>
      <div className={style.row}>
        {[6, 7, 8, 9, 0].map((num: number) => (
          <button
            key={num}
            className={style.button}
            onClick={() => {
              input(num);
            }}
          >
            {num !== 0 ? num : <CrossIcon className={style.icon} />}
          </button>
        ))}
        <button
          className={style.button}
          aria-selected={mode === "notes"}
          onClick={toggleNotes}
          title={
            mode === "notes" ? "Switch to insert mode" : "Switch to notes mode"
          }
        >
          <NotesIcon className={style.icon} />
        </button>
      </div>
    </div>
  );
};

export default Input;
