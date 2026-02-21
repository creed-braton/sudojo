import type { CSSProperties, ReactElement } from "react";

const EnterIcon = ({
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
        fill="currentColor"
        strokeWidth="0"
        viewBox="0 0 512 512"
        height="1em"
        width="1em"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          fill="none"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="48"
          d="m268 112 144 144-144 144m124-144H100"
        ></path>
      </svg>
    </div>
  );
};

export default EnterIcon;
