import { useEffect, type ReactElement } from "react";
import {
  Route,
  Routes,
  useNavigate,
  type NavigateFunction,
} from "react-router-dom";
import type { LobbyProps } from "./hooks/lobby";
import useLobby from "./hooks/lobby";
import Home from "./components/Home/Home";
import Lobby from "./components/Lobby/Lobby";

const App = (): ReactElement => {
  const lobby: LobbyProps = useLobby();
  const navigate: NavigateFunction = useNavigate();

  useEffect((): void => {
    lobby.id && navigate(`/l/${lobby.id}`);
  }, [lobby.id]);

  return (
    <>
      <Routes>
        <Route path="/" element={<Home createLobby={lobby.create} />} />
        <Route
          path="/l/:id"
          element={<Lobby joinLobby={lobby.join} getToken={lobby.getToken} />}
        />
      </Routes>
    </>
  );
};

export default App;
