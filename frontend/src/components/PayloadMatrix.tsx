import { ChevronDown } from "lucide-react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CATEGORY_META, DEFAULT_CATEGORY_META, PAYLOAD_INFO, payloadKey } from "@/lib/payloads";
import type { Category } from "@/types";

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

function PayloadToggle({
  categoryId,
  payloadId,
  lane,
  checked,
  onToggle,
}: {
  categoryId: string;
  payloadId: string;
  lane: string;
  checked: boolean;
  onToggle: () => void;
}) {
  const meta = CATEGORY_META[categoryId] ?? DEFAULT_CATEGORY_META;
  return (
    <label
      title={PAYLOAD_INFO[payloadId] ?? ""}
      className={cn(
        "flex cursor-pointer items-center justify-between gap-4 rounded-sm border border-transparent px-3 py-2.5 transition-colors",
        checked ? "border-border bg-panel-raised" : "hover:bg-panel-raised/50",
      )}
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "font-mono text-base",
              checked ? LANE_TEXT[lane] : "text-text",
            )}
          >
            {payloadId}
          </span>
          {meta.warnGlyph && (
            <span className="font-mono text-base text-warn" aria-hidden>
              !
            </span>
          )}
        </div>
        <p className="mt-1 line-clamp-1 text-sm text-text-dim">
          {PAYLOAD_INFO[payloadId] ?? "no description available"}
        </p>
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
}: {
  category: Category;
  selected: Set<string>;
  onToggle: (key: string) => void;
  onSetAll: (keys: string[], on: boolean) => void;
}) {
  const meta = CATEGORY_META[category.id] ?? DEFAULT_CATEGORY_META;
  const keys = category.payloads.map((p) => payloadKey(category.id, p));
  const armedCount = keys.filter((k) => selected.has(k)).length;
  const allArmed = armedCount === keys.length && keys.length > 0;

  return (
    <Collapsible
      defaultOpen
      className={cn(
        "rounded-sm border border-l-2 border-border bg-panel",
        LANE_BORDER[meta.color],
      )}
    >
      <div className="flex items-center justify-between gap-2 px-3 py-3">
        <div className="flex items-center gap-2">
          <span className="text-base font-semibold text-text">{category.name}</span>
          <Badge variant="outline" className="rounded-sm font-mono text-sm text-text-dim">
            {armedCount}/{keys.length}
          </Badge>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 font-mono text-sm text-text-faint">
            arm all
            <Switch
              checked={allArmed}
              onCheckedChange={(on) => onSetAll(keys, Boolean(on))}
            />
          </label>
          <CollapsibleTrigger className="flex h-5 w-5 items-center justify-center text-text-dim [&[data-panel-open]>svg]:rotate-180">
            <ChevronDown className="size-3.5 transition-transform" />
          </CollapsibleTrigger>
        </div>
      </div>
      <CollapsibleContent>
        <div className="flex flex-col gap-0.5 border-t border-border px-2 py-2">
          {category.payloads.map((payloadId) => {
            const key = payloadKey(category.id, payloadId);
            return (
              <PayloadToggle
                key={key}
                categoryId={category.id}
                payloadId={payloadId}
                lane={meta.color}
                checked={selected.has(key)}
                onToggle={() => onToggle(key)}
              />
            );
          })}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}

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
  if (categoriesError) {
    return (
      <div className="rounded-sm border border-danger/40 bg-danger/5 px-3 py-2 font-mono text-xs text-danger">
        failed to load payload catalog: {categoriesError}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <div className="rounded-sm border border-border bg-panel px-3 py-2 font-mono text-xs text-text-faint">
        loading payload catalog…
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {categories.map((category) => (
        <CategoryCard
          key={category.id}
          category={category}
          selected={selected}
          onToggle={onToggle}
          onSetAll={onSetAll}
        />
      ))}
    </div>
  );
}
