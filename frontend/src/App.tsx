import { useEffect, type ReactElement } from "react";
import useApi, { type ApiProps } from "./hooks/useApi";
import Background from "./components/Background/Background";
import { Route, Routes, useNavigate } from "react-router-dom";
import Home from "./components/Home/Home";
import Lobby from "./components/Lobby/Lobby";

const App = (): ReactElement => {
  const api: ApiProps = useApi();
  const navigate = useNavigate();

  useEffect((): void => {
    api.lobbyId && navigate(`/l/${api.lobbyId}`);
  }, [api.lobbyId]);

  return (
    <>
      <Routes>
        <Route path="/" element={<Home createLobby={api.createLobby} />} />
        <Route
          path="/l/:id"
          element={
            <Lobby
              joinLobby={api.joinLobby}
              sendMove={api.sendMove}
              initialState={api.initialState}
              currentState={api.currentState}
            />
          }
        />
      </Routes>
      <Background />
    </>
  );
};

export default App;
