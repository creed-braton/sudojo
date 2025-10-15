import { useState } from "react";
import { HTTP_URL } from "../config";
import { ApiError, type Player } from "../types";

const postLobby = async (): Promise<string> => {
  const response: Response = await fetch(HTTP_URL + "/lobbies", {
    method: "POST",
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.text();
};

const patchLobby = async (id: string, name?: string): Promise<string> => {
  const url = new URL(`${HTTP_URL}/lobbies/${id}`);
  if (name) url.searchParams.append("name", name);

  const response: Response = await fetch(url.toString(), { method: "PATCH" });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.text();
};

export type LobbyProps = {
  id: string | null;
  setId: (state: string | null) => void;
  getPlayer: (id: string) => Player | undefined;
  create: () => void;
  join: (id: string, name: string) => Promise<string>;
};

const useLobby = (): LobbyProps => {
  const [id, setId] = useState<string | null>(null);

  const getPlayer = (id: string): Player | undefined => {
    try {
      return JSON.parse(localStorage.getItem(id) || "") as Player;
    } catch {
      return undefined;
    }
  };

  const setPlayer = (id: string, player: Player): void => {
    localStorage.setItem(id, JSON.stringify(player));
  };

  const create = (): void => {
    postLobby()
      .then((id: string) => setId(id))
      .catch((error: Error) => console.log(error));
  };

  const join = async (id: string, name: string): Promise<string> => {
    try {
      const token: string = await patchLobby(id, name);
      setPlayer(id, { token, name } as Player);
      return token;
    } catch (error) {
      console.error(error);
      throw error;
    }
  };

  return { id, setId, getPlayer, create, join };
};

export default useLobby;
