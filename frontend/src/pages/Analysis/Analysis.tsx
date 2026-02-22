import { type ReactElement } from "react";
import {
  useAnalysis,
  type AnalysisContextProps,
} from "../../providers/analysis";
import style from "./Analysis.module.css";
import Board from "../../components/Board/Board";

const Analysis = (): ReactElement => {
  const analysis: AnalysisContextProps = useAnalysis();

  return (
    <div>
      {analysis.board !== null && (
        <Board board={analysis.board} cursor={null} setCursor={() => {}} />
      )}
    </div>
  );
};

export default Analysis;
