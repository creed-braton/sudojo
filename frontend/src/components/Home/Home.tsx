import type { ReactElement } from "react";
import style from "./Home.module.css";
import Button from "../Button/Button";

const Home = ({ createLobby }: { createLobby: () => void }): ReactElement => {
  return (
    <div className={style.home}>
      <div className={`glassmorphism ${style.container}`}>
        <h1 className={style.headline}>Welcome</h1>
        <h3 className={style.subtext}>have a seat, play a game!</h3>
        <Button onClick={createLobby} label={"Play"} />
      </div>
    </div>
  );
};

export default Home;
