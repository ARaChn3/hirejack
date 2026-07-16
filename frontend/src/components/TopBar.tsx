import { useEffect, useState } from "react";
import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils";
import { applyTheme, getStoredTheme, type Theme } from "@/lib/theme";
import type { Status } from "@/types";

const STATUS_LABEL: Record<Status, string> = {
  idle: "idle",
  armed: "armed",
  injecting: "injecting",
  done: "done",
  error: "error",
};

const STATUS_DOT: Record<Status, string> = {
  idle: "bg-text-faint",
  armed: "bg-accent",
  injecting: "bg-accent animate-pulse",
  done: "bg-accent",
  error: "bg-danger",
};

const STATUS_TEXT: Record<Status, string> = {
  idle: "text-text-faint",
  armed: "text-accent",
  injecting: "text-accent",
  done: "text-accent",
  error: "text-danger",
};

function StatusPill({ status }: { status: Status }) {
  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-full border border-border bg-panel px-3 py-1 font-mono text-xs",
        STATUS_TEXT[status],
      )}
    >
      <span className={cn("h-1.5 w-1.5 rounded-full", STATUS_DOT[status])} />
      {STATUS_LABEL[status]}
    </div>
  );
}

function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>("dark");

  useEffect(() => {
    setTheme(getStoredTheme());
  }, []);

  function toggle() {
    const next: Theme = theme === "dark" ? "light" : "dark";
    setTheme(next);
    applyTheme(next);
  }

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={theme === "dark" ? "switch to light theme" : "switch to dark theme"}
      className="flex h-7 w-7 items-center justify-center rounded-sm border border-border bg-panel text-text-dim transition-colors hover:border-border-active hover:text-text"
    >
      {theme === "dark" ? <Sun className="size-3.5" /> : <Moon className="size-3.5" />}
    </button>
  );
}

export function TopBar({ status }: { status: Status }) {
  return (
    <header className="fixed inset-x-0 top-0 z-20 border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex h-12 max-w-[880px] items-center justify-between px-4">
        <div className="flex items-baseline gap-1 font-mono text-sm">
          <span className="text-text">hirejack</span>
          <span className="text-accent">▸</span>
          <span className="text-text-faint">ats audit</span>
        </div>
        <div className="flex items-center gap-3">
          <StatusPill status={status} />
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
