import { useState, useMemo } from "react";
import { ChevronDown, Search, Zap, Eye, EyeOff } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CATEGORY_META, DEFAULT_CATEGORY_META, payloadKey } from "@/lib/payloads";
import type { Category, Payload } from "@/types";

const LANE_BORDER: Record<string, string> = {
  danger: "border-l-danger",
  warn: "border-l-warn",
  info: "border-l-info",
  accent: "border-l-accent",
};

const LANE_TEXT: Record<string, string> = {
  danger: "text-danger",
  warn: "text-warn",
  info: "text-info",
  accent: "text-accent",
};

const LANE_BG: Record<string, string> = {
  danger: "bg-danger/10",
  warn: "bg-warn/10",
  info: "bg-info/10",
  accent: "bg-accent/10",
};

function StealthBadge({ levels }: { levels: number[] }) {
  if (!levels?.length) return null;
  const min = Math.min(...levels);
  const max = Math.max(...levels);
  const label = min === max ? `L${min}` : `L${min}-${max}`;
  const color = min >= 4 ? "text-info" : min >= 2 ? "text-warn" : "text-danger";
  return (
    <span className={cn("flex items-center gap-0.5 font-mono text-xs", color)} title={`Stealth level ${label}`}>
      {min >= 3 ? <EyeOff className="size-3" /> : <Eye className="size-3" />}
      {label}
    </span>
  );
}

function ImpactBadge({ impact }: { impact: string }) {
  if (!impact) return null;
  const color =
    impact === "critical" ? "bg-danger/20 text-danger" :
    impact === "high" ? "bg-warn/20 text-warn" :
    "bg-info/20 text-info";
  return (
    <Badge variant="outline" className={cn("rounded-sm px-1.5 py-0 font-mono text-[10px] uppercase", color)}>
      {impact}
    </Badge>
  );
}

function PayloadToggle({
  categoryId,
  payload,
  lane,
  checked,
  onToggle,
  compact,
}: {
  categoryId: string;
  payload: Payload;
  lane: string;
  checked: boolean;
  onToggle: () => void;
  compact: boolean;
}) {
  const meta = CATEGORY_META[categoryId] ?? DEFAULT_CATEGORY_META;
  return (
    <label
      title={payload.description}
      className={cn(
        "flex cursor-pointer items-center justify-between gap-3 rounded-sm border border-transparent px-3 transition-colors",
        compact ? "py-1.5" : "py-2.5",
        checked ? "border-border bg-panel-raised" : "hover:bg-panel-raised/50",
      )}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className={cn("font-mono text-sm", checked ? LANE_TEXT[lane] : "text-text")}>
            {payload.id}
          </span>
          {meta.warnGlyph && (
            <span className="font-mono text-xs text-warn" aria-hidden>!</span>
          )}
          <ImpactBadge impact={payload.impact} />
          <StealthBadge levels={payload.stealth} />
        </div>
        {!compact && (
          <p className="mt-0.5 line-clamp-1 text-xs text-text-dim">
            {payload.description || "no description"}
          </p>
        )}
      </div>
      <Switch checked={checked} onCheckedChange={onToggle} />
    </label>
  );
}

function CategoryCard({
  category,
  selected,
  onToggle,
  onSetAll,
  compact,
  searchFilter,
}: {
  category: Category;
  selected: Set<string>;
  onToggle: (key: string) => void;
  onSetAll: (keys: string[], on: boolean) => void;
  compact: boolean;
  searchFilter: string;
}) {
  const meta = CATEGORY_META[category.id] ?? DEFAULT_CATEGORY_META;

  const filteredPayloads = useMemo(() => {
    if (!searchFilter) return category.payloads;
    const q = searchFilter.toLowerCase();
    return category.payloads.filter(
      (p) => p.id.toLowerCase().includes(q) || p.description?.toLowerCase().includes(q)
    );
  }, [category.payloads, searchFilter]);

  if (filteredPayloads.length === 0) return null;

  const allKeys = category.payloads.map((p) => payloadKey(category.id, p.id));
  const armedCount = allKeys.filter((k) => selected.has(k)).length;
  const allArmed = armedCount === category.payloads.length && category.payloads.length > 0;

  return (
    <Collapsible
      defaultOpen={false}
      className={cn(
        "rounded-sm border border-l-2 border-border bg-panel",
        LANE_BORDER[meta.color],
      )}
    >
      <div className="flex items-center justify-between gap-2 px-3 py-2.5">
        <div className="flex items-center gap-2">
          <span className={cn("text-sm font-semibold", LANE_TEXT[meta.color])}>{category.name}</span>
          <Badge variant="outline" className={cn("rounded-sm px-1.5 py-0 font-mono text-xs", LANE_BG[meta.color])}>
            {armedCount}/{category.payloads.length}
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5 font-mono text-xs text-text-faint">
            all
            <Switch
              checked={allArmed}
              onCheckedChange={(on) => onSetAll(allKeys, Boolean(on))}
            />
          </label>
          <CollapsibleTrigger className="flex size-5 items-center justify-center text-text-dim [&[data-state=open]>svg]:rotate-180">
            <ChevronDown className="size-3.5 transition-transform" />
          </CollapsibleTrigger>
        </div>
      </div>
      <CollapsibleContent>
        <div className="flex flex-col gap-px border-t border-border px-1.5 py-1.5">
          {filteredPayloads.map((payload) => {
            const key = payloadKey(category.id, payload.id);
            return (
              <PayloadToggle
                key={key}
                categoryId={category.id}
                payload={payload}
                lane={meta.color}
                checked={selected.has(key)}
                onToggle={() => onToggle(key)}
                compact={compact}
              />
            );
          })}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}

const PROFILES = [
  { id: "full_audit", label: "Full", desc: "All payloads", color: "accent" },
  { id: "sqli_focus", label: "SQLi", desc: "SQL injection only", color: "warn" },
  { id: "xxe_focus", label: "XXE", desc: "XXE attacks only", color: "danger" },
  { id: "llm_injection_focus", label: "LLM", desc: "Prompt injection for AI screeners", color: "danger" },
  { id: "stealth_recon", label: "Stealth", desc: "High stealth payloads only", color: "info" },
];

export function PayloadMatrix({
  categories,
  categoriesError,
  selected,
  onToggle,
  onSetAll,
}: {
  categories: Category[];
  categoriesError: string | null;
  selected: Set<string>;
  onToggle: (key: string) => void;
  onSetAll: (keys: string[], on: boolean) => void;
}) {
  const [search, setSearch] = useState("");
  const [compact, setCompact] = useState(false);

  const stats = useMemo(() => {
    const total = categories.reduce((acc, c) => acc + c.payloads.length, 0);
    const armed = selected.size;
    return { total, armed };
  }, [categories, selected]);

  if (categoriesError) {
    return (
      <div className="rounded-sm border border-danger/40 bg-danger/5 px-3 py-2 font-mono text-xs text-danger">
        failed to load: {categoriesError}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <div className="rounded-sm border border-border bg-panel px-3 py-2 font-mono text-xs text-text-faint">
        loading payloads…
      </div>
    );
  }

  const allKeys = categories.flatMap((cat) =>
    cat.payloads.map((p) => payloadKey(cat.id, p.id))
  );

  const applyProfile = (profileId: string) => {
    onSetAll(allKeys, false);

    if (profileId === "full_audit") {
      onSetAll(allKeys, true);
      return;
    }

    if (profileId === "stealth_recon") {
      const stealthKeys = categories.flatMap((c) =>
        c.payloads.filter((p) => p.stealth?.some((l) => l >= 4)).map((p) => payloadKey(c.id, p.id))
      );
      onSetAll(stealthKeys, true);
      return;
    }

    const catId = profileId.replace("_focus", "");
    const keys = categories
      .filter((c) => c.id === catId)
      .flatMap((c) => c.payloads.map((p) => payloadKey(c.id, p.id)));
    onSetAll(keys, true);
  };

  return (
    <div className="flex flex-col gap-3">
      {/* Header: search + stats + compact toggle */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-faint" />
          <input
            type="text"
            placeholder="filter payloads…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-sm border border-border bg-panel py-1.5 pl-8 pr-3 font-mono text-sm text-text placeholder:text-text-faint focus:border-accent focus:outline-none"
          />
        </div>
        <div className="flex items-center gap-3">
          <span className="font-mono text-xs text-text-dim">
            <Zap className="mr-1 inline size-3 text-accent" />
            {stats.armed}/{stats.total}
          </span>
          <label className="flex items-center gap-1.5 font-mono text-xs text-text-faint">
            compact
            <Switch checked={compact} onCheckedChange={setCompact} />
          </label>
        </div>
      </div>

      {/* Profiles */}
      <div className="flex flex-wrap gap-1.5">
        {PROFILES.map((p) => (
          <button
            key={p.id}
            onClick={() => applyProfile(p.id)}
            className={cn(
              "rounded-sm border border-border bg-panel px-2 py-1 font-mono text-xs transition-colors hover:border-accent hover:bg-panel-raised",
              LANE_TEXT[p.color]
            )}
            title={p.desc}
          >
            {p.label}
          </button>
        ))}
        <button
          onClick={() => onSetAll(allKeys, false)}
          className="rounded-sm border border-border bg-panel px-2 py-1 font-mono text-xs text-text-faint transition-colors hover:border-danger hover:text-danger"
        >
          Clear
        </button>
      </div>

      {/* Categories grid */}
      <div className="grid grid-cols-1 gap-2.5 lg:grid-cols-2">
        {categories.map((category) => (
          <CategoryCard
            key={category.id}
            category={category}
            selected={selected}
            onToggle={onToggle}
            onSetAll={onSetAll}
            compact={compact}
            searchFilter={search}
          />
        ))}
      </div>
    </div>
  );
}
