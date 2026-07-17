import { ChevronDown } from "lucide-react";
import { Dithering } from "@paper-design/shaders-react";

export function Hero({ onScrollToTool }: { onScrollToTool: () => void }) {
  return (
    <section className="relative flex min-h-screen flex-col items-center justify-center px-4 pt-20">
      {/* Dither shader bg */}
      <div className="pointer-events-none absolute inset-0 -z-10 opacity-20">
        <Dithering
          colorBack="#0a0b0d"
          colorFront="#a3e635"
          shape="swirl"
          type="8x8"
          size={2}
          speed={0.05}
          style={{ width: "100%", height: "100%" }}
        />
      </div>

      <h1 className="max-w-3xl text-center font-serif text-4xl tracking-tight text-text sm:text-5xl md:text-6xl">
        Test your ATS
        <br />
        <em className="text-accent">before attackers do</em>
      </h1>

      <button
        onClick={onScrollToTool}
        className="absolute bottom-8 left-1/2 -translate-x-1/2 animate-bounce text-text-faint transition-colors hover:text-accent"
        aria-label="Scroll to tool"
      >
        <ChevronDown className="size-6" />
      </button>
    </section>
  );
}
