import type { ReactElement, ReactNode } from "react";
import style from "./Button.module.css";

type ButtonProps = {
  children: ReactNode;
  onClick?: () => void;
};

const Button = ({
  children,
  onClick = () => {},
}: ButtonProps): ReactElement => {
  return <button className={style.button}>{children}</button>;
};

export default Button;
