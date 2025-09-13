import type { ReactElement, ReactNode } from "react";
import Background from "../Background/Background";
import styles from "./styles.module.css";

type LayoutProps = {
  children: ReactNode;
};

const Layout = ({ children }: LayoutProps): ReactElement => {
  return (
    <div className={styles.container}>
      <Background />
      <div className={styles.content}>{children}</div>
    </div>
  );
};

export default Layout;
