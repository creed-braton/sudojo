import { createContext, useEffect, type ReactNode } from "react";
import { useInput, type InputContextProps } from "./input";

export type KeyboardContextProps = {};

const KeyboardContext = createContext<null>(null);

type KeyboardProviderProps = {
  children: ReactNode;
};

export const KeyboardProvider = ({ children }: KeyboardProviderProps) => {
  const input: InputContextProps = useInput();

  useEffect((): (() => void) => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (/^[0-9]$/.test(event.key)) {
        const num: number = parseInt(event.key, 10);
        input.input(num);
      } else if (event.key === "n") {
        input.toggleNotes();
      } else if (event.key === "p") {
        input.togglePing();
      } else if (event.key === "h" || event.key === "ArrowLeft") {
        input.left();
      } else if (event.key === "j" || event.key === "ArrowDown") {
        input.down();
      } else if (event.key === "k" || event.key === "ArrowUp") {
        input.up();
      } else if (event.key === "l" || event.key === "ArrowRight") {
        input.right();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [
    input.input,
    input.toggleNotes,
    input.togglePing,
    input.up,
    input.down,
    input.left,
    input.right,
  ]);

  return (
    <KeyboardContext.Provider value={null}>{children}</KeyboardContext.Provider>
  );
};
