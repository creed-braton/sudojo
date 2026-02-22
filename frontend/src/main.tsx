import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { SocketProvider } from "./providers/socket.tsx";
import { WS_URL } from "./api/config.ts";
import { NotesProvider } from "./providers/notes.tsx";
import { BrowserRouter } from "react-router-dom";
import { BoardProvider } from "./providers/board.tsx";
import { InputProvider } from "./providers/input.tsx";
import { KeyboardProvider } from "./providers/keyboard.tsx";
import { AnalysisProvider } from "./providers/analysis.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <SocketProvider url={WS_URL}>
        <AnalysisProvider>
          <NotesProvider>
            <BoardProvider>
              <InputProvider>
                <KeyboardProvider>
                  <App />
                </KeyboardProvider>
              </InputProvider>
            </BoardProvider>
          </NotesProvider>
        </AnalysisProvider>
      </SocketProvider>
    </BrowserRouter>
  </StrictMode>,
);
