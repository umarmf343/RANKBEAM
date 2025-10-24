import { useRankBeamStore } from "@/lib/state";
import { ClipboardCheck, ClipboardCopy, Wand2 } from "lucide-react";
import { useState } from "react";

export function CreativeStudio() {
  const { headlineIdeas, keyword } = useRankBeamStore();
  const [copied, setCopied] = useState<string | null>(null);

  const handleCopy = async (idea: string) => {
    try {
      await navigator.clipboard.writeText(idea);
      setCopied(idea);
      window.setTimeout(() => setCopied(null), 2000);
    } catch (error) {
      console.error("Clipboard not available", error);
    }
  };

  return (
    <section className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-3">
            <span className="inline-flex items-center gap-2 rounded-full border border-aurora-500/20 bg-aurora-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-wide text-aurora-100">
              <Wand2 className="h-4 w-4" /> Listing ideation
            </span>
            <h2 className="font-display text-3xl font-semibold text-white">Headline studio</h2>
            <p className="max-w-xl text-sm text-white/70">
              Instantly repurpose keyword intelligence into listing copy. Blend strategic adjectives with your seed idea
              to generate angle-rich headlines for A/B testing, product descriptions and ad hooks.
            </p>
          </div>
          <div className="rounded-3xl border border-white/10 bg-black/40 p-5 text-sm text-white/60">
            <p className="text-xs uppercase tracking-wide text-white/40">Seed keyword</p>
            <p className="mt-2 text-base font-semibold text-white">{keyword || "Add a seed keyword above"}</p>
            <p className="mt-3 text-xs text-white/50">
              Ideas update automatically as you refine your niche selection.
            </p>
          </div>
        </div>

        <div className="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {headlineIdeas.map((idea) => {
            const isCopied = copied === idea;
            return (
              <article key={idea} className="flex flex-col justify-between rounded-2xl border border-white/10 bg-black/40 p-5">
                <p className="text-base font-semibold text-white">{idea}</p>
                <button
                  type="button"
                  onClick={() => handleCopy(idea)}
                  className={`mt-6 inline-flex items-center gap-2 rounded-full border px-4 py-2 text-xs font-semibold uppercase tracking-wide transition ${
                    isCopied
                      ? "border-emerald-400/40 bg-emerald-500/20 text-emerald-100"
                      : "border-white/10 text-white/70 hover:border-aurora-400 hover:text-white"
                  }`}
                >
                  {isCopied ? (
                    <>
                      <ClipboardCheck className="h-4 w-4" /> Copied
                    </>
                  ) : (
                    <>
                      <ClipboardCopy className="h-4 w-4" /> Copy headline
                    </>
                  )}
                </button>
              </article>
            );
          })}
        </div>
      </div>
    </section>
  );
}

