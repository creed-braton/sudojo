import { useEffect, useState, useRef, type ReactElement } from "react";
import * as d3 from "d3";
import style from "./Statistic.module.css";
import type { StatsProps } from "../../hooks/useStats";
import useStats from "../../hooks/useStats";
import { useLocation, type Location } from "react-router-dom";
import type { Points } from "../../types";

const Statistic = (): ReactElement => {
  const [id, setId] = useState<string | undefined>(undefined);
  const location: Location = useLocation();
  const stats: StatsProps = useStats(id);
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect((): void => {
    setId(location.pathname.split("/")[2]);
  }, [location.pathname]);

  useEffect((): void => {
    if (!stats.points.length || !svgRef.current) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove(); // Clear previous chart

    const margin = { top: 20, right: 30, bottom: 60, left: 60 };
    const width = 700 - margin.left - margin.right;
    const height = 350 - margin.top - margin.bottom;

    const g = svg
      .append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    // Scales
    const xScale = d3
      .scaleBand()
      .domain(stats.points.map((d: Points) => d.player))
      .range([0, width])
      .padding(0.2);

    const yScale = d3
      .scaleLinear()
      .domain([0, d3.max(stats.points, (d: Points) => d.points) || 0])
      .nice()
      .range([height, 0]);

    // Color scale
    const colorScale = d3
      .scaleOrdinal()
      .domain(stats.points.map((d: Points) => d.player))
      .range(d3.schemeCategory10);

    // Bars
    g.selectAll(".bar")
      .data(stats.points)
      .enter()
      .append("rect")
      .attr("class", "bar")
      .attr("x", (d: Points) => xScale(d.player) || 0)
      .attr("width", xScale.bandwidth())
      .attr("y", (d: Points) => yScale(d.points))
      .attr("height", (d: Points) => height - yScale(d.points))
      .attr("fill", (d: Points) => colorScale(d.player) as string)
      .attr("opacity", 0.8)
      .attr("rx", 4)
      .attr("ry", 4);

    // Add value labels on top of bars
    g.selectAll(".label")
      .data(stats.points)
      .enter()
      .append("text")
      .attr("class", "label")
      .attr(
        "x",
        (d: Points) => (xScale(d.player) || 0) + xScale.bandwidth() / 2
      )
      .attr("y", (d: Points) => yScale(d.points) - 5)
      .attr("text-anchor", "middle")
      .attr("fill", "#fff")
      .attr("font-size", "14px")
      .attr("font-weight", "bold")
      .text((d: Points) => d.points);

    // X-axis
    g.append("g")
      .attr("transform", `translate(0,${height})`)
      .call(d3.axisBottom(xScale))
      .selectAll("text")
      .attr("fill", "#fff")
      .attr("font-size", "12px")
      .style("text-anchor", "middle");

    // Y-axis
    g.append("g")
      .call(d3.axisLeft(yScale))
      .selectAll("text")
      .attr("fill", "#fff")
      .attr("font-size", "12px");

    // Style axis lines
    g.selectAll(".domain").attr("stroke", "rgba(255, 255, 255, 0.3)");

    g.selectAll(".tick line").attr("stroke", "rgba(255, 255, 255, 0.2)");

    // X-axis label
    g.append("text")
      .attr("transform", `translate(${width / 2}, ${height + 50})`)
      .style("text-anchor", "middle")
      .attr("fill", "#fff")
      .attr("font-size", "14px")
      .attr("font-weight", "500")
      .text("Players");

    // Y-axis label
    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 0 - margin.left)
      .attr("x", 0 - height / 2)
      .attr("dy", "1em")
      .style("text-anchor", "middle")
      .attr("fill", "#fff")
      .attr("font-size", "14px")
      .attr("font-weight", "500")
      .text("Points");
  }, [stats.points]);

  if (stats.loading) {
    return (
      <div className={style.statistic}>
        <div className={style.container}>
          <p style={{ color: "#fff", fontSize: "1.2rem" }}>
            Loading Statistics...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={style.statistic}>
      <div className={style.container}>
        <div className={style.chartContainer}>
          {stats.points.length > 0 ? (
            <svg
              ref={svgRef}
              width="700"
              height="350"
              style={{ overflow: "visible" }}
            />
          ) : (
            <p style={{ color: "#fff", fontSize: "1.2rem" }}>
              No statistics available
            </p>
          )}
        </div>
      </div>
    </div>
  );
};

export default Statistic;
