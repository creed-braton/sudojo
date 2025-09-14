import type { ReactElement, ButtonHTMLAttributes } from "react";
import styles from "./styles.module.css";

type ButtonProps = {
  children: React.ReactNode;
  variant?: "primary" | "secondary";
} & ButtonHTMLAttributes<HTMLButtonElement>;

const Button = ({
  children,
  variant = "primary",
  className = "",
  ...props
}: ButtonProps): ReactElement => {
  return (
    <button
      className={`${styles.button} ${styles[variant]} glassmorphism ${className}`}
      {...props}
    >
      {children}
    </button>
  );
};

export default Button;
