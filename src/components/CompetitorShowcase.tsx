import { useRankBeamStore, type RankBeamState } from "@/lib/state";
import {
  BarChart3,
  Download,
  ExternalLink,
  Filter,
  HelpCircle,
  Search,
  Sparkles,
  Star,
  TrendingUp
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState, type KeyboardEvent, type ReactNode } from "react";

type KeywordInsight = RankBeamState["keywordInsights"][number];

const COMPETITOR_SORT_OPTIONS = [
  { label: "Best seller rank", value: "rank" as const },
  { label: "Review count", value: "reviews" as const },
  { label: "Rating", value: "rating" as const }
];

type CompetitorSort = (typeof COMPETITOR_SORT_OPTIONS)[number]["value"];

const KEYWORD_SORT_OPTIONS = [
  { label: "Search volume", value: "volume" as const },
  { label: "Opportunity score", value: "opportunity" as const },
  { label: "Lowest competition", value: "competition" as const },
  { label: "Fewest competitors", value: "competitors" as const }
];

type KeywordSort = (typeof KEYWORD_SORT_OPTIONS)[number]["value"];

function normalisePrice(price: string): number | undefined {
  const digits = price.replace(/[^0-9.,-]/g, "").replace(/,/g, "");
  if (!digits) return undefined;
  const parsed = Number.parseFloat(digits);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function deriveCurrencySymbol(price: string): string {
  const match = price.match(/[^0-9.,\s-]/);
  return match ? match[0] : "$";
}

function buildWordCloudData(keywords: KeywordInsight[]) {
  const topKeywords = keywords.slice(0, 60);
  const maxVolume = Math.max(1, ...topKeywords.map((row) => row.searchVolume));
  return topKeywords.map((row) => ({
    keyword: row.keyword,
    weight: Math.max(0.6, row.searchVolume / maxVolume),
    opportunity: row.opportunityScore,
    competition: row.competitionScore
  }));
}

function StatPill({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs text-white/80">
      {icon}
      <span className="font-semibold text-white">{value}</span>
      <span className="text-white/50">{label}</span>
    </span>
  );
}
export function CompetitorShowcase() {
  const {
    competitors,
    country,
    keyword,
    updateKeyword,
    refresh,
    loading,
    keywordInsights
  } = useRankBeamStore();

  const [localKeyword, setLocalKeyword] = useState(keyword);
  const [activeKeyword, setActiveKeyword] = useState(keyword);
  const [sortBy, setSortBy] = useState<CompetitorSort>("rank");
  const [keywordSearch, setKeywordSearch] = useState("");
  const [keywordSort, setKeywordSort] = useState<KeywordSort>("volume");
  const [hasTriggeredSearch, setHasTriggeredSearch] = useState(false);

  useEffect(() => {
    setLocalKeyword(keyword);
    setActiveKeyword(keyword);
  }, [keyword]);

  const handleSearch = useCallback(() => {
    const trimmedKeyword = localKeyword.trim();
    if (!trimmedKeyword) return;
    setHasTriggeredSearch(true);
    setActiveKeyword(trimmedKeyword);
    updateKeyword(trimmedKeyword);
    refresh();
  }, [localKeyword, refresh, updateKeyword, setHasTriggeredSearch]);

  const handleKeyDown = useCallback(
    (event: KeyboardEvent<HTMLInputElement>) => {
      if (event.key === "Enter") {
        event.preventDefault();
        handleSearch();
      }
    },
    [handleSearch]
  );

  const sortedCompetitors = useMemo(() => {
    const sorted = [...competitors];
    sorted.sort((a, b) => {
      switch (sortBy) {
        case "reviews":
          return b.reviewCount - a.reviewCount;
        case "rating":
          return b.rating - a.rating;
        case "rank":
        default:
          return a.rank - b.rank;
      }
    });
    return sorted;
  }, [competitors, sortBy]);

  const summary = useMemo(() => {
    if (sortedCompetitors.length === 0) {
      return { avgRating: 0, avgReviews: 0, avgPrice: undefined as number | undefined, currencySymbol: "$" };
    }

    const ratingTotal = sortedCompetitors.reduce((sum, competitor) => sum + (competitor.rating || 0), 0);
    const reviewTotal = sortedCompetitors.reduce((sum, competitor) => sum + (competitor.reviewCount || 0), 0);

    let priceSum = 0;
    let priceCount = 0;
    let currencySymbol = "$";

    sortedCompetitors.forEach((competitor) => {
      const value = normalisePrice(competitor.price);
      if (value !== undefined) {
        priceSum += value;
        priceCount += 1;
        if (currencySymbol === "$") {
          currencySymbol = deriveCurrencySymbol(competitor.price) || currencySymbol;
        }
      }
    });

    const avgRating = ratingTotal / sortedCompetitors.length;
    const avgReviews = reviewTotal / sortedCompetitors.length;
    const avgPrice = priceCount > 0 ? priceSum / priceCount : undefined;

    return { avgRating, avgReviews, avgPrice, currencySymbol };
  }, [sortedCompetitors]);

  const handleExport = useCallback(() => {
    if (sortedCompetitors.length === 0) return;
    const rows = [
      ["Rank", "Title", "ASIN", "Price", "Rating", "Reviews", "Best Seller Rank", "URL"],
      ...sortedCompetitors.map((competitor) => [
        competitor.rank,
        competitor.title,
        competitor.asin,
        competitor.price,
        competitor.rating.toFixed(1),
        competitor.reviewCount,
        competitor.bestSellerRank,
        competitor.url
      ])
    ];

    const csv = rows
      .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(","))
      .join("\n");

    const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `rankbeam-competitors-${activeKeyword.replace(/\s+/g, "-")}.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  }, [sortedCompetitors, activeKeyword]);

  const keywordWordCloud = useMemo(() => buildWordCloudData(keywordInsights), [keywordInsights]);

  const keywordTable = useMemo(() => {
    const filtered = keywordInsights.filter((row) =>
      row.keyword.toLowerCase().includes(keywordSearch.toLowerCase())
    );
    const sorted = [...filtered].sort((a, b) => {
      switch (keywordSort) {
        case "opportunity":
          return b.opportunityScore - a.opportunityScore;
        case "competition":
          return a.competitionScore - b.competitionScore;
        case "competitors":
          return a.competitors - b.competitors;
        case "volume":
        default:
          return b.searchVolume - a.searchVolume;
      }
    });
    return sorted;
  }, [keywordInsights, keywordSearch, keywordSort]);

  const showLoading = hasTriggeredSearch && loading;
  const statusIconClass = showLoading ? "animate-spin" : "text-aurora-400";
  const statusText = showLoading
    ? "Scanning competitor landscape…"
    : hasTriggeredSearch
      ? "Updated competitor intelligence"
      : "Ready to scan competitor landscape";

  return (
    <section id="competitors" className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-2xl space-y-4">
            <div>
              <h2 className="font-display text-3xl font-semibold text-white">Competitor analysis hub</h2>
              <p className="mt-2 text-sm text-white/70">
                Diagnose market leaders, benchmark your positioning, and surface keyword strategies that fuel their sales.
                RankBeam now streams live catalogue data directly from Amazon's {country.label} marketplace whenever you run a
                search.
              </p>
            </div>
            <span className="inline-flex items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs uppercase tracking-wide text-white/60">
              Marketplace: {country.label}
            </span>
          </div>
          <div className="w-full max-w-md space-y-4 rounded-3xl border border-white/10 bg-black/40 p-6 shadow-[0_40px_120px_-60px_rgba(76,102,241,0.4)]">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-white/60" htmlFor="competitor-keyword">
                Seed keyword search
              </label>
              <div className="mt-3 flex flex-col gap-3 sm:flex-row">
                <div className="relative flex-1">
                  <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
                  <input
                    id="competitor-keyword"
                    value={localKeyword}
                    onChange={(event) => setLocalKeyword(event.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Search competitors by keyword"
                    className="w-full rounded-full border border-white/10 bg-night/60 px-10 py-3 text-sm text-white placeholder:text-white/40 focus:border-aurora-400 focus:outline-none"
                  />
                </div>
                <button
                  type="button"
                  onClick={handleSearch}
                  disabled={loading || localKeyword.trim().length === 0}
                  className="rounded-full bg-aurora-500 px-6 py-3 text-sm font-semibold text-night transition hover:bg-aurora-400 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Go
                </button>
              </div>
            </div>
            <div className="flex items-center gap-2 text-xs text-white/60" aria-live="polite">
              <Sparkles className={`h-4 w-4 ${statusIconClass}`} />
              {statusText}
            </div>
          </div>
        </div>

        {activeKeyword && (
          <div className="mt-12 space-y-6 rounded-3xl border border-white/5 bg-black/40 p-8">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-aurora-200">RankBeam key capabilities</p>
                <h3 className="mt-1 text-2xl font-semibold text-white">Competitive intelligence for "{activeKeyword}"</h3>
              </div>
              <div className="flex flex-wrap gap-2 text-xs text-white/70">
                <StatPill icon={<Star className="h-4 w-4 text-aurora-300" />} label="avg. rating" value={summary.avgRating ? `${summary.avgRating.toFixed(1)}★` : "N/A"} />
                <StatPill icon={<BarChart3 className="h-4 w-4 text-aurora-300" />} label="avg. reviews" value={Math.round(summary.avgReviews).toLocaleString()} />
                {summary.avgPrice !== undefined && (
                  <StatPill
                    icon={<TrendingUp className="h-4 w-4 text-aurora-300" />}
                    label="avg. price"
                    value={`${summary.currencySymbol}${summary.avgPrice.toFixed(2)}`}
                  />
                )}
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              {[
                {
                  title: "Real-time Amazon listings",
                  description:
                    "Every search triggers a live scrape so you can evaluate pricing, review velocity and catalogue coverage without stale mock data.",
                  icon: <TrendingUp className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Keyword ranking insights",
                  description:
                    "Spot high-performing keywords, evaluate competition and cross-reference with your own listing for coverage gaps.",
                  icon: <Search className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Competitor benchmarks",
                  description:
                    "Compare live star ratings, review totals and bestseller positioning to understand the strength of each listing.",
                  icon: <Star className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Organised exports",
                  description:
                    "Download a clean CSV of the live data set for deeper modelling, outreach or sharing with your team.",
                  icon: <Filter className="h-5 w-5 text-aurora-300" />
                }
              ].map((feature) => (
                <article key={feature.title} className="rounded-2xl border border-white/10 bg-night/60 p-5">
                  <div className="flex items-center gap-3 text-white">
                    {feature.icon}
                    <h4 className="text-sm font-semibold">{feature.title}</h4>
                  </div>
                  <p className="mt-2 text-sm text-white/70">{feature.description}</p>
                </article>
              ))}
            </div>
          </div>
        )}

        <div className="mt-12 flex flex-wrap items-center gap-3 rounded-3xl border border-white/10 bg-black/30 px-5 py-4 text-xs text-white/70">
          <div className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-aurora-300" />
            <span className="uppercase tracking-wide">Sort by</span>
            <select
              value={sortBy}
              onChange={(event) => setSortBy(event.target.value as CompetitorSort)}
              className="rounded-full border border-white/10 bg-black/60 px-3 py-1.5 text-white/70 focus:border-aurora-400 focus:outline-none"
            >
              {COMPETITOR_SORT_OPTIONS.map((option) => (
                <option key={option.value} value={option.value} className="bg-night">
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <button
              type="button"
              onClick={handleExport}
              className="inline-flex items-center gap-2 rounded-full border border-aurora-400 px-4 py-2 text-xs font-semibold text-aurora-200 transition hover:bg-aurora-500/10"
            >
              Export CSV
              <Download className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
        <div className="mt-8 grid gap-6 md:grid-cols-2">
          {sortedCompetitors.map((competitor) => (
            <article key={competitor.asin} className="flex flex-col gap-5 rounded-3xl border border-white/10 bg-black/40 p-6">
              <div className="flex flex-col gap-4 sm:flex-row">
                <div className="relative h-48 w-full overflow-hidden rounded-2xl sm:w-40">
                  <img src={competitor.cover} alt={`${competitor.title} cover art`} className="h-full w-full object-cover" />
                  {competitor.isIndie && (
                    <span className="absolute left-3 top-3 rounded-full bg-emerald-500/80 px-3 py-1 text-xs font-semibold text-emerald-50">
                      Indie spotlight
                    </span>
                  )}
                </div>
                <div className="flex flex-1 flex-col gap-4 text-sm text-white/80">
                  <div>
                    <h3 className="text-lg font-semibold text-white">{competitor.title}</h3>
                    <p className="text-xs text-white/50">ASIN {competitor.asin}</p>
                  </div>
                  <div className="flex flex-wrap gap-2 text-xs text-white/60">
                    <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">Rank #{competitor.rank}</span>
                    <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">{competitor.bestSellerRank}</span>
                  </div>
                  <dl className="grid gap-3 text-xs text-white/70 sm:grid-cols-2">
                    <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                      <dt className="font-semibold text-white">Price</dt>
                      <dd className="mt-2 text-lg font-semibold text-white">
                        {competitor.price === "N/A" ? "Not available" : competitor.price}
                      </dd>
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                      <dt className="font-semibold text-white">Rating</dt>
                      <dd className="mt-2 text-lg font-semibold text-white">
                        {competitor.rating > 0 ? `${competitor.rating.toFixed(1)}★` : "Not available"}
                      </dd>
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                      <dt className="font-semibold text-white">Reviews</dt>
                      <dd className="mt-2 text-lg font-semibold text-white">{competitor.reviewCount.toLocaleString()}</dd>
                    </div>
                    <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                      <dt className="font-semibold text-white">Listing</dt>
                      <dd className="mt-2 text-sm text-white/80">Live data retrieved directly from Amazon during this session.</dd>
                    </div>
                  </dl>
                  <a
                    href={competitor.url}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex w-max items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-200 transition hover:border-aurora-400"
                  >
                    View listing <ExternalLink className="h-3.5 w-3.5" />
                  </a>
                </div>
              </div>
            </article>
          ))}
        </div>
        

        <section className="mt-12 space-y-6 rounded-3xl border border-white/10 bg-black/40 p-8">
          <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h3 className="text-2xl font-semibold text-white">Comprehensive keyword data</h3>
              <p className="mt-1 text-sm text-white/70">
                RankBeam returns hundreds of keywords related to "{activeKeyword}" with search volume, opportunity, competition
                and market density metrics. Filter and sort to assemble your launch roadmap.
              </p>
            </div>
            <div className="flex flex-col gap-2 text-xs text-white/60 sm:flex-row sm:items-center">
              <label className="flex items-center gap-2 rounded-full border border-white/10 bg-black/40 px-3 py-2">
                <Search className="h-4 w-4 text-white/40" />
                <input
                  value={keywordSearch}
                  onChange={(event) => setKeywordSearch(event.target.value)}
                  placeholder="Filter keywords"
                  className="bg-transparent text-sm text-white placeholder:text-white/40 focus:outline-none"
                />
              </label>
              <select
                value={keywordSort}
                onChange={(event) => setKeywordSort(event.target.value as KeywordSort)}
                className="rounded-full border border-white/10 bg-black/60 px-3 py-2 text-white/70 focus:border-aurora-400 focus:outline-none"
              >
                {KEYWORD_SORT_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value} className="bg-night">
                    Sort by {option.label.toLowerCase()}
                  </option>
                ))}
              </select>
            </div>
          </header>
          <div className="overflow-hidden rounded-2xl border border-white/10">
            {keywordTable.length > 0 ? (
              <table className="min-w-full text-left text-sm text-white/80">
                <thead className="bg-white/5 text-xs uppercase tracking-wide text-white/60">
                  <tr>
                    <th className="px-6 py-3">Keyword</th>
                    <th className="px-6 py-3">Search volume</th>
                    <th className="px-6 py-3">Opportunity</th>
                    <th className="px-6 py-3">Competition</th>
                    <th className="px-6 py-3">Competitors</th>
                    <th className="px-6 py-3">Avg reviews</th>
                    <th className="px-6 py-3">Avg price</th>
                    <th className="px-6 py-3">Avg age</th>
                  </tr>
                </thead>
                <tbody>
                  {keywordTable.slice(0, 1000).map((row, index) => (
                    <tr key={`${row.keyword}-${index}`} className="border-t border-white/5">
                      <td className="px-6 py-3 font-semibold text-white">{row.keyword}</td>
                      <td className="px-6 py-3">{row.searchVolume.toLocaleString()}</td>
                      <td className="px-6 py-3">{row.opportunityScore}</td>
                      <td className="px-6 py-3">{row.competitionScore.toFixed(2)}</td>
                      <td className="px-6 py-3">{row.competitors.toLocaleString()}</td>
                      <td className="px-6 py-3">{row.avgReviews.toLocaleString()}</td>
                      <td className="px-6 py-3">${row.avgPrice.toFixed(2)}</td>
                      <td className="px-6 py-3">{row.avgAge} months</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="px-6 py-10 text-sm text-white/60">Search for a seed keyword above to load keyword intelligence.</div>
            )}
          </div>
          <div className="rounded-2xl border border-white/10 bg-black/40 p-6">
            <h4 className="text-sm font-semibold uppercase tracking-wide text-white/70">Word cloud visualisation</h4>
            <div className="mt-4 flex flex-wrap gap-4">
              {keywordWordCloud.map((entry) => (
                <span
                  key={entry.keyword}
                  className="rounded-full bg-white/5 px-4 py-2 text-white/80"
                  style={{
                    fontSize: `${entry.weight * 1.2}rem`,
                    opacity: Math.max(0.45, entry.weight)
                  }}
                >
                  {entry.keyword}
                </span>
              ))}
              {keywordWordCloud.length === 0 && (
                <p className="text-sm text-white/60">Keyword cloud will populate after your first search.</p>
              )}
            </div>
          </div>
        </section>

        <section className="mt-12 grid gap-6 lg:grid-cols-2">
          <div className="space-y-4 rounded-3xl border border-aurora-500/30 bg-aurora-500/10 p-6 text-white">
            <h3 className="text-xl font-semibold">Unlock unlimited searches</h3>
            <p className="text-sm text-white/80">
              Publisher and Publisher Pro subscribers can run unlimited live competitor scans, export full data sets and access
              AI-powered listing blueprints.
            </p>
            <div className="grid gap-3 text-sm">
              <div className="flex items-start gap-2">
                <Sparkles className="mt-0.5 h-4 w-4 text-aurora-200" />
                <span>Unlimited seed keyword lookups</span>
              </div>
              <div className="flex items-start gap-2">
                <BarChart3 className="mt-0.5 h-4 w-4 text-aurora-200" />
                <span>Downloadable CSV, Excel and PDF exports</span>
              </div>
              <div className="flex items-start gap-2">
                <TrendingUp className="mt-0.5 h-4 w-4 text-aurora-200" />
                <span>Advanced forecasting dashboards and alerts</span>
              </div>
            </div>
            <button
              type="button"
              className="inline-flex w-max items-center gap-2 rounded-full bg-aurora-500 px-6 py-3 text-sm font-semibold text-night transition hover:bg-aurora-400"
            >
              Explore plans
            </button>
          </div>
          <div className="space-y-4 rounded-3xl border border-white/10 bg-black/40 p-6 text-white/80">
            <h3 className="text-xl font-semibold text-white">Help & FAQs</h3>
            <dl className="space-y-3 text-sm">
              <div>
                <dt className="font-semibold text-white">How often is competitor data refreshed?</dt>
                <dd className="mt-1 text-white/70">
                  Live scrapes are triggered whenever you run a search, pulling the latest pricing, review counts and bestseller rankings straight from Amazon.
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-white">Can I export the data?</dt>
                <dd className="mt-1 text-white/70">
                  Use the export controls above to download CSV files. Publisher Pro subscribers also unlock Excel and PDF exports directly from this dashboard.
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-white">What does the opportunity score represent?</dt>
                <dd className="mt-1 text-white/70">
                  Opportunity combines demand, competition, and review density to highlight keywords where new titles can realistically rank.
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-white">Need more help?</dt>
                <dd className="mt-1 flex items-center gap-2 text-white/70">
                  <HelpCircle className="h-4 w-4 text-aurora-200" /> Visit the knowledge base or message our strategy team for hands-on guidance.
                </dd>
              </div>
            </dl>
          </div>
        </section>
      </div>
    </section>
  );
}
