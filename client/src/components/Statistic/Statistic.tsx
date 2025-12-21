import { useEffect, useState, type ReactElement } from "react";
import { getLobby } from "../../api/api";
import type { Lobby } from "../../api/types";

const Statistic = (): ReactElement => {
  const [id, setId] = useState<string>("");
  const [lobby, setLobby] = useState<Lobby | undefined>(undefined);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    setId(id);
  }, [location.pathname]);

  useEffect((): void => {
    id.length > 0 && getLobby(id).then((lobby: Lobby) => setLobby(lobby));
  }, [id]);

  return <>{lobby && lobby.started_at}</>;
};

export default Statistic;
