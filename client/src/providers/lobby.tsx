import { createContext, useContext, useState, type ReactNode } from "react";

export type LobbyContextProps = {};

const LobbyContext = createContext<LobbyContextProps | null>(null);

type SocketProviderProps = {
  children: ReactNode;
};

export const LobbyProvider = ({ children }: SocketProviderProps) => {
  const [id, setId] = useState<string>("");

  const value: LobbyContextProps = {
    setId,
  };

  return (
    <LobbyContext.Provider value={value}>{children}</LobbyContext.Provider>
  );
};

export const useLobby = (): LobbyContextProps => {
  const context = useContext(LobbyContext);
  if (!context) {
    throw new Error("useLobby must be used within a LobbyProvider");
  }
  return context;
};
