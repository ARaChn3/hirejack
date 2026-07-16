/**
 * Fixed, full-viewport backdrop: a faint static dither/grain over a dark
 * radial gradient. Single layer, no animation, low contrast — seasoning,
 * not a feature. See UI-SPEC.md §6.
 */
export function BackdropShader() {
  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-0 -z-10"
      style={{
        backgroundColor: "var(--background)",
        backgroundImage: [
          "radial-gradient(ellipse 80% 60% at 50% -10%, color-mix(in oklab, var(--accent) 6%, transparent), transparent)",
          "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.05'/%3E%3C/svg%3E\")",
        ].join(", "),
      }}
    />
  );
}
