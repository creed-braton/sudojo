import { useEffect, useState } from "react";
import { HTTP_URL } from "../config";
import { ApiError, type GameStats, type Points, type Score } from "../types";

const getStats = async (id: string): Promise<GameStats> => {
  const response: Response = await fetch(`${HTTP_URL}/lobbies/${id}/stats`, {
    method: "GET",
  });

  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }

  return response.json();
};

export type StatsProps = {
  points: Points[];
  loading: boolean;
};

const useStats = (id: string | undefined): StatsProps => {
  const [data, setData] = useState<GameStats | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [points, setPoints] = useState<Points[]>([]);

  useEffect((): void => {
    if (!id) {
      setData(null);
      return;
    }
    setLoading(true);
    getStats(id)
      .then((data: GameStats) => setData(data))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect((): void => {
    data
      ? setPoints(
          data.scores.map((score: Score): Points => {
            return {
              player: score.player_name,
              points: score.points.length,
            } as Points;
          }),
        )
      : setPoints([]);
  }, [data]);

  return { points, loading };
};

export default useStats;
