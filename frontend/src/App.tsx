import { useEffect, type ReactElement } from "react";
import Background from "./components/Background/Background";
import { Route, Routes, useNavigate } from "react-router-dom";
import Home from "./components/Home/Home";
import type { LobbyProps } from "./hooks/useLobby";
import useLobby from "./hooks/useLobby";
import Lobby from "./components/Lobby/Lobby";

const App = (): ReactElement => {
  const lobby: LobbyProps = useLobby();
  const navigate = useNavigate();

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
      <Background />
    </>
  );
};

export default App;
