import { useRankBeamStore } from "@/lib/state";
import { ArrowUpRight } from "lucide-react";

export function KeywordTable() {
  const { keywordInsights } = useRankBeamStore();

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
        <div className="mt-8 overflow-hidden rounded-3xl border border-white/10 bg-black/40">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-white/5 text-xs uppercase tracking-wide text-white/60">
              <tr>
                <th className="px-6 py-4">Keyword</th>
                <th className="px-6 py-4">Search volume</th>
                <th className="px-6 py-4">Competition</th>
                <th className="px-6 py-4">Relevancy</th>
                <th className="px-6 py-4">Title density</th>
              </tr>
            </thead>
            <tbody>
              {keywordInsights.map((row, index) => (
                <tr key={row.keyword} className="border-t border-white/5 text-white/80">
                  <td className="px-6 py-4 font-medium text-white">
                    <span className="mr-2 text-white/40">{String(index + 1).padStart(2, "0")}</span>
                    {row.keyword}
                  </td>
                  <td className="px-6 py-4">{row.searchVolume.toLocaleString()}</td>
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
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
