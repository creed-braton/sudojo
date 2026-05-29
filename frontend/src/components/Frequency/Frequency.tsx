import {
  useEffect,
  useRef,
  useState,
  type ReactElement,
  type RefObject,
} from "react";
import style from "./Frequency.module.css";

type FrequencyProps = {
  frequency: number[];
};

const PADDING = 6;

const Frequency = ({ frequency }: FrequencyProps): ReactElement => {
  const svgRef: RefObject<SVGSVGElement | null> = useRef(null);
  const [{ w, h }, setSize] = useState<{ w: number; h: number }>({
    w: 200,
    h: 150,
  });

  useEffect((): (() => void) | void => {
    const svg = svgRef.current;
    if (svg === null) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        setSize({ w: width, h: height });
      }
    });

    observer.observe(svg);
    return () => observer.disconnect();
  }, []);

  const max = Math.max(...frequency);
  const innerW = w - PADDING * 2;
  const innerH = h - PADDING * 2;
  const baselineY = PADDING + innerH;

  const points = frequency.map((value, i) => {
    const x =
      frequency.length === 1
        ? PADDING + innerW / 2
        : PADDING + (i / (frequency.length - 1)) * innerW;
    const y =
      max === 0
        ? baselineY
        : PADDING + innerH - (value / max) * innerH;
    return `${x},${y}`;
  });

  const linePoints = points.join(" ");
  const areaPoints = `${PADDING},${baselineY} ${linePoints} ${PADDING + innerW},${baselineY}`;

  return (
    <div className={style.frequency}>
      <svg ref={svgRef} className={style.svg}>
        <polygon
          points={areaPoints}
          fill="rgba(57, 211, 83, 0.08)"
        />
        <polyline
          points={linePoints}
          fill="none"
          stroke="rgba(57, 211, 83, 0.6)"
          strokeWidth={1}
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>
    </div>
  );
};

export default Frequency;
