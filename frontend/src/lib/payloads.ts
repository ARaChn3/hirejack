export type LaneColor = "danger" | "warn" | "info" | "accent";

export const CATEGORY_META: Record<
  string,
  { color: LaneColor; impact: string; warnGlyph: boolean }
> = {
  dos: { color: "danger", impact: "high", warnGlyph: true },
  sqli: { color: "warn", impact: "injection", warnGlyph: true },
  integrity: { color: "info", impact: "silent", warnGlyph: false },
  ml_poison: { color: "accent", impact: "signature", warnGlyph: false },
};

export const DEFAULT_CATEGORY_META = {
  color: "info" as LaneColor,
  impact: "unknown",
  warnGlyph: false,
};

export const PAYLOAD_INFO: Record<string, string> = {
  entity_expansion:
    "Nests XML entity references to force exponential expansion during parsing (billion laughs).",
  recursion:
    "Embeds self-referential structures that recurse until the parser's stack or heap is exhausted.",
  bloat:
    "Pads the document with redundant objects to bloat memory and processing time on ingestion.",
  basic:
    "Injects a raw SQL fragment into resume text fields, hoping for unsanitized string concatenation.",
  blind:
    "Uses boolean/time-based conditions to exfiltrate data without visible query errors.",
  union:
    "Appends a UNION SELECT to pivot extracted text into an unintended result set.",
  unicode:
    "Embeds visually-identical unicode homoglyphs to desync string matching and keyword filters.",
  truncation:
    "Inserts null-byte / control-character sequences that truncate parsed fields early.",
  encoding:
    "Mixes encodings so the same content parses differently across pipeline stages.",
  skills:
    "Stuffs invisible or off-canvas text with high-value skill keywords to game ATS ranking.",
  titles:
    "Injects fabricated job titles tuned to trigger seniority/role heuristics in ranking models.",
  keywords:
    "Repeats target-role keywords in low-contrast or metadata fields to inflate match scores.",
};

export function payloadKey(categoryId: string, payloadId: string): string {
  return `${categoryId}:${payloadId}`;
}
