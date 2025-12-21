import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import type { Insertion } from "../Board/Board";
import styles from "./Plot.module.css";

type PlayerData = {
  name: string;
  color: string;
};

type Props = {
  insertions: Insertion[];
  players: PlayerData[];
  startTimestamp: number;
};

const Plot = ({ insertions, players, startTimestamp }: Props) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });

  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect;
        setDimensions({ width, height });
      }
    });

    resizeObserver.observe(svg);
    return () => resizeObserver.disconnect();
  }, []);

  useEffect(() => {
    if (!svgRef.current || insertions.length === 0 || players.length === 0) {
      return;
    }

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const margin = { top: 20, right: 20, bottom: 40, left: 50 };
    const width = svgRef.current.clientWidth - margin.left - margin.right;
    const height = svgRef.current.clientHeight - margin.top - margin.bottom;

    const g = svg
      .append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    // Sort insertions by timestamp
    const sortedInsertions = [...insertions].sort(
      (a, b) => a.timestamp - b.timestamp,
    );

    // Calculate time range (relative to start)
    const lastTimestamp =
      sortedInsertions[sortedInsertions.length - 1]?.timestamp ??
      startTimestamp;

    // X scale: time from start (0) to last insertion
    const xScale = d3
      .scaleLinear()
      .domain([0, lastTimestamp - startTimestamp])
      .range([0, width]);

    // Find max count for Y scale
    const playerCounts = new Map<string, number>();
    for (const player of players) {
      playerCounts.set(player.name, 0);
    }
    for (const ins of sortedInsertions) {
      const current = playerCounts.get(ins.playerName) ?? 0;
      playerCounts.set(ins.playerName, current + 1);
    }
    const maxCount = Math.max(...playerCounts.values(), 1);

    // Y scale: correct inserts count
    const yScale = d3.scaleLinear().domain([0, maxCount]).range([height, 0]);

    // Add grid lines
    const xGridlines = d3
      .axisBottom(xScale)
      .ticks(6)
      .tickSize(-height)
      .tickFormat(() => "");

    g.append("g")
      .attr("class", styles.grid)
      .attr("transform", `translate(0,${height})`)
      .call(xGridlines);

    const yGridlines = d3
      .axisLeft(yScale)
      .ticks(Math.min(maxCount, 10))
      .tickSize(-width)
      .tickFormat(() => "");

    g.append("g").attr("class", styles.grid).call(yGridlines);

    // Build line data for each player
    const playerLines = new Map<
      string,
      { time: number; count: number; color: string }[]
    >();

    for (const player of players) {
      // Start at (0, 0) for each player
      playerLines.set(player.name, [
        { time: 0, count: 0, color: player.color },
      ]);
    }

    // Track running counts
    const runningCounts = new Map<string, number>();
    for (const player of players) {
      runningCounts.set(player.name, 0);
    }

    // Add data points for each insertion
    for (const ins of sortedInsertions) {
      const currentCount = (runningCounts.get(ins.playerName) ?? 0) + 1;
      runningCounts.set(ins.playerName, currentCount);

      const player = players.find((p) => p.name === ins.playerName);
      if (player) {
        const lineData = playerLines.get(ins.playerName);
        if (lineData) {
          lineData.push({
            time: ins.timestamp - startTimestamp,
            count: currentCount,
            color: player.color,
          });
        }
      }
    }

    // Create line generator
    const line = d3
      .line<{ time: number; count: number }>()
      .x((d) => xScale(d.time))
      .y((d) => yScale(d.count))
      .curve(d3.curveLinear);

    // Draw lines for each player
    const lineGroups: {
      visible: d3.Selection<
        SVGPathElement,
        { time: number; count: number; color: string }[],
        null,
        undefined
      >;
      hitArea: d3.Selection<
        SVGPathElement,
        { time: number; count: number; color: string }[],
        null,
        undefined
      >;
    }[] = [];

    for (const player of players) {
      const lineData = playerLines.get(player.name);
      if (lineData && lineData.length > 1) {
        // Invisible hit area with larger stroke width
        const hitArea = g
          .append("path")
          .datum(lineData)
          .attr("fill", "none")
          .attr("stroke", "transparent")
          .attr("stroke-width", 15)
          .attr("d", line)
          .attr("class", styles.line);

        // Visible line
        const visible = g
          .append("path")
          .datum(lineData)
          .attr("fill", "none")
          .attr("stroke", player.color)
          .attr("stroke-width", 2)
          .attr("d", line)
          .style("pointer-events", "none")
          .style("transition", "opacity 0.2s ease");

        lineGroups.push({ visible, hitArea });
      }
    }

    // Create tooltip
    const tooltip = d3
      .select(svgRef.current.parentElement)
      .append("div")
      .attr("class", styles.tooltip)
      .style("opacity", 0);

    // Add hover interactions
    lineGroups.forEach((group, index) => {
      const lineData = Array.from(playerLines.values())[index];

      group.hitArea
        .on("mouseenter", () => {
          lineGroups.forEach((otherGroup, otherIndex) => {
            if (otherIndex !== index) {
              otherGroup.visible.style("opacity", 0.2);
            }
          });
          tooltip.style("opacity", 1).style("visibility", "visible");
        })
        .on("mousemove", (event) => {
          const [mouseX] = d3.pointer(event);
          const xValue = xScale.invert(mouseX);

          // Find closest data point
          let closestPoint = lineData[0];
          let minDist = Math.abs(lineData[0].time - xValue);

          for (const point of lineData) {
            const dist = Math.abs(point.time - xValue);
            if (dist < minDist) {
              minDist = dist;
              closestPoint = point;
            }
          }

          const svgRect = svgRef.current!.getBoundingClientRect();
          const tooltipX = event.clientX - svgRect.left + 10;
          const tooltipY = event.clientY - svgRect.top - 10;

          tooltip
            .html(`${Math.round(closestPoint.count)}`)
            .style("left", `${tooltipX}px`)
            .style("top", `${tooltipY}px`);
        })
        .on("mouseleave", () => {
          lineGroups.forEach((otherGroup) => {
            otherGroup.visible.style("opacity", 1);
          });
          tooltip.style("opacity", 0).style("visibility", "hidden");
        });
    });

    // Format time for axis labels
    const formatTime = (nanos: number): string => {
      const seconds = nanos / 1_000_000_000;
      if (seconds < 60) {
        return `${Math.floor(seconds)}s`;
      }
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = Math.floor(seconds % 60);
      if (minutes < 60) {
        return `${minutes}m ${remainingSeconds}s`;
      }
      const hours = Math.floor(minutes / 60);
      const remainingMinutes = minutes % 60;
      return `${hours}h ${remainingMinutes}m`;
    };

    // X axis
    const xDomain = lastTimestamp - startTimestamp;
    const isMobile = window.innerWidth < 800;
    const xAxis = d3
      .axisBottom(xScale)
      .tickFormat((d) => formatTime(d as number));
    if (isMobile) {
      xAxis.tickValues([0, xDomain / 2, xDomain]);
    } else {
      xAxis.ticks(6);
    }

    g.append("g")
      .attr("transform", `translate(0,${height})`)
      .call(xAxis)
      .attr("class", styles.axis);

    // Y axis - only show labels at 0, 25%, 50%, 75%, and 100%
    const quarterValues = [
      0,
      Math.round(maxCount * 0.25),
      Math.round(maxCount * 0.5),
      Math.round(maxCount * 0.75),
      maxCount,
    ];
    const yAxis = d3
      .axisLeft(yScale)
      .tickValues(quarterValues)
      .tickFormat(d3.format("d"));

    g.append("g").call(yAxis).attr("class", styles.axis);

    // X axis label
    g.append("text")
      .attr("x", width / 2)
      .attr("y", height + 35)
      .attr("text-anchor", "middle")
      .attr("class", styles.axisLabel)
      .text("Time");

    // Y axis label
    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("x", -height / 2)
      .attr("y", -40)
      .attr("text-anchor", "middle")
      .attr("class", styles.axisLabel)
      .text("Points");

    // Cleanup tooltip on unmount
    return () => {
      tooltip.remove();
    };
  }, [insertions, players, startTimestamp, dimensions]);

  return (
    <div className={styles.container}>
      <svg ref={svgRef} className={styles.svg} />
    </div>
  );
};

export default Plot;
