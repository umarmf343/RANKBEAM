import { useRankBeamStore } from "@/lib/state";
import { BrainCircuit, Filter, Sparkles, Target } from "lucide-react";
import { useMemo } from "react";

function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

export function InsightSummary() {
  const { keyword, keywordInsights, growthSignals } = useRankBeamStore();

  const summary = useMemo(() => {
    if (keywordInsights.length === 0) {
      return {
        topKeyword: "",
        topVolume: 0,
        totalVolume: 0,
        averageCompetition: 0,
        opportunityCount: 0,
        averageRelevancy: 0
      };
    }

    const sorted = [...keywordInsights].sort((a, b) => b.searchVolume - a.searchVolume);
    const top = sorted[0];
    const totalVolume = keywordInsights.reduce((acc, row) => acc + row.searchVolume, 0);
    const averageCompetition =
      keywordInsights.reduce((acc, row) => acc + row.competitionScore, 0) / keywordInsights.length;
    const opportunityCount = keywordInsights.filter((row) => row.competitionScore < 3.2 && row.titleDensity < 4).length;
    const averageRelevancy =
      keywordInsights.reduce((acc, row) => acc + row.relevancyScore, 0) / keywordInsights.length;

    return {
      topKeyword: top.keyword,
      topVolume: top.searchVolume,
      totalVolume,
      averageCompetition,
      opportunityCount,
      averageRelevancy
    };
  }, [keywordInsights]);

  const highlightSignal = useMemo(() => {
    if (!growthSignals.length) return null;
    return [...growthSignals].sort((a, b) => b.score - a.score)[0];
  }, [growthSignals]);

  return (
    <section className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
          <div className="max-w-2xl space-y-3">
            <span className="inline-flex items-center gap-2 rounded-full border border-aurora-500/20 bg-aurora-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-wide text-aurora-100">
              <BrainCircuit className="h-4 w-4" /> Opportunity Intelligence
            </span>
            <h2 className="font-display text-3xl font-semibold text-white">Strategic snapshot</h2>
            <p className="text-sm text-white/70">
              RankBeam analysed <strong className="text-white">{keywordInsights.length}</strong> keyword variants for
              “{keyword || "your next launch"}”. We combine search volume, listing density and relevancy to expose the
              clearest path to organic visibility.
            </p>
          </div>
          {highlightSignal && (
            <div className="rounded-3xl border border-aurora-500/30 bg-aurora-500/10 p-5 text-sm text-aurora-100">
              <p className="text-xs uppercase tracking-wide text-aurora-200/80">Leading signal</p>
              <p className="mt-2 text-base font-semibold text-white">{highlightSignal.label}</p>
              <p className="mt-1 text-xs text-aurora-100/80">{highlightSignal.description}</p>
              <p className="mt-3 text-2xl font-semibold text-white">{highlightSignal.score}</p>
            </div>
          )}
        </div>

        <div className="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div className="rounded-2xl border border-white/10 bg-black/40 p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs uppercase tracking-wide text-white/50">Top keyword</p>
              <Target className="h-4 w-4 text-aurora-400" />
            </div>
            <p className="mt-3 text-base font-semibold text-white">{summary.topKeyword || "Collecting insights"}</p>
            <p className="mt-2 text-xs text-white/50">{summary.topVolume ? `${formatNumber(summary.topVolume)} searches/mo` : "Driven by live refresh"}</p>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/40 p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs uppercase tracking-wide text-white/50">Opportunity window</p>
              <Filter className="h-4 w-4 text-aurora-400" />
            </div>
            <p className="mt-3 text-base font-semibold text-white">{summary.opportunityCount}</p>
            <p className="mt-2 text-xs text-white/50">Low competition keywords with thin title density</p>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/40 p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs uppercase tracking-wide text-white/50">Visibility potential</p>
              <Sparkles className="h-4 w-4 text-aurora-400" />
            </div>
            <p className="mt-3 text-base font-semibold text-white">{formatNumber(summary.totalVolume)}</p>
            <p className="mt-2 text-xs text-white/50">Combined monthly searches across the cluster</p>
          </div>

          <div className="rounded-2xl border border-white/10 bg-black/40 p-5">
            <div className="flex items-center justify-between">
              <p className="text-xs uppercase tracking-wide text-white/50">Market friction</p>
              <BrainCircuit className="h-4 w-4 text-aurora-400" />
            </div>
            <p className="mt-3 text-base font-semibold text-white">{summary.averageCompetition.toFixed(2)}</p>
            <p className="mt-2 text-xs text-white/50">Average competition score · {Math.round(summary.averageRelevancy * 100)}% relevancy</p>
          </div>
        </div>
      </div>
    </section>
  );
}

