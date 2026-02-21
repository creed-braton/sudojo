import type { CSSProperties, ReactElement } from "react";

const NotesIcon = ({
  className = undefined,
  style = undefined,
}: {
  className?: string;
  style?: CSSProperties;
}): ReactElement => {
  return (
    <div className={className} style={style}>
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
        <path d="m15 5 4 4" />
      </svg>
    </div>
  );
};

export default NotesIcon;
