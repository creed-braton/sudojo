import { HTTP_URL } from "./config";

class ApiError extends Error {
  status: number;
  message: string;
  constructor(status: number, message: string) {
    super();
    this.status = status;
    this.message = message;
  }
}

const postLobby = async (
  maxPlayer: number,
  strict: boolean,
): Promise<string> => {
  const response: Response = await fetch(HTTP_URL + "/lobbies", {
    method: "POST",
    body: JSON.stringify({
      max_player: maxPlayer,
      strict: strict,
    }),
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.text();
};

const postPlayer = async (id: string, name?: string): Promise<string> => {
  const url = new URL(`${HTTP_URL}/lobbies/${id}`);
  if (name) url.searchParams.append("name", name);

  const response: Response = await fetch(url.toString(), { method: "PATCH" });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.text();
};
