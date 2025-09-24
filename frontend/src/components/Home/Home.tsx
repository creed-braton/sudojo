import type { ReactElement } from "react";
import style from "./Home.module.css";

const Home = ({ createLobby }: { createLobby: () => void }): ReactElement => {
  return (
    <div className={style.home}>
      <div className={`glassmorphism ${style.container}`}>
        <h1 className={style.headline}>Welcome</h1>
        <h3 className={style.subtext}>have a seat, play a game!</h3>
        <button className={style.button} onClick={createLobby}>
          Play
        </button>
      </div>
    </div>
  );
};

export default Home;
