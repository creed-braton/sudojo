import { useState } from "react";
import { HTTP_URL } from "../config";

class ApiError extends Error {
  status: number;
  message: string;
  constructor(status: number, message: string) {
    super();
    this.status = status;
    this.message = message;
  }
}

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
    postLobby()
      .then((id: string) => setId(id))
      .catch((error: Error) => console.log(error));
  };

  const join = async (id: string, name: string): Promise<string> => {
    try {
      const token: string = await patchLobby(id, name);
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
