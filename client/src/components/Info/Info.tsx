import type { ReactElement } from "react";
import type { Player } from "../../api/types";
import Dropdown from "../Dropdown/Dropdown";
import Button from "../Button/Button";
import ShareIcon from "../../icons/Share";
import CheckIcon from "../../icons/Check";
import style from "./Info.module.css";

type InfoProps = {
  players: Player[];
  maxPlayers: number;
  onShare: () => void;
  copied: boolean;
};

const Info = ({
  players,
  maxPlayers,
  onShare,
  copied,
}: InfoProps): ReactElement => {
  const activePlayers = players.filter((player) => player.active).length;

  return (
    <div className={style.info}>
      <Dropdown
        label={
          <>
            Players{" "}
            <span className={style.count}>
              {activePlayers}/{maxPlayers}
            </span>
          </>
        }
      >
        {players.map((player) => (
          <li key={player.name} className={style.player}>
            <span
              className={`${style.dot} ${player.active ? style.active : style.inactive}`}
            />
            <span>{player.name || "<anonym>"}</span>
          </li>
        ))}
      </Dropdown>
      <div className={style.shareButton}>
        <span
          className={`${style.tooltip} ${copied ? style.tooltipVisible : ""}`}
        >
          Copied
        </span>
        <Button onClick={onShare} title="Share game link">
          {copied ? (
            <CheckIcon className={style.copied} />
          ) : (
            <ShareIcon className={style.share} />
          )}
        </Button>
      </div>
    </div>
  );
};

export default Info;
