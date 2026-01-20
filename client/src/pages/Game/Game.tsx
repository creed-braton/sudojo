import {
  useEffect,
  useRef,
  useState,
  type ReactElement,
  type RefObject,
} from "react";
import { useBoard, type BoardContextProps } from "../../providers/board";
import { useInput, type InputContextProps } from "../../providers/input";
import { useSocket, type SocketContextProps } from "../../providers/socket";
import Board from "../../components/Board/Board";
import Input from "../../components/Input/Input";

const Game = (): ReactElement => {
  const socket: SocketContextProps = useSocket();
  const board: BoardContextProps = useBoard();
  const input: InputContextProps = useInput();

  const boardRef: RefObject<HTMLDivElement | null> =
    useRef<HTMLDivElement>(null);
  const [boardWidth, setBoardWidth] = useState<number>(0);

  useEffect((): void => {
    const id: string = location.pathname.split("/")[2];
    socket.setId(id);
  }, [location.pathname]);

  useEffect(() => {
    if (!boardRef.current) return;

    const observer = new ResizeObserver((entries) => {
      setBoardWidth(entries[0].contentRect.width);
    });

    observer.observe(boardRef.current);
    return () => observer.disconnect();
  }, [board.board]);

  useEffect((): void => {
    console.log(boardWidth);
  }, [boardWidth]);

  return (
    <div
      style={{
        height: "100%",
        width: "100%",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: "2%",
      }}
    >
      {board.board !== null && (
        <>
          <div
            ref={boardRef}
            style={{
              height: "80%",
              width: "fit-content",
              display: "flex",
              justifyContent: "center",
            }}
          >
            <Board
              board={board.board}
              cursor={board.cursor}
              setCursor={input.setCursor}
            />
          </div>
          <div
            style={{
              height: "20%",
              width: `${boardWidth}px`,
            }}
          >
            <Input />
          </div>
        </>
      )}
    </div>
  );
};

export default Game;
