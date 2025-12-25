import { useState, type ReactElement } from "react";
import style from "./Home.module.css";
import Button from "../Button/Button";
import type { LobbyConfig } from "../../hooks/lobby";

const DIFFICULTIES = [
  { value: "easy", label: "Easy" },
  { value: "medium", label: "Medium" },
  { value: "hard", label: "Hard" },
  { value: "extreme", label: "Extreme" },
  { value: "joker", label: "Joker" },
];

const Home = ({
  createLobby,
}: {
  createLobby: (config: LobbyConfig) => void;
}): ReactElement => {
  const [maxPlayers, setMaxPlayers] = useState(6);
  const [strict, setStrict] = useState(true);
  const [difficulty, setDifficulty] = useState("medium");

  const handleCreate = () => {
    createLobby({ maxPlayers, strict, difficulty });
  };

  return (
    <div className={style.home}>
      <div className={`glass-container ${style.container}`}>
        <h1 className={style.headline}>Welcome</h1>
        <h3 className={style.subtext}>have a seat, play a game!</h3>

        <div className={style.config}>
          <div className={style.configSection}>
            <label className={style.label}>
              Players: <span className={style.value}>{maxPlayers}</span>
            </label>
            <input
              type="range"
              min={1}
              max={8}
              value={maxPlayers}
              onChange={(e) => setMaxPlayers(Number(e.target.value))}
              className={style.slider}
            />
            <div className={style.sliderLabels}>
              <span>1</span>
              <span>8</span>
            </div>
          </div>

          <div className={style.configSection}>
            <label className={style.label}>Difficulty</label>
            <div className={style.difficultyGrid}>
              {DIFFICULTIES.map((d) => (
                <button
                  key={d.value}
                  className={`${style.difficultyOption} ${difficulty === d.value ? style.selected : ""}`}
                  onClick={() => setDifficulty(d.value)}
                >
                  {d.label}
                </button>
              ))}
            </div>
          </div>

          <div className={style.configSection}>
            <label className={style.label}>Game Mode</label>
            <div className={style.toggleContainer}>
              <button
                className={`${style.toggleOption} ${strict ? style.selected : ""}`}
                onClick={() => setStrict(true)}
              >
                Strict
              </button>
              <button
                className={`${style.toggleOption} ${!strict ? style.selected : ""}`}
                onClick={() => setStrict(false)}
              >
                Relaxed
              </button>
            </div>
          </div>
        </div>

        <Button onClick={handleCreate} label={"Create Lobby"} />
      </div>
    </div>
  );
};

export default Home;
