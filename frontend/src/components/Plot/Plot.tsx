import {
  useEffect,
  useRef,
  useState,
  type ReactElement,
  type RefObject,
} from "react";
import type { Series } from "../../providers/analysis";
import style from "./Plot.module.css";

type PlotProps = {
  series: Series;
};

const PAD = { top: 16, right: 16, bottom: 36, left: 44 };
const PAD_COMPACT = { top: 8, right: 8, bottom: 28, left: 36 };

// Hysteresis thresholds derived from the 450px border-box compact breakpoint:
//   enter compact  when svgW < 450 - 2×(non-compact padding 16px) = 418
//   exit  compact  when svgW ≥ 450 - 2×(compact padding     8px)  = 434
// The 16px dead-zone equals 2×ΔcssP padding so a flip can never immediately
// push the SVG width back across the threshold that triggered it.
const COMPACT_ENTER = 418;
const COMPACT_EXIT = 434;

const formatTime = (ns: number): string => {
  const totalSeconds = Math.round(ns / 1e9);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes === 0) return `${seconds}s`;
  if (seconds === 0) return `${minutes}m`;
  return `${minutes}m${seconds}s`;
};

const niceTickInterval = (maxVal: number, targetCount: number): number => {
  if (maxVal <= 0) return 1;
  const rawInterval = maxVal / targetCount;
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)));
  const normalized = rawInterval / magnitude;
  let nice: number;
  if (normalized < 1.5) nice = 1;
  else if (normalized < 3.5) nice = 2;
  else if (normalized < 7.5) nice = 5;
  else nice = 10;
  return nice * magnitude;
};

const Plot = ({ series }: PlotProps): ReactElement => {
  const svgRef: RefObject<SVGSVGElement | null> = useRef(null);
  const [borderRadius, setBorderRadius] = useState<number>(8);
  const [compact, setCompact] = useState<boolean>(false);
  const [{ w, h }, setSize] = useState<{ w: number; h: number }>({
    w: 400,
    h: 250,
  });

  useEffect((): (() => void) | void => {
    const svg = svgRef.current;
    if (svg === null) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        setSize({ w: width, h: height });
        setBorderRadius(Math.min(width, height) * 0.03);
        setCompact((prev) => {
          if (!prev && width < COMPACT_ENTER) return true;
          if (prev && width >= COMPACT_EXIT) return false;
          return prev;
        });
      }
    });

    observer.observe(svg);
    return () => observer.disconnect();
  }, []);

  const pad = compact ? PAD_COMPACT : PAD;
  const labelFontSize = Math.round(Math.max(8, Math.min(14, Math.min(w, h) * 0.045)));
  const innerW = w - pad.left - pad.right;
  const innerH = h - pad.top - pad.bottom;

  const maxX = series.xValues[series.xValues.length - 1] ?? 0;
  const maxY = Math.max(
    1,
    ...Array.from(series.yValues.values()).flatMap((ys) => ys),
  );

  const toSvgX = (x: number): number =>
    pad.left + (maxX > 0 ? (x / maxX) * innerW : 0);
  const toSvgY = (y: number): number =>
    pad.top + innerH - (y / maxY) * innerH;

  const yTickInterval = niceTickInterval(maxY, 5);
  const yTicks: number[] = [];
  for (
    let v = 0;
    v <= Math.ceil(maxY / yTickInterval) * yTickInterval;
    v += yTickInterval
  ) {
    yTicks.push(v);
  }

  const xTickInterval = niceTickInterval(maxX, 5);
  const xTicks: number[] = [];
  for (let v = 0; v <= maxX; v += xTickInterval) {
    xTicks.push(v);
  }

  return (
    <div className={style.plot} data-compact={compact} style={{ borderRadius }}>
      <svg ref={svgRef} className={style.svg}>
        {yTicks.map((v) => (
          <line
            key={v}
            className={style.gridLine}
            x1={pad.left}
            x2={pad.left + innerW}
            y1={toSvgY(v)}
            y2={toSvgY(v)}
          />
        ))}
        <line
          className={style.axis}
          x1={pad.left}
          x2={pad.left}
          y1={pad.top}
          y2={pad.top + innerH}
        />
        <line
          className={style.axis}
          x1={pad.left}
          x2={pad.left + innerW}
          y1={pad.top + innerH}
          y2={pad.top + innerH}
        />
        {yTicks.map((v) => (
          <text
            key={v}
            className={style.label}
            fontSize={labelFontSize}
            x={pad.left - 6}
            y={toSvgY(v)}
            textAnchor="end"
            dominantBaseline="middle"
          >
            {Math.round(v)}
          </text>
        ))}
        {xTicks.map((v) => (
          <text
            key={v}
            className={style.label}
            fontSize={labelFontSize}
            x={toSvgX(v)}
            y={pad.top + innerH + 20}
            textAnchor="middle"
            dominantBaseline="middle"
          >
            {formatTime(v)}
          </text>
        ))}
        {Array.from(series.yValues.entries()).map(([color, ys]) => {
          const points = series.xValues
            .map((x, i) => `${toSvgX(x)},${toSvgY(ys[i] ?? 0)}`)
            .join(" ");
          return (
            <polyline
              key={color}
              className={style.line}
              points={points}
              stroke={color}
            />
          );
        })}
      </svg>
    </div>
  );
};

export default Plot;
