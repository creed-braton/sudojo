import type { ReactElement } from "react";
import style from "./Player.module.css";

type PlayerProps = {
  color: string;
  name: string;
};

const Player = ({ color, name }: PlayerProps): ReactElement => {
  return (
    <div className={style.player}>
      <span
        className={style.dot}
        style={{
          backgroundColor: color,
          boxShadow: `0 0 4px ${color}99, 0 0 8px ${color}55`,
        }}
      />
      <span className={style.name}>{name}</span>
    </div>
  );
};

export default Player;
