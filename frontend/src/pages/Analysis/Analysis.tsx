import { type ReactElement } from "react";
import {
  useAnalysis,
  type AnalysisContextProps,
} from "../../providers/analysis";
import Board from "../../components/Board/Board";
import Plot from "../../components/Plot/Plot";
import PlayerList from "../../components/PlayerList/PlayerList";
import style from "./Analysis.module.css";

const Analysis = (): ReactElement => {
  const analysis: AnalysisContextProps = useAnalysis();

  return (
    <div className={style.analysis}>
      <div className={style.stack}>
        {analysis.board !== null && (
          <div className={style.board}>
            <Board board={analysis.board} cursor={null} setCursor={() => {}} />
          </div>
        )}
        <div className={style.sidebar}>
          <div className={style.players}>
            <PlayerList players={analysis.players} maxPlayers={analysis.maxPlayers} />
          </div>
          {analysis.series !== null && (
            <div className={style.chart}>
              <Plot series={analysis.series} />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Analysis;
