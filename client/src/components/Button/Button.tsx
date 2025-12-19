import type { ReactElement } from "react";
import style from "./Button.module.css";

const Button = ({
  onClick,
  label,
}: {
  onClick: () => void;
  label: string;
}): ReactElement => {
  return (
    <button className={style.button} onClick={onClick}>
      {label}
    </button>
  );
};

export default Button;
