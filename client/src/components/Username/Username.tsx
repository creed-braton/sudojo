import type { ChangeEvent, ReactElement } from "react";
import Button from "../Button/Button";
import EnterIcon from "../../icons/Enter";
import style from "./Username.module.css";

const MAX_USERNAME_LENGTH = 16;
const VALID_USERNAME_PATTERN = /^[a-zA-Z0-9_-]*$/;

type UsernameProps = {
  username: string;
  setUsername: (username: string) => void;
  onClick: () => void;
};

const Username = ({
  username,
  setUsername,
  onClick,
}: UsernameProps): ReactElement => {
  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (value.length > MAX_USERNAME_LENGTH) {
      return;
    }
    if (!VALID_USERNAME_PATTERN.test(value)) {
      return;
    }
    setUsername(value);
  };

  return (
    <div className={style.username}>
      <input
        type="text"
        className={style.input}
        value={username}
        onChange={handleChange}
        placeholder="Enter username"
        maxLength={MAX_USERNAME_LENGTH}
      />
      <div className={style.button}>
        <Button onClick={onClick} title="Join game">
          <EnterIcon className={style.icon} />
        </Button>
      </div>
    </div>
  );
};

export default Username;
