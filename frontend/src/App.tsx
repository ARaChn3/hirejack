import { useEffect, useReducer, useRef, useState } from "react";
import { BackdropShader } from "@/components/BackdropShader";
import { TopBar } from "@/components/TopBar";
import { Hero } from "@/components/Hero";
import { Stage, type StageState } from "@/components/Stage";
import { SourcePanel } from "@/components/SourcePanel";
import { PayloadMatrix } from "@/components/PayloadMatrix";
import { InjectPanel } from "@/components/InjectPanel";
import { ArtifactPanel } from "@/components/ArtifactPanel";
import { Button } from "@/components/ui/button";
import { reducer, initialState } from "@/reducer";
import { delay, mkArtifactFilename } from "@/lib/format";
import type { Category, FileMeta } from "@/types";

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function base64ToBlob(base64: string, type: string): Blob {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return new Blob([bytes], { type });
}

async function runInject(
  file: FileMeta,
  selected: Set<string>,
  dispatch: React.Dispatch<Parameters<typeof reducer>[1]>,
  canaryUrl?: string,
) {
  dispatch({ type: "START_INJECT" });

  const payloadIds = [...selected].map((k) => k.split(":")[1]);

  dispatch({ type: "LOG", line: "target loaded ✓" });
  await delay(200);
  dispatch({ type: "LOG", line: "encoding pdf..." });

  try {
    const buf = await file.file.arrayBuffer();
    const pdfBase64 = arrayBufferToBase64(buf);

    dispatch({ type: "LOG", line: `sending to injection engine (${payloadIds.length} payloads)...` });

    const res = await fetch("/api/inject", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        pdf: pdfBase64,
        payload_ids: payloadIds,
        stealth: 5, // Max stealth
        add_canary: true,
        canary_label: file.file.name,
        canary_url: canaryUrl || undefined,
      }),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || `HTTP ${res.status}`);
    }

    const data = await res.json();

    for (const inj of data.injections || []) {
      dispatch({ type: "LOG", line: `injected: ${inj.location}/${inj.field || "stream"} ${inj.encoded ? "(encoded)" : ""}` });
      await delay(100);
    }

    if (data.canary_token) {
      dispatch({ type: "LOG", line: `canary token: ${data.canary_token.id}` });
    }

    const blob = base64ToBlob(data.pdf, "application/pdf");
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
  const [canaryUrl, setCanaryUrl] = useState("");
  const toolRef = useRef<HTMLDivElement>(null);

  function scrollToTool() {
    toolRef.current?.scrollIntoView({ behavior: "smooth" });
  }

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

      <Hero onScrollToTool={scrollToTool} />

      <main ref={toolRef} className="mx-auto max-w-[880px] px-4 pt-24 pb-32">
        <div className="mb-12 border-b border-border pb-6">
          <h2 className="text-2xl font-bold tracking-tight text-text">Injection Pipeline</h2>
          <p className="mt-2 text-text-dim">
            Upload a resume, select payloads, and generate a poisoned artifact.
          </p>
        </div>

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
            canaryUrl={canaryUrl}
            onCanaryUrlChange={setCanaryUrl}
            onRemove={(key) => dispatch({ type: "TOGGLE_PAYLOAD", key })}
            onInject={() => {
              if (state.file) void runInject(state.file, state.selected, dispatch, canaryUrl);
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
          <div className="mx-auto flex h-16 max-w-[880px] items-center justify-between px-4">
            <span className="font-mono text-sm text-text-dim">
              {state.selected.size} payload{state.selected.size === 1 ? "" : "s"} armed
            </span>
            <Button
              size="sm"
              disabled={!stage03Unlocked || state.status === "injecting"}
              onClick={() => {
                if (state.file) void runInject(state.file, state.selected, dispatch, canaryUrl);
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
