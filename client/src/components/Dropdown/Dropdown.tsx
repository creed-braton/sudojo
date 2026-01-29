import {
  useState,
  useEffect,
  useRef,
  type ReactElement,
  type ReactNode,
} from "react";
import ArrowIcon from "../../icons/Arrow";
import style from "./Dropdown.module.css";

type DropdownProps = {
  label: ReactNode;
  children: ReactNode;
};

const Dropdown = ({ label, children }: DropdownProps): ReactElement => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  return (
    <div className={style.container} ref={containerRef}>
      <button
        className={style.trigger}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
      >
        <span className={style.label}>{label}</span>
        <ArrowIcon
          className={style.arrow}
          style={{
            transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
            transition: "transform 400ms ease",
          }}
        />
      </button>
      {isOpen && <ul className={style.list}>{children}</ul>}
    </div>
  );
};

export default Dropdown;
