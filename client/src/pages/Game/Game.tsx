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

  const boardRef: RefObject<HTMLTableElement | null> =
    useRef<HTMLTableElement>(null);
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

  return (
    <div
      style={{
        height: "100%",
        width: "100%",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "1%",
        gap: "2%",
        containerType: "size",
      }}
    >
      {board.board !== null && (
        <>
          <div
            style={{
              width: "min(100cqw, 80cqh)",
            }}
          >
            <Board
              board={board.board}
              cursor={board.cursor}
              setCursor={input.setCursor}
              ref={boardRef}
            />
          </div>
          <div style={{ maxWidth: `${boardWidth}px` }}>
            <Input
              mode={input.mode}
              togglePing={input.togglePing}
              toggleNotes={input.toggleNotes}
              input={input.input}
            />
          </div>
        </>
      )}
    </div>
  );
};

export default Game;
