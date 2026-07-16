import { useRef, useState } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import { estimatePageCount, formatBytes, sha256Hex } from "@/lib/format";
import type { FileMeta } from "@/types";

export function SourcePanel({
  file,
  onLoad,
  onClear,
}: {
  file: FileMeta | null;
  onLoad: (file: FileMeta) => void;
  onClear: () => void;
}) {
  const [dragOver, setDragOver] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  async function loadFile(f: File) {
    if (f.type !== "application/pdf" && !f.name.toLowerCase().endsWith(".pdf")) {
      setError("unsupported: expected application/pdf");
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const buf = await f.arrayBuffer();
      const hash = await sha256Hex(buf);
      const pages = estimatePageCount(buf);
      onLoad({ file: f, size: f.size, hash, pages });
    } catch {
      setError("failed to read file");
    } finally {
      setLoading(false);
    }
  }

  if (file) {
    return (
      <div className="rounded-sm border border-border-active bg-panel-raised px-4 py-3">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="truncate font-mono text-sm text-text">{file.file.name}</div>
            <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 font-mono text-xs text-text-dim">
              <span>{formatBytes(file.size)}</span>
              {file.pages !== null && <span>{file.pages} page{file.pages === 1 ? "" : "s"}</span>}
              <span className="text-text-faint">sha256:{file.hash.slice(0, 16)}…</span>
            </div>
          </div>
          <button
            type="button"
            onClick={onClear}
            aria-label="clear resume"
            className="flex h-6 w-6 shrink-0 items-center justify-center rounded-sm border border-border text-text-dim hover:border-danger hover:text-danger"
          >
            <X className="size-3.5" />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault();
          setDragOver(false);
          const f = e.dataTransfer.files[0];
          if (f) void loadFile(f);
        }}
        className={cn(
          "relative flex cursor-pointer flex-col items-center justify-center overflow-hidden rounded-sm border border-dashed px-6 py-10 text-center transition-colors",
          dragOver
            ? "border-accent bg-accent/5"
            : error
              ? "border-danger"
              : "border-border hover:border-border-active",
        )}
      >
        {loading && (
          <div className="absolute inset-x-0 top-0 h-px overflow-hidden bg-border">
            <div className="h-full w-1/3 animate-[scan_1.2s_ease-in-out_infinite] bg-accent" />
          </div>
        )}
        <div className={cn("font-mono text-xs", dragOver ? "text-accent" : "text-text-dim")}>
          {dragOver
            ? "release to load"
            : error
              ? error
              : "drop resume.pdf or click to browse · pdf only"}
        </div>
        <input
          ref={inputRef}
          type="file"
          accept="application/pdf,.pdf"
          className="hidden"
          onChange={(e) => {
            const f = e.target.files?.[0];
            if (f) void loadFile(f);
            e.target.value = "";
          }}
        />
      </div>
      <style>{`
        @keyframes scan {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(300%); }
        }
      `}</style>
    </div>
  );
}
