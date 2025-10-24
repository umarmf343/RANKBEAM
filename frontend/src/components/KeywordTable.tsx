import type { RankBeamState } from "@/lib/state";
import { useRankBeamStore } from "@/lib/state";
import { ArrowUpRight, Loader2 } from "lucide-react";
import { useMemo, useState } from "react";

type FilterId = "all" | "opportunity" | "launch" | "density";
type KeywordRow = RankBeamState["keywordInsights"][number];

const FILTERS: { id: FilterId; label: string; description: string; predicate: (row: KeywordRow) => boolean }[] = [
  {
    id: "all",
    label: "All keywords",
    description: "Entire dataset sorted by search demand",
    predicate: () => true
  },
  {
    id: "opportunity",
    label: "Low competition",
    description: "Competition score under 3.2",
    predicate: (row) => row.competitionScore < 3.2
  },
  {
    id: "launch",
    label: "Launch ready",
    description: "Relevancy above 0.85",
    predicate: (row) => row.relevancyScore >= 0.85
  },
  {
    id: "density",
    label: "Title gap",
    description: "Fewer than 3 exact title matches",
    predicate: (row) => row.titleDensity < 3
  }
];

export function KeywordTable() {
  const { keywordInsights, loading } = useRankBeamStore();
  const [activeFilter, setActiveFilter] = useState<FilterId>("all");

  const filteredRows = useMemo(() => {
    const filter = FILTERS.find((option) => option.id === activeFilter) ?? FILTERS[0];
    const rows = keywordInsights.filter(filter.predicate);
    return rows.sort((a, b) => b.searchVolume - a.searchVolume);
  }, [activeFilter, keywordInsights]);

  const maxSearchVolume = useMemo(() => {
    return filteredRows.reduce((max, row) => Math.max(max, row.searchVolume), 0);
  }, [filteredRows]);

  const opportunityScore = (row: KeywordRow) => {
    if (maxSearchVolume === 0) return 0;
    const volumeWeight = row.searchVolume / maxSearchVolume;
    const competitionWeight = 1 - Math.min(row.competitionScore / 10, 0.95);
    const relevancyWeight = row.relevancyScore;
    const densityWeight = 1 / (1 + row.titleDensity / 10);
    return Math.round(volumeWeight * competitionWeight * relevancyWeight * densityWeight * 100);
  };

  return (
    <section className="border-b border-white/5 bg-night" id="platform">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">Opportunity scorecard</h2>
            <p className="mt-2 max-w-2xl text-sm text-white/70">
              RankBeam combines deterministic scraping fallbacks with AI-powered clustering to surface keyword themes with
              immediate monetisation potential. Highlight long-tail phrases with high volume, low title density and strong
              relevancy to your catalog.
            </p>
          </div>
          <a
            href="#cta"
            className="hidden items-center gap-2 rounded-full border border-aurora-500/30 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-200 transition hover:bg-aurora-500/10 md:flex"
          >
            Export CSV <ArrowUpRight className="h-4 w-4" />
          </a>
        </div>
        <div className="mt-8 flex flex-wrap gap-3">
          {FILTERS.map((filter) => {
            const isActive = filter.id === activeFilter;
            return (
              <button
                key={filter.id}
                type="button"
                onClick={() => setActiveFilter(filter.id)}
                className={`rounded-full border px-4 py-2 text-xs font-semibold uppercase tracking-wide transition ${
                  isActive
                    ? "border-aurora-400 bg-aurora-500/10 text-aurora-100"
                    : "border-white/10 text-white/60 hover:border-aurora-400 hover:text-white"
                }`}
              >
                {filter.label}
              </button>
            );
          })}
        </div>
        <p className="mt-3 text-xs text-white/50">
          {FILTERS.find((option) => option.id === activeFilter)?.description}
        </p>
        <div className="mt-8 overflow-hidden rounded-3xl border border-white/10 bg-black/40">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-white/60">
              <tr>
                <th className="px-6 py-4">Keyword</th>
                <th className="px-6 py-4">Search volume</th>
                <th className="px-6 py-4">Competition</th>
                <th className="px-6 py-4">Relevancy</th>
                <th className="px-6 py-4">Title density</th>
                <th className="px-6 py-4">Opportunity</th>
              </tr>
            </thead>
            <tbody>
              {filteredRows.map((row, index) => (
                <tr key={row.keyword} className="border-t border-white/5 text-white/80">
                  <td className="px-6 py-4 font-medium text-white">
                    <span className="mr-2 text-white/40">{String(index + 1).padStart(2, "0")}</span>
                    {row.keyword}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-col">
                      <span>{row.searchVolume.toLocaleString()}</span>
                      <span className="mt-1 h-1.5 rounded-full bg-white/10">
                        <span
                          className="block h-full rounded-full bg-gradient-to-r from-aurora-400 to-aurora-600"
                          style={{ width: `${(row.searchVolume / Math.max(maxSearchVolume, 1)) * 100}%` }}
                        />
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex rounded-full bg-aurora-500/10 px-3 py-1 text-xs font-semibold text-aurora-200">
                      {row.competitionScore.toFixed(2)}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex items-center gap-2">
                      <span className="h-2.5 w-2.5 rounded-full bg-aurora-400" />
                      {(row.relevancyScore * 100).toFixed(0)}%
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    {row.titleDensity.toFixed(0)}
                    <span className="ml-2 text-xs text-white/50">titles</span>
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex items-center gap-2 rounded-full border border-white/10 px-3 py-1 text-xs font-semibold text-white/80">
                      {opportunityScore(row)} / 100
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {loading && (
          <div className="flex items-center gap-2 pt-4 text-xs text-white/50">
            <Loader2 className="h-4 w-4 animate-spin" /> Refreshing insights…
          </div>
        )}
      </div>
    </section>
  );
}
