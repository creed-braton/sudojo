import { useEffect, type ReactElement } from "react";
import {
  Route,
  Routes,
  useLocation,
  useNavigate,
  type NavigateFunction,
} from "react-router-dom";
import { postLobby, postPlayer } from "./api/api";
import Game from "./pages/Game/Game";

const App = (): ReactElement => {
  const navigate: NavigateFunction = useNavigate();
  const location = useLocation();

  useEffect((): void => {
    const id = location.pathname.split("/");
    if (id.length === 3) return;

    postLobby(6, true, true, true, "joker").then(
      (id: string): void =>
        void postPlayer(id).then((): void => void navigate(`/l/${id}`)),
    );
  }, []);

  return (
    <div
      style={{
        width: "100%",
        height: "100%",
      }}
    >
      <Routes>
        <Route path="/l/:id" element={<Game />} />
      </Routes>
    </div>
  );
};

export default App;
