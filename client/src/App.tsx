import { useEffect, type ReactElement } from "react";
import {
  Route,
  Routes,
  useLocation,
  useNavigate,
  type NavigateFunction,
} from "react-router-dom";
import { postLobby } from "./api/api";
import Lobby from "./pages/Lobby/Lobby";

const App = (): ReactElement => {
  const navigate: NavigateFunction = useNavigate();
  const location = useLocation();

  useEffect((): void => {
    const id: string[] = location.pathname.split("/");
    if (id.length === 3) return;

    postLobby(8, true, true, true, "joker").then(
      (id: string): void => void navigate(`/l/${id}`),
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
        <Route path="/l/:id" element={<Lobby />} />
      </Routes>
    </div>
  );
};

export default App;
