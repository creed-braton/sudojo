import type { ReactElement } from "react";
import Button from "./Button";
import style from "./Input.module.css";

const Input = (): ReactElement => {
  return (
    <div className={style.input}>
      <Button>1</Button>
    </div>
  );
};

export default Input;
