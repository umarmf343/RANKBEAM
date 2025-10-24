import { useRankBeamStore } from "@/lib/state";
import { Globe } from "lucide-react";

export function ExpansionPanel() {
  const { internationalKeywords } = useRankBeamStore();

  return (
    <section id="expansion" className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">International Expansion Lab</h2>
            <p className="mt-2 max-w-2xl text-sm text-white/70">
              Instantly localise your keyword portfolio. RankBeam evaluates linguistic variants, currency demand and
              seasonal trends to translate purchase intent across Amazon marketplaces.
            </p>
          </div>
          <span className="inline-flex items-center gap-2 rounded-full border border-aurora-500/20 bg-aurora-500/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-100">
            <Globe className="h-4 w-4" /> Multilingual ready
          </span>
        </div>
        <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {internationalKeywords.map((entry) => (
            <div key={entry.countryCode} className="rounded-2xl border border-white/10 bg-black/40 p-5">
              <div className="flex items-center justify-between">
                <p className="text-sm font-semibold text-white">{entry.countryName}</p>
                <span className="text-xs text-white/40">{entry.countryCode}</span>
              </div>
              <p className="mt-3 text-sm text-white/70">Suggested keyword</p>
              <p className="mt-1 text-base font-medium text-aurora-200">{entry.keyword}</p>
              <p className="mt-4 text-xs uppercase tracking-wide text-white/50">Search Volume</p>
              <p className="text-xl font-semibold text-white">{entry.searchVolume.toLocaleString()}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
