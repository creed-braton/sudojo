import type { CSSProperties, ReactElement, ReactNode } from "react";
import ErrorIcon from "../../icons/Error";
import WarnIcon from "../../icons/Warn";
import style from "./Banner.module.css";

type Variant = "error" | "warning";

type BannerProps = {
  variant: Variant;
  children: ReactNode;
  width?: number;
  height?: number;
};

const Banner = ({ variant, children, width, height }: BannerProps): ReactElement => {
  const inlineStyle: CSSProperties = {
    width: width !== undefined ? `${width}px` : "100%",
    height: height !== undefined ? `${height}px` : "100%",
  };

  return (
    <aside className={`${style.banner} ${style[variant]}`} role="alert" style={inlineStyle}>
      {variant === "error" && <ErrorIcon className={style.icon} />}
      {variant === "warning" && <WarnIcon className={style.icon} />}
      <div className={style.divider} />
      <div className={style.content}>{children}</div>
    </aside>
  );
};

export default Banner;
