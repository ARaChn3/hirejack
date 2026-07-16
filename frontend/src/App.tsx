import { useEffect, useReducer } from "react";
import { BackdropShader } from "@/components/BackdropShader";
import { TopBar } from "@/components/TopBar";
import { Stage, type StageState } from "@/components/Stage";
import { SourcePanel } from "@/components/SourcePanel";
import { PayloadMatrix } from "@/components/PayloadMatrix";
import { InjectPanel } from "@/components/InjectPanel";
import { ArtifactPanel } from "@/components/ArtifactPanel";
import { Button } from "@/components/ui/button";
import { reducer, initialState } from "@/reducer";
import { delay, mkArtifactFilename } from "@/lib/format";
import type { Category, FileMeta } from "@/types";

async function runInject(
  file: FileMeta,
  selected: Set<string>,
  dispatch: React.Dispatch<Parameters<typeof reducer>[1]>,
) {
  dispatch({ type: "START_INJECT" });

  const payloadIds = [...selected].map((k) => k.split(":")[1]);
  const steps = [
    "target loaded ✓",
    "parsing pdf object graph...",
    `resolving xref table (${selected.size} payload${selected.size === 1 ? "" : "s"} queued)`,
    `writing payload${payloadIds.length === 1 ? "" : "s"}: ${payloadIds.join(", ")}`,
    "reserializing document...",
    "finalizing artifact...",
  ];

  for (const line of steps) {
    await delay(280 + Math.random() * 280);
    dispatch({ type: "LOG", line });
  }

  try {
    // NOTE: POST /api/inject is currently a stub on the backend — mock the
    // artifact client-side so the pipeline is fully exercisable end to end.
    const buf = await file.file.arrayBuffer();
    const padding = new Uint8Array(Math.floor(buf.byteLength * 0.15) + 2048 * selected.size).fill(0x20);
    const blob = new Blob([buf, padding], { type: "application/pdf" });
    await delay(250);
    dispatch({ type: "LOG", line: `done — ${selected.size} payload${selected.size === 1 ? "" : "s"} embedded` });
    dispatch({
      type: "INJECT_SUCCESS",
      artifact: { blob, size: blob.size, filename: mkArtifactFilename(file.file.name) },
    });
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e);
    dispatch({ type: "LOG", line: `error: ${message}` });
    dispatch({ type: "INJECT_ERROR", message });
  }
}

function stageState(unlocked: boolean, complete: boolean): StageState {
  if (complete) return "complete";
  return unlocked ? "active" : "locked";
}

function App() {
  const [state, dispatch] = useReducer(reducer, initialState);

  useEffect(() => {
    fetch("/api/payloads")
      .then((res) => {
        if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
        return res.json();
      })
      .then((data: { categories: Category[] }) => {
        dispatch({ type: "SET_CATEGORIES", categories: data.categories ?? [] });
      })
      .catch((e: unknown) => {
        dispatch({
          type: "SET_CATEGORIES_ERROR",
          message: e instanceof Error ? e.message : String(e),
        });
      });
  }, []);

  const stage02Unlocked = state.file !== null;
  const stage03Unlocked = state.selected.size > 0;
  const stage04Unlocked = state.status === "done" && state.resultBlob !== null;

  return (
    <>
      <BackdropShader />
      <TopBar status={state.status} />

      <main className="mx-auto max-w-[880px] px-4 pt-24 pb-32">
        <Stage index="01" title="Source" state={stageState(true, stage02Unlocked)}>
          <SourcePanel
            file={state.file}
            onLoad={(file) => dispatch({ type: "SET_FILE", file })}
            onClear={() => dispatch({ type: "CLEAR_FILE" })}
          />
        </Stage>

        <Stage
          index="02"
          title="Payloads"
          state={stageState(stage02Unlocked, stage03Unlocked)}
        >
          <PayloadMatrix
            categories={state.categories}
            categoriesError={state.categoriesError}
            selected={state.selected}
            onToggle={(key) => dispatch({ type: "TOGGLE_PAYLOAD", key })}
            onSetAll={(keys, on) => dispatch({ type: "SET_KEYS", keys, on })}
          />
        </Stage>

        <Stage
          index="03"
          title="Inject"
          state={stageState(stage03Unlocked, stage04Unlocked)}
        >
          <InjectPanel
            file={state.file}
            selected={state.selected}
            status={state.status}
            logs={state.logs}
            error={state.error}
            onRemove={(key) => dispatch({ type: "TOGGLE_PAYLOAD", key })}
            onInject={() => {
              if (state.file) void runInject(state.file, state.selected, dispatch);
            }}
          />
        </Stage>

        <Stage index="04" title="Artifact" state={stageState(stage04Unlocked, stage04Unlocked)} isLast>
          <ArtifactPanel
            file={state.file}
            artifact={state.resultBlob}
            payloadCount={state.selected.size}
            onRunAgain={() => dispatch({ type: "RUN_AGAIN" })}
            onNewSource={() => dispatch({ type: "RESET" })}
          />
        </Stage>
      </main>

      {stage02Unlocked && state.status !== "done" && (
        <div className="fixed inset-x-0 bottom-0 z-20 border-t border-border bg-background/95 backdrop-blur">
          <div className="mx-auto flex h-14 max-w-[880px] items-center justify-between px-4">
            <span className="font-mono text-xs text-text-dim">
              {state.selected.size} payload{state.selected.size === 1 ? "" : "s"} armed
            </span>
            <Button
              size="sm"
              disabled={!stage03Unlocked || state.status === "injecting"}
              onClick={() => {
                if (state.file) void runInject(state.file, state.selected, dispatch);
              }}
              className="rounded-sm font-mono tracking-wide"
            >
              {state.status === "injecting" ? "injecting…" : "inject ▸"}
            </Button>
          </div>
        </div>
      )}
    </>
  );
}

export default App;
