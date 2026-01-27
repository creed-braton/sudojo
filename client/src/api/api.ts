import { HTTP_URL } from "./config";
import { isLobby, isUUID, type Lobby, type UUID } from "./types";

export class HttpError extends Error {
  status: number;
  message: string;
  constructor(status: number, message: string) {
    super();
    this.status = status;
    this.message = message;
  }
}

export const getLobby = async (id: string): Promise<Lobby> => {
  const response: Response = await fetch(HTTP_URL + `/lobbies/${id}`);

  if (!response.ok) {
    throw new HttpError(response.status, await response.text());
  }

  const data: unknown = await response.json();
  if (!isLobby(data)) {
    throw new Error("Invalid lobby data received from API");
  }

  return data;
};

export const postLobby = async (
  maxPlayer: number,
  strict: boolean,
  pings: boolean,
  notes: boolean,
  difficulty: string,
): Promise<UUID> => {
  const response: Response = await fetch(HTTP_URL + "/lobbies", {
    method: "POST",
    body: JSON.stringify({
      strict_mode: strict,
      pings_allowed: pings,
      notes_allowed: notes,
      max_player: maxPlayer,
      difficulty: difficulty,
    }),
  });

  if (!response.ok) {
    throw new HttpError(response.status, await response.text());
  }
  const data: unknown = await response.text();
  if (!isUUID(data)) {
    throw new Error("Invalid lobby UUID received from API");
  }

  return data;
};

export const postPlayer = async (
  id: string,
  name?: string,
): Promise<string> => {
  const url = new URL(`${HTTP_URL}/lobbies/${id}/players`);
  if (name) url.searchParams.append("name", name);

  const response: Response = await fetch(url.toString(), {
    method: "POST",
    credentials: "include",
  });

  if (!response.ok) {
    throw new HttpError(response.status, await response.text());
  }

  return response.text();
};

export const getPlayer = async (id: string): Promise<string> => {
  const url = new URL(`${HTTP_URL}/lobbies/${id}/players`);

  const response: Response = await fetch(url.toString(), {
    method: "GET",
    credentials: "include",
  });

  if (!response.ok) {
    throw new HttpError(response.status, await response.text());
  }

  return response.text();
};
