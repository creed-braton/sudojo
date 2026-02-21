import type { ReactElement, ReactNode } from "react";
import style from "./Button.module.css";

type ButtonProps = {
  children: ReactNode;
  onClick: () => void;
  selected?: boolean;
  title?: string | undefined;
};

const Button = ({
  children,
  onClick = () => {},
  selected = false,
  title = undefined,
}: ButtonProps): ReactElement => {
  return (
    <button
      className={style.button}
      onClick={onClick}
      aria-selected={selected}
      title={title}
    >
      {children}
    </button>
  );
};

export default Button;
