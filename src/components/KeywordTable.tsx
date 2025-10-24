import { useRankBeamStore } from "@/lib/state";
import { AlertTriangle, ArrowUpRight, Filter } from "lucide-react";
import { Link } from "react-router-dom";
import { useMemo, useState } from "react";

export function KeywordTable() {
  const { keywordInsights, suggestedKeywords, dataSource, lastUpdated, error } = useRankBeamStore();
  const [minVolume, setMinVolume] = useState(300);
  const [maxCompetition, setMaxCompetition] = useState(6);
  const [maxTitleDensity, setMaxTitleDensity] = useState(25);
  const [focusOpportunities, setFocusOpportunities] = useState(true);

  const sortedInsights = useMemo(() => {
    const base = [...keywordInsights].sort((a, b) => {
      if (b.searchVolume === a.searchVolume) {
        return b.relevancyScore - a.relevancyScore;
      }
      return b.searchVolume - a.searchVolume;
    });
    return base;
  }, [keywordInsights]);

  const filteredInsights = useMemo(() => {
    return sortedInsights.filter((row) => {
      if (row.searchVolume < minVolume) return false;
      if (row.competitionScore > maxCompetition) return false;
      if (row.titleDensity > maxTitleDensity) return false;
      if (focusOpportunities && row.relevancyScore < 0.6) return false;
      return true;
    });
  }, [sortedInsights, minVolume, maxCompetition, maxTitleDensity, focusOpportunities]);

  const topOpportunity = filteredInsights[0];

  return (
    <section className="border-b border-white/5 bg-night" id="platform">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">Opportunity scorecard</h2>
            <p className="mt-2 max-w-2xl text-sm text-white/70">
              RankBeam scrapes live Amazon search results and applies clustering to surface keyword themes with immediate
              monetisation potential. Highlight long-tail phrases with high volume, low title density and strong relevancy to
              your catalog.
            </p>
          </div>
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2 text-xs uppercase tracking-wide text-white/50">
              <Filter className="h-3.5 w-3.5" /> Dynamic filters
            </div>
            <div className="grid gap-3 text-xs text-white/60 sm:grid-cols-2">
              <label className="flex flex-col gap-2">
                <span>Min search volume: {minVolume.toLocaleString()}</span>
                <input
                  type="range"
                  min={50}
                  max={4000}
                  step={50}
                  value={minVolume}
                  onChange={(event) => setMinVolume(Number(event.target.value))}
                  className="accent-aurora-400"
                />
              </label>
              <label className="flex flex-col gap-2">
                <span>Max competition: {maxCompetition.toFixed(1)}</span>
                <input
                  type="range"
                  min={1}
                  max={10}
                  step={0.1}
                  value={maxCompetition}
                  onChange={(event) => setMaxCompetition(Number(event.target.value))}
                  className="accent-aurora-400"
                />
              </label>
              <label className="flex flex-col gap-2">
                <span>Max title density: {maxTitleDensity}</span>
                <input
                  type="range"
                  min={2}
                  max={60}
                  step={1}
                  value={maxTitleDensity}
                  onChange={(event) => setMaxTitleDensity(Number(event.target.value))}
                  className="accent-aurora-400"
                />
              </label>
              <label className="flex items-center gap-2 rounded-full border border-white/10 bg-black/30 px-4 py-2 text-xs text-white/70">
                <input
                  type="checkbox"
                  checked={focusOpportunities}
                  onChange={(event) => setFocusOpportunities(event.target.checked)}
                  className="h-4 w-4 rounded border-white/30 bg-night text-aurora-400 focus:ring-aurora-400"
                />
                Prioritise high relevancy matches
              </label>
            </div>
            <Link
              to="/#cta"
              className="inline-flex items-center gap-2 rounded-full border border-aurora-500/30 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-200 transition hover:bg-aurora-500/10"
            >
              Export CSV <ArrowUpRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
        <div className="mt-4 flex flex-col gap-2 text-xs text-white/50 sm:flex-row sm:items-center sm:justify-between">
          <span>
            Source: <span className="font-semibold text-white/80">{dataSource === "scraped" ? "Live Amazon scrape" : "Unavailable"}</span>
          </span>
          {lastUpdated && (
            <span>
              Last updated <time dateTime={lastUpdated}>{new Date(lastUpdated).toLocaleString()}</time>
            </span>
          )}
        </div>
        {dataSource === "error" && (
          <div className="mt-4 flex items-center gap-3 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-200">
            <AlertTriangle className="h-5 w-5 flex-shrink-0" aria-hidden />
            <div>
              <p className="font-semibold">Unable to retrieve live keyword intelligence.</p>
              <p className="text-rose-100/70">{error ?? "Amazon scraping failed for the current seed keyword. Please try again."}</p>
            </div>
          </div>
        )}
        <div className="mt-8 overflow-hidden rounded-3xl border border-white/10 bg-black/40">
          {dataSource === "error" ? (
            <div className="px-6 py-16 text-center text-sm text-rose-200">
              We could not load live data from Amazon. Update the seed keyword or try again later.
            </div>
          ) : filteredInsights.length > 0 ? (
            <table className="min-w-full text-left text-sm">
              <thead className="bg-white/5 text-xs uppercase tracking-wide text-white/60">
                <tr>
                  <th className="px-6 py-4">Keyword</th>
                  <th className="px-6 py-4">Search volume</th>
                  <th className="px-6 py-4">Competitors</th>
                  <th className="px-6 py-4">Avg reviews</th>
                  <th className="px-6 py-4">Avg price</th>
                  <th className="px-6 py-4">Title density</th>
                  <th className="px-6 py-4">Avg age</th>
                  <th className="px-6 py-4">Opportunity score</th>
                  <th className="px-6 py-4">Demand score</th>
                  <th className="px-6 py-4">Competition score</th>
                </tr>
              </thead>
              <tbody>
                {filteredInsights.map((row, index) => {
                  const isLeader = topOpportunity?.keyword === row.keyword;
                  return (
                    <tr
                      key={row.keyword}
                      className={`border-t border-white/5 text-white/80 transition ${
                        isLeader ? "bg-aurora-500/5" : index % 2 === 1 ? "bg-white/[0.04]" : ""
                      }`}
                    >
                      <td className="px-6 py-4 font-medium text-white">
                        <span className="mr-2 text-white/40">{String(index + 1).padStart(2, "0")}</span>
                        {row.keyword}
                        {isLeader && <span className="ml-2 rounded-full bg-aurora-500/20 px-2 py-0.5 text-xs text-aurora-100">Top pick</span>}
                      </td>
                      <td className="px-6 py-4">{row.searchVolume.toLocaleString()}</td>
                      <td className="px-6 py-4">{row.competitors.toLocaleString()}</td>
                      <td className="px-6 py-4">{row.avgReviews.toLocaleString()}</td>
                      <td className="px-6 py-4">${row.avgPrice.toFixed(2)}</td>
                      <td className="px-6 py-4">
                        {row.titleDensity.toFixed(0)}
                        <span className="ml-2 text-xs text-white/50">titles</span>
                      </td>
                      <td className="px-6 py-4">{row.avgAge} months</td>
                      <td className="px-6 py-4">
                        <span className="inline-flex items-center gap-2 rounded-full bg-white/5 px-3 py-1 text-xs font-semibold text-white">
                          {row.opportunityScore}
                          {isLeader && <span className="text-aurora-300">▲</span>}
                        </span>
                      </td>
                      <td className="px-6 py-4">{row.demandScore}</td>
                      <td className="px-6 py-4">{Math.round(row.competitionScore * 10)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          ) : (
            <div className="px-6 py-16 text-center text-sm text-white/60">
              No keyword insights yet. Try updating the seed keyword above to generate fresh opportunities.
            </div>
          )}
        </div>
        {suggestedKeywords.length > 0 && (
          <div className="mt-8 rounded-3xl border border-white/10 bg-black/40 p-6">
            <h3 className="text-sm font-semibold uppercase tracking-wide text-white/70">Suggested keywords</h3>
            <div className="mt-4 flex flex-wrap gap-2 text-xs">
              {suggestedKeywords.map((keyword) => (
                <span key={keyword} className="rounded-full bg-white/10 px-3 py-1 text-white/70">
                  {keyword}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
