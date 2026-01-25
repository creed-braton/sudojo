import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
  type RefObject,
} from "react";
import { useSocket, type SocketContextProps } from "./socket";
import { isUUID } from "../api/types";

export type NotesContextProps = {
  notes: Map<string, Set<number>>;
  insert: (row: number, column: number, value: number) => void;
};

const encode = (notes: Map<string, Set<number>>): string => {
  const obj: Record<string, number[]> = {};
  for (const [key, value] of notes) {
    obj[key] = [...value];
  }
  return JSON.stringify(obj);
};

const decode = (data: string): Map<string, Set<number>> => {
  const obj: Record<string, number[]> = JSON.parse(data);
  const map = new Map<string, Set<number>>();
  for (const [key, value] of Object.entries(obj)) {
    map.set(key, new Set(value));
  }
  return map;
};

const save = (id: string, notes: Map<string, Set<number>>): void => {
  localStorage.setItem(id, encode(notes));
};

const load = (id: string): Map<string, Set<number>> => {
  const data: string | null = localStorage.getItem(id);
  if (data === null) return new Map();

  try {
    return decode(data);
  } catch {
    return new Map();
  }
};

const NotesContext = createContext<NotesContextProps | null>(null);

type NotesProviderProps = {
  children: ReactNode;
};

export const NotesProvider = ({ children }: NotesProviderProps) => {
  const { id, initial, current, setOnInsert }: SocketContextProps = useSocket();
  const [notes, setNotes] = useState<Map<string, Set<number>>>(new Map());
  const idRef: RefObject<string> = useRef<string>(id);

  useEffect(() => {
    idRef.current = id;
  }, [id]);

  useEffect((): void => {
    isUUID(id) && setNotes(load(id));
  }, [id]);

  const onInsert = (row: number, column: number, value: number): void =>
    setNotes((prev: Map<string, Set<number>>) => {
      const newNotes: Map<string, Set<number>> = new Map(prev);
      let changed: boolean = false;

      const boxRow = Math.floor(row / 3) * 3;
      const boxCol = Math.floor(column / 3) * 3;

      for (const [key, cell] of newNotes) {
        const [r, c] = key.split("-").map(Number);

        const sameRow: boolean = r === row;
        const sameCol: boolean = c === column;
        const sameBox: boolean =
          r >= boxRow && r < boxRow + 3 && c >= boxCol && c < boxCol + 3;

        if ((sameRow || sameCol || sameBox) && cell.has(value)) {
          cell.delete(value);
          changed = true;
          if (cell.size === 0) {
            newNotes.delete(key);
          }
        }
      }

      if (changed) {
        save(idRef.current, newNotes);
      }

      return changed ? newNotes : prev;
    });

  useEffect(() => {
    setOnInsert(onInsert);
  }, [setOnInsert]);

  const insert = (row: number, column: number, value: number): void => {
    if (initial === null || current === null) return;
    if (initial[row][column] !== 0 || current[row][column] !== 0) return;

    const key: string = `${row}-${column}`;
    const newNotes: Map<string, Set<number>> = new Map(notes);

    let cell: Set<number> | undefined = newNotes.get(key);
    if (value === 0) {
      cell = undefined;
    } else if (cell === undefined) {
      cell = new Set([value]);
    } else if (cell.has(value)) {
      cell.delete(value);
    } else {
      cell.add(value);
    }

    if (cell === undefined) {
      newNotes.delete(key);
    } else {
      newNotes.set(key, cell);
    }
    setNotes(newNotes);
    save(id, newNotes);
  };

  const value: NotesContextProps = { notes, insert };

  return (
    <NotesContext.Provider value={value}>{children}</NotesContext.Provider>
  );
};

export const useNotes = (): NotesContextProps => {
  const context = useContext(NotesContext);
  if (!context) {
    throw new Error("useNotes must be used within a NotesProvider");
  }
  return context;
};
