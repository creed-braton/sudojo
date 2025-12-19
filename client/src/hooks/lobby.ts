import { useState } from "react";
import { postLobby, postPlayer } from "../api/api";

export type LobbyProps = {
  id: string | null;
  setId: (state: string | null) => void;
  getToken: (id: string) => string | undefined;
  create: () => void;
  join: (id: string, name: string) => Promise<string>;
};

const useLobby = (): LobbyProps => {
  const [id, setId] = useState<string | null>(null);

  const getToken = (id: string): string | undefined => {
    return localStorage.getItem(id) || undefined;
  };

  const setToken = (id: string, token: string): void => {
    localStorage.setItem(id, token);
  };

  const create = (): void => {
    postLobby(8, false)
      .then((id: string) => setId(id))
      .catch((error: Error) => console.log(error));
  };

  const join = async (id: string, name: string): Promise<string> => {
    try {
      const token: string = await postPlayer(id, name);
      setToken(id, token);
      return token;
    } catch (error) {
      console.error(error);
      throw error;
    }
  };

  return { id, setId, getToken, create, join };
};

export default useLobby;
