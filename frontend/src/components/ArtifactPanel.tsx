import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { formatBytes } from "@/lib/format";
import type { ArtifactMeta, FileMeta } from "@/types";

export function ArtifactPanel({
  file,
  artifact,
  payloadCount,
  onRunAgain,
  onNewSource,
}: {
  file: FileMeta | null;
  artifact: ArtifactMeta | null;
  payloadCount: number;
  onRunAgain: () => void;
  onNewSource: () => void;
}) {
  const urlRef = useRef<string | null>(null);

  useEffect(() => {
    return () => {
      if (urlRef.current) URL.revokeObjectURL(urlRef.current);
    };
  }, [artifact]);

  if (!artifact || !file) {
    return <p className="font-mono text-sm text-text-faint">awaiting artifact…</p>;
  }

  const delta = artifact.size - file.size;

  function download() {
    if (!artifact) return;
    const url = URL.createObjectURL(artifact.blob);
    urlRef.current = url;
    const a = document.createElement("a");
    a.href = url;
    a.download = artifact.filename;
    a.click();
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-sm border border-border bg-panel px-3 py-3">
        <dl className="grid grid-cols-2 gap-x-6 gap-y-3 font-mono text-sm sm:grid-cols-4">
          <div>
            <dt className="text-text-faint">original</dt>
            <dd className="text-text">{formatBytes(file.size)}</dd>
          </div>
          <div>
            <dt className="text-text-faint">poisoned</dt>
            <dd className="text-text">{formatBytes(artifact.size)}</dd>
          </div>
          <div>
            <dt className="text-text-faint">delta</dt>
            <dd className={delta >= 0 ? "text-accent" : "text-warn"}>
              {formatBytes(delta, true)}
            </dd>
          </div>
          <div>
            <dt className="text-text-faint">payloads</dt>
            <dd className="text-text">{payloadCount}</dd>
          </div>
        </dl>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Button
          onClick={download}
          className="rounded-sm font-mono tracking-wide shadow-[0_0_0_0_var(--color-accent)] transition-shadow hover:shadow-[0_0_16px_-2px_var(--color-accent)]"
        >
          download poisoned.pdf ▸
        </Button>
        <Button
          variant="outline"
          onClick={onRunAgain}
          className="rounded-sm font-mono text-sm"
        >
          run again
        </Button>
        <Button
          variant="ghost"
          onClick={onNewSource}
          className="rounded-sm font-mono text-sm text-text-dim"
        >
          new source
        </Button>
      </div>
    </div>
  );
}
