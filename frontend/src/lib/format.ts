export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.min(
    units.length - 1,
    Math.floor(Math.log(bytes) / Math.log(1024)),
  );
  const value = bytes / 1024 ** i;
  return `${i === 0 ? value : value.toFixed(2)} ${units[i]}`;
}

export async function sha256Hex(buf: ArrayBuffer): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", buf);
  return [...new Uint8Array(digest)]
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

/** Cheap heuristic page count for uncompressed / lightly-compressed PDFs. */
export function estimatePageCount(buf: ArrayBuffer): number | null {
  try {
    const bytes = new Uint8Array(buf);
    let text = "";
    for (let i = 0; i < bytes.length; i++) text += String.fromCharCode(bytes[i]);
    const matches = text.match(/\/Type\s*\/Page(?!s)/g);
    return matches ? matches.length : null;
  } catch {
    return null;
  }
}

export function mkArtifactFilename(originalName: string): string {
  const dot = originalName.lastIndexOf(".");
  const base = dot === -1 ? originalName : originalName.slice(0, dot);
  return `${base}__hirejacked.pdf`;
}

export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
