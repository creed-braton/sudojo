import type { ReactElement } from "react";
import style from "./Button.module.css";

const Button = (): ReactElement => {
  return <button className={style.button}>1</button>;
};

export default Button;
