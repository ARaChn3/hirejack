import { useEffect, useRef, useState } from "react";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { formatBytes } from "@/lib/format";
import { CATEGORY_META, DEFAULT_CATEGORY_META } from "@/lib/payloads";
import type { FileMeta, Status } from "@/types";

const LANE_TEXT: Record<string, string> = {
  danger: "text-danger border-danger/40",
  warn: "text-warn border-warn/40",
  info: "text-info border-info/40",
  accent: "text-accent border-accent/40",
};

function laneFor(categoryId: string) {
  return CATEGORY_META[categoryId] ?? DEFAULT_CATEGORY_META;
}

function useElapsed(active: boolean) {
  const [ms, setMs] = useState(0);
  const startRef = useRef(0);
  useEffect(() => {
    if (!active) {
      setMs(0);
      return;
    }
    startRef.current = Date.now();
    const id = setInterval(() => setMs(Date.now() - startRef.current), 100);
    return () => clearInterval(id);
  }, [active]);
  return ms;
}

export function InjectPanel({
  file,
  selected,
  status,
  logs,
  error,
  onRemove,
  onInject,
}: {
  file: FileMeta | null;
  selected: Set<string>;
  status: Status;
  logs: string[];
  error: string | null;
  onRemove: (key: string) => void;
  onInject: () => void;
}) {
  const elapsed = useElapsed(status === "injecting");
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [logs]);

  const armedKeys = [...selected];
  const estBytes = file ? Math.round(file.size * 0.15) + 2048 * armedKeys.length : 0;

  const reason = !file
    ? "load a resume"
    : selected.size === 0
      ? "arm ≥ 1 payload"
      : null;

  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-sm border border-border bg-panel px-3 py-3">
        {armedKeys.length === 0 ? (
          <p className="font-mono text-xs text-text-faint">no payloads armed</p>
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {armedKeys.map((key) => {
              const [categoryId, payloadId] = key.split(":");
              const lane = laneFor(categoryId);
              return (
                <span
                  key={key}
                  className={cn(
                    "flex items-center gap-1 rounded-sm border bg-panel-raised px-1.5 py-0.5 font-mono text-[11px]",
                    LANE_TEXT[lane.color],
                  )}
                >
                  {payloadId}
                  <button
                    type="button"
                    aria-label={`remove ${payloadId}`}
                    onClick={() => onRemove(key)}
                    disabled={status === "injecting"}
                    className="text-text-faint hover:text-text"
                  >
                    <X className="size-2.5" />
                  </button>
                </span>
              );
            })}
          </div>
        )}
        {armedKeys.length > 0 && (
          <p className="mt-2 font-mono text-[11px] text-text-dim">
            {armedKeys.length} payload{armedKeys.length === 1 ? "" : "s"} armed · est. artifact +
            {formatBytes(estBytes)}
          </p>
        )}
      </div>

      <div className="flex items-center gap-3">
        <Button
          onClick={onInject}
          disabled={Boolean(reason) || status === "injecting"}
          className="rounded-sm font-mono tracking-wide shadow-[0_0_0_0_var(--color-accent)] transition-shadow hover:shadow-[0_0_16px_-2px_var(--color-accent)] disabled:shadow-none"
        >
          {status === "injecting"
            ? `injecting… ${(elapsed / 1000).toFixed(1)}s`
            : status === "error"
              ? "retry ▸"
              : "inject ▸"}
        </Button>
        {reason && <span className="font-mono text-xs text-text-faint">{reason}</span>}
      </div>

      {(status === "injecting" || logs.length > 0) && (
        <div
          className={cn(
            "relative overflow-hidden rounded-sm border border-border bg-[#050607] px-3 py-2",
            status === "injecting" &&
              "shadow-[inset_0_0_40px_-10px_var(--color-accent)]",
          )}
        >
          <ScrollArea className="h-32">
            <div ref={scrollRef} className="flex flex-col gap-0.5 pr-2 font-mono text-xs">
              {logs.map((line, i) => (
                <div key={i} className="text-text-dim">
                  <span className="text-text-faint">$ </span>
                  {line}
                </div>
              ))}
              {status === "error" && error && (
                <div className="text-danger">
                  <span className="text-text-faint">$ </span>
                  error: {error}
                </div>
              )}
              {status === "injecting" && (
                <div className="text-accent">
                  <span className="text-text-faint">$ </span>_
                </div>
              )}
            </div>
          </ScrollArea>
        </div>
      )}
    </div>
  );
}
