import type { CSSProperties, ReactElement } from "react";

const PingIcon = ({
  className = undefined,
  style = undefined,
}: {
  className?: string;
  style?: CSSProperties;
}): ReactElement => {
  return (
    <div className={className} style={style}>
      <svg
        stroke="currentColor"
        fill="none"
        strokeWidth="2"
        viewBox="0 0 24 24"
        strokeLinecap="round"
        strokeLinejoin="round"
        height="1em"
        width="1em"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path d="M12 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
        <path d="M12 5l0 -2"></path>
        <path d="M17 7l1.4 -1.4"></path>
        <path d="M19 12l2 0"></path>
        <path d="M17 17l1.4 1.4"></path>
        <path d="M12 19l0 2"></path>
        <path d="M7 17l-1.4 1.4"></path>
        <path d="M6 12l-2 0"></path>
        <path d="M7 7l-1.4 -1.4"></path>
      </svg>
    </div>
  );
};

export default PingIcon;
