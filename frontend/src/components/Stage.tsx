import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export type StageState = "locked" | "active" | "complete";

export function Stage({
  index,
  title,
  state,
  isLast = false,
  children,
}: {
  index: string;
  title: string;
  state: StageState;
  isLast?: boolean;
  children: ReactNode;
}) {
  return (
    <div
      className={cn(
        "relative flex gap-5 transition-opacity duration-300",
        state === "locked" && "pointer-events-none opacity-40",
      )}
    >
      <div className="flex w-3 shrink-0 flex-col items-center pt-1.5">
        <div
          className={cn(
            "h-2.5 w-2.5 shrink-0 rounded-full border transition-all",
            state === "complete" && "border-accent bg-accent",
            state === "active" &&
              "border-accent bg-transparent shadow-[0_0_10px_var(--color-accent)]",
            state === "locked" && "border-border bg-panel",
          )}
        />
        {!isLast && (
          <div
            className={cn(
              "mt-1 w-px flex-1",
              state === "complete" ? "bg-accent-dim" : "bg-border",
            )}
          />
        )}
      </div>

      <div className="min-w-0 flex-1 pb-14">
        <div className="mb-5 flex items-baseline gap-3">
          <span className="font-mono text-base text-text-faint">[{index}]</span>
          <h2 className="text-lg font-semibold tracking-[0.12em] text-text uppercase">
            {title}
          </h2>
        </div>
        {children}
      </div>
    </div>
  );
}
