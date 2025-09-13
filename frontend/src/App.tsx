import { useState, type ReactElement } from "react";
import Home from "./components/Home/Home";
import CreateLobby from "./components/CreateLobby/CreateLobby";
import JoinLobby from "./components/JoinLobby/JoinLobby";
import SudokuGame from "./components/SudokuGame/SudokuGame";
import styles from "./App.module.css";

type View = "home" | "create" | "join" | "game";

const App = (): ReactElement => {
  const [currentView, setCurrentView] = useState<View>("home");
  const [currentLobbyId, setCurrentLobbyId] = useState<string>("");

  const handleNavigateToHome = (): void => {
    setCurrentView("home");
  };

  const handleNavigateToCreate = (): void => {
    setCurrentView("create");
  };

  const handleNavigateToJoin = (): void => {
    setCurrentView("join");
  };

  const handleNavigateToGame = (lobbyId: string): void => {
    setCurrentLobbyId(lobbyId);
    setCurrentView("game");
  };

  return (
    <div className={styles.container}>
      {currentView === "home" && (
        <Home
          onCreateLobby={handleNavigateToCreate}
          onJoinLobby={handleNavigateToJoin}
        />
      )}

      {currentView === "create" && (
        <CreateLobby
          onBack={handleNavigateToHome}
          onJoinSuccess={handleNavigateToGame}
        />
      )}

      {currentView === "join" && (
        <JoinLobby
          onBack={handleNavigateToHome}
          onJoinSuccess={handleNavigateToGame}
        />
      )}

      {currentView === "game" && (
        <SudokuGame onBack={handleNavigateToHome} lobbyId={currentLobbyId} />
      )}
    </div>
  );
};

export default App;
