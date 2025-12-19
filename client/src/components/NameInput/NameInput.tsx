import type { ReactElement } from "react";
import { useState } from "react";
import style from "./NameInput.module.css";
import Button from "../Button/Button";

const NameInput = ({
  lobbyId,
  joinLobby,
  setToken,
  close,
}: {
  lobbyId: string;
  joinLobby: (id: string, name: string) => Promise<string>;
  setToken: (state: string) => void;
  close: () => void;
}): ReactElement => {
  const [name, setName] = useState("");

  const validName = (value: string): boolean => {
    const validPattern = /^[a-zA-Z0-9_-]*$/;
    return validPattern.test(value) && value.length <= 12;
  };

  const handleNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value: string = event.target.value;
    validName(value) && setName(value);
  };

  const handleJoin = () => {
    if (!validName(name)) return;

    joinLobby(lobbyId, name.trim()).then((token: string) => setToken(token));
    close();
  };

  return (
    <div className={style.nameInput}>
      <div className={`glass-container ${style.container}`}>
        <div className={style.inputContainer}>
          <input
            type="text"
            value={name}
            onChange={handleNameChange}
            placeholder="Display name"
            className={style.input}
            maxLength={12}
            autoFocus
          />
        </div>
        <div className={style.buttonContainer}>
          <Button onClick={handleJoin} label="Join" />
        </div>
      </div>
    </div>
  );
};

export default NameInput;
