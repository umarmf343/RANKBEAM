import { useRankBeamStore, type RankBeamState } from "@/lib/state";
import { stableFloat } from "@/lib/keywordEngine";
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
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type KeyboardEvent,
  type ReactNode
} from "react";

type CompetitorRecord = RankBeamState["competitors"][number];
type KeywordInsight = RankBeamState["keywordInsights"][number];

type Format = "Kindle" | "Paperback" | "Audiobook";

type ReviewEntry = {
  id: string;
  reviewer: string;
  rating: number;
  date: Date;
  headline: string;
  body: string;
};

type KeywordRanking = {
  keyword: string;
  searchVolume: number;
  competitionScore: number;
  opportunityScore: number;
  rankPosition: number;
  trafficShare: number;
};

type ReverseKeyword = {
  keyword: string;
  searchVolume: number;
  opportunityScore: number;
  competitionLevel: number;
  avgReviews: number;
  relevance: number;
};

type EnrichedCompetitor = CompetitorRecord & {
  format: Format;
  priceValue: number;
  currencySymbol: string;
  salesSeries: number[];
  royaltySeries: number[];
  rankSeries: number[];
  reviewSeries: number[];
  keywords: KeywordRanking[];
  reviews: ReviewEntry[];
  salesMomentum: number;
  rankMomentum: number;
  reviewMomentum: number;
  monthlySales: number;
  monthlyRoyalties: number;
};

const FORMATS: Format[] = ["Kindle", "Paperback", "Audiobook"];

const TIME_RANGES = [
  { label: "Last 30 days", value: 30 },
  { label: "Last 90 days", value: 90 },
  { label: "Last 180 days", value: 180 }
];

const REVIEW_WINDOWS = [
  { label: "30 days", value: 30 },
  { label: "90 days", value: 90 },
  { label: "180 days", value: 180 },
  { label: "1 year", value: 365 }
];

const RATING_FILTERS = [
  { label: "All", value: 0 },
  { label: "5★", value: 5 },
  { label: "4★", value: 4 },
  { label: "3★", value: 3 },
  { label: "2★", value: 2 },
  { label: "1★", value: 1 }
];

const COMPETITOR_SORT_OPTIONS = [
  { label: "Best seller rank", value: "rank" as const },
  { label: "Review count", value: "reviews" as const },
  { label: "Estimated sales", value: "sales" as const },
  { label: "Estimated royalties", value: "royalties" as const }
];

type CompetitorSort = (typeof COMPETITOR_SORT_OPTIONS)[number]["value"];

const KEYWORD_SORT_OPTIONS = [
  { label: "Search volume", value: "volume" as const },
  { label: "Opportunity score", value: "opportunity" as const },
  { label: "Lowest competition", value: "competition" as const },
  { label: "Fewest competitors", value: "competitors" as const }
];

type KeywordSort = (typeof KEYWORD_SORT_OPTIONS)[number]["value"];

function normalisePrice(price: string): number {
  const digits = price.replace(/[^0-9.,-]/g, "").replace(/,/g, "");
  const parsed = Number.parseFloat(digits);
  return Number.isFinite(parsed) ? parsed : 0;
}

function buildSeries(seed: string, length: number, base: number, volatility: number, invert = false): number[] {
  const series: number[] = [];
  let previous = base;
  for (let index = 0; index < length; index += 1) {
    const noise = stableFloat(`${seed}-${index}`) * volatility * base;
    const direction = stableFloat(`${seed}-dir-${index}`) > 0.5 ? 1 : -1;
    const value = Math.max(0, previous + direction * noise);
    previous = value;
    series.push(invert ? Math.max(20, base * 2 - value) : value);
  }
  return series;
}

function percentageChange(series: number[]): number {
  if (series.length < 2) return 0;
  const first = series[0];
  const last = series[series.length - 1];
  if (first === 0) return last === 0 ? 0 : 100;
  return ((last - first) / first) * 100;
}

function average(series: number[]): number {
  if (series.length === 0) return 0;
  return series.reduce((total, value) => total + value, 0) / series.length;
}

function pickFormat(asin: string): Format {
  const index = Math.floor(stableFloat(`${asin}-format`) * FORMATS.length);
  return FORMATS[index] ?? "Kindle";
}

function buildReviewFeed(asin: string): ReviewEntry[] {
  const reviewers = ["Alex", "Jordan", "Morgan", "Taylor", "Priya", "Hannah", "Lee", "Sasha", "Devon", "Kai"];
  const headlines = [
    "Transformative insights",
    "Great for consistent publishing",
    "A solid framework",
    "Actionable tips",
    "Exactly what I needed",
    "Could go deeper",
    "Comprehensive and clear",
    "Worth every penny",
    "High-level overview",
    "Fantastic companion"
  ];
  const bodies = [
    "The strategy layout helped me reposition my series and increase conversions within weeks.",
    "Loved the chapter on automation. I implemented the workbook and saw immediate clarity.",
    "It's a thorough read with templates that make keyword planning less overwhelming.",
    "I appreciated the market analysis breakdown but wanted more real examples.",
    "Helped me rank for more competitive phrases after applying the launch checklist.",
    "Quality visuals and step-by-step instructions. Would recommend to fellow publishers.",
    "There are sections that feel repetitive but overall it's a smart investment.",
    "Review monitoring tips alone justify the purchase for my small imprint.",
    "Solid walkthrough of ads and organic ranking strategies for busy authors.",
    "A fantastic roundup of tactics with practical worksheets included."
  ];

  return Array.from({ length: 12 }).map((_, index) => {
    const rating = Math.min(5, Math.max(1, Math.round(3 + stableFloat(`${asin}-rating-${index}`) * 2)));
    const dayOffset = Math.round(stableFloat(`${asin}-day-${index}`) * 320);
    const date = new Date();
    date.setDate(date.getDate() - dayOffset);
    return {
      id: `${asin}-review-${index}`,
      reviewer: reviewers[index % reviewers.length],
      rating,
      date,
      headline: headlines[index % headlines.length],
      body: bodies[index % bodies.length]
    } satisfies ReviewEntry;
  });
}

function buildKeywordRankings(asin: string, seed: string, insights: KeywordInsight[]): KeywordRanking[] {
  const fallback = Array.from({ length: 12 }).map((_, index) => ({
    keyword: `${seed || "amazon publishing"} strategy ${index + 1}`,
    searchVolume: Math.round(400 + index * 45 + stableFloat(`${asin}-kw-${index}`) * 900),
    competitionScore: Number.parseFloat((3.2 + stableFloat(`${asin}-kwc-${index}`) * 3).toFixed(2)),
    opportunityScore: Math.round(50 + stableFloat(`${asin}-kwo-${index}`) * 45)
  }));

  const source = insights.length > 0 ? insights : fallback;

  return source.slice(0, 6).map((row, index) => {
    const base = stableFloat(`${asin}-${row.keyword}-${index}`);
    return {
      keyword: row.keyword,
      searchVolume: row.searchVolume,
      competitionScore: row.competitionScore,
      opportunityScore: row.opportunityScore,
      rankPosition: Math.max(1, Math.round(base * 18) + 1),
      trafficShare: Math.max(6, Math.round(stableFloat(`${asin}-${row.keyword}-traffic`) * 45))
    } satisfies KeywordRanking;
  });
}

function formatCurrency(amount: number, currencySymbol: string): string {
  return `${currencySymbol}${Math.round(amount).toLocaleString()}`;
}

function formatChange(value: number): string {
  if (!Number.isFinite(value) || Math.abs(value) < 0.05) return "0%";
  const rounded = Number.parseFloat(value.toFixed(1));
  return `${rounded > 0 ? "+" : ""}${rounded}%`;
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

function generateReverseResults(asin: string, insights: KeywordInsight[], seed: string): ReverseKeyword[] {
  const base = insights.length > 0 ? insights : Array.from({ length: 24 }).map((_, index) => ({
        keyword: `${seed || "amazon publishing"} niche ${index + 1}`,
        searchVolume: Math.round(450 + stableFloat(`${asin}-rv-${index}`) * 1400),
        opportunityScore: Math.round(55 + stableFloat(`${asin}-rv-opp-${index}`) * 40),
        competitionScore: Number.parseFloat((3.4 + stableFloat(`${asin}-rv-comp-${index}`) * 3).toFixed(2)),
        avgReviews: Math.round(35 + stableFloat(`${asin}-rv-rev-${index}`) * 220),
        relevancyScore: 0.6 + stableFloat(`${asin}-rv-rel-${index}`) * 0.35
      }));

  return base.slice(0, 30).map<ReverseKeyword>((row, index) => ({
    keyword: row.keyword,
    searchVolume: row.searchVolume,
    opportunityScore: row.opportunityScore,
    competitionLevel: row.competitionScore,
    avgReviews: row.avgReviews,
    relevance: Number.parseFloat(((row.relevancyScore ?? 0.7) * 100).toFixed(1)) + index % 2
  }));
}

function buildCompetitorProfile(
  competitor: CompetitorRecord,
  seed: string,
  insights: KeywordInsight[],
  days: number
): EnrichedCompetitor {
  const format = pickFormat(competitor.asin);
  const priceValue = normalisePrice(competitor.price);
  const currencySymbol = competitor.price.match(/[^0-9.,\s-]/)?.[0] ?? "$";
  const baseSales = 45 + stableFloat(`${competitor.asin}-sales-base`) * 220;
  const salesSeries = buildSeries(`${competitor.asin}-sales`, days, baseSales, 0.3);
  const royaltySeries = salesSeries.map((value) => value * priceValue * 0.6);
  const rankSeries = buildSeries(`${competitor.asin}-rank`, days, 3500 - competitor.rank * 120, 0.25, true).map((value) =>
    Math.max(25, Math.round(value))
  );
  const reviewSeries = buildSeries(`${competitor.asin}-reviews`, days, Math.max(4, competitor.reviewCount / 18), 0.4);
  const keywords = buildKeywordRankings(competitor.asin, seed, insights);
  const reviews = buildReviewFeed(competitor.asin).sort((a, b) => b.date.getTime() - a.date.getTime());
  const monthlySales = Math.round(average(salesSeries) * 30);
  const monthlyRoyalties = Math.round(average(royaltySeries));

  return {
    ...competitor,
    format,
    priceValue,
    currencySymbol,
    salesSeries,
    royaltySeries,
    rankSeries,
    reviewSeries,
    keywords,
    reviews,
    salesMomentum: percentageChange(salesSeries),
    rankMomentum: percentageChange(rankSeries) * -1,
    reviewMomentum: percentageChange(reviewSeries),
    monthlySales,
    monthlyRoyalties
  } satisfies EnrichedCompetitor;
}

function TrendSparkline({ series, color, ariaLabel }: { series: number[]; color: string; ariaLabel: string }) {
  if (series.length === 0) {
    return <div className="h-14" aria-hidden />;
  }

  const max = Math.max(...series);
  const min = Math.min(...series);
  const range = max - min || 1;
  const points = series
    .map((value, index) => {
      const x = (index / (series.length - 1 || 1)) * 100;
      const y = 100 - ((value - min) / range) * 100;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <svg viewBox="0 0 100 100" role="img" aria-label={ariaLabel} className="h-14 w-full overflow-visible" preserveAspectRatio="none">
      <polyline
        fill="none"
        stroke={color}
        strokeWidth={3}
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
      />
    </svg>
  );
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
  const [timeRange, setTimeRange] = useState(TIME_RANGES[0]);
  const [formatFilter, setFormatFilter] = useState<Format | "all">("all");
  const [sortBy, setSortBy] = useState<CompetitorSort>("rank");
  const [reviewWindow, setReviewWindow] = useState(REVIEW_WINDOWS[0]);
  const [ratingFilter, setRatingFilter] = useState<number>(0);
  const [reverseInput, setReverseInput] = useState("");
  const [reverseResults, setReverseResults] = useState<ReverseKeyword[]>([]);
  const [keywordSearch, setKeywordSearch] = useState("");
  const [keywordSort, setKeywordSort] = useState<KeywordSort>("volume");
  const [reverseError, setReverseError] = useState<string | undefined>();

  useEffect(() => {
    setLocalKeyword(keyword);
    setActiveKeyword(keyword);
  }, [keyword]);

  const handleSearch = useCallback(() => {
    const trimmedKeyword = localKeyword.trim();
    if (!trimmedKeyword) return;
    setActiveKeyword(trimmedKeyword);
    updateKeyword(trimmedKeyword);
    refresh();
  }, [localKeyword, refresh, updateKeyword]);

  const handleKeyDown = useCallback(
    (event: KeyboardEvent<HTMLInputElement>) => {
      if (event.key === "Enter") {
        event.preventDefault();
        handleSearch();
      }
    },
    [handleSearch]
  );

  const enrichedCompetitors = useMemo(() => {
    return competitors.map((competitor) =>
      buildCompetitorProfile(competitor, activeKeyword, keywordInsights, timeRange.value)
    );
  }, [competitors, activeKeyword, keywordInsights, timeRange.value]);

  const filteredCompetitors = useMemo(() => {
    const base = formatFilter === "all"
      ? enrichedCompetitors
      : enrichedCompetitors.filter((competitor) => competitor.format === formatFilter);

    const sorted = [...base].sort((a, b) => {
      switch (sortBy) {
        case "reviews":
          return b.reviewCount - a.reviewCount;
        case "sales":
          return b.monthlySales - a.monthlySales;
        case "royalties":
          return b.monthlyRoyalties - a.monthlyRoyalties;
        case "rank":
        default:
          return a.rank - b.rank;
      }
    });
    return sorted;
  }, [enrichedCompetitors, formatFilter, sortBy]);

  const totals = useMemo(() => {
    const totalSales = filteredCompetitors.reduce((sum, competitor) => sum + competitor.monthlySales, 0);
    const totalRoyalties = filteredCompetitors.reduce((sum, competitor) => sum + competitor.monthlyRoyalties, 0);
    const avgRating = filteredCompetitors.length
      ? filteredCompetitors.reduce((sum, competitor) => sum + competitor.rating, 0) / filteredCompetitors.length
      : 0;
    const currencySymbol = filteredCompetitors[0]?.currencySymbol ?? "$";
    return {
      totalSales,
      totalRoyalties,
      avgRating,
      currencySymbol
    };
  }, [filteredCompetitors]);

  const handleExport = useCallback(() => {
    if (filteredCompetitors.length === 0) return;
    const rows = [
      ["Rank", "Title", "ASIN", "Format", "Price", "Rating", "Reviews", "Monthly Sales", "Monthly Royalties", "Best Seller Rank"],
      ...filteredCompetitors.map((competitor) => [
        competitor.rank,
        competitor.title,
        competitor.asin,
        competitor.format,
        competitor.price,
        competitor.rating.toFixed(1),
        competitor.reviewCount,
        competitor.monthlySales,
        competitor.monthlyRoyalties,
        competitor.bestSellerRank
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
  }, [filteredCompetitors, activeKeyword]);

  const filteredReviews = useCallback(
    (reviews: ReviewEntry[]) => {
      const cutoff = new Date();
      cutoff.setDate(cutoff.getDate() - reviewWindow.value);
      return reviews.filter((review) => {
        if (ratingFilter > 0 && review.rating !== ratingFilter) return false;
        if (review.date < cutoff) return false;
        return true;
      });
    },
    [ratingFilter, reviewWindow]
  );

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

  const handleReverseSearch = useCallback(() => {
    const trimmed = reverseInput.trim();
    if (!trimmed) {
      setReverseError("Enter an ASIN or ISBN-10 to discover ranking keywords.");
      setReverseResults([]);
      return;
    }
    if (trimmed.length < 8) {
      setReverseError("Identifiers should be at least 8 characters long.");
      setReverseResults([]);
      return;
    }
    setReverseError(undefined);
    setReverseResults(generateReverseResults(trimmed.toUpperCase(), keywordInsights, activeKeyword));
  }, [reverseInput, keywordInsights, activeKeyword]);

  const reviewWindowLabel = reviewWindow.label.toLowerCase();

  return (
    <section id="competitors" className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-2xl space-y-4">
            <div>
              <h2 className="font-display text-3xl font-semibold text-white">Competitor analysis hub</h2>
              <p className="mt-2 text-sm text-white/70">
                Diagnose market leaders, benchmark your positioning, and surface keyword strategies that fuel their sales.
                RankBeam synthesises sales, rank, review and keyword telemetry for the {country.label} marketplace.
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
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wide text-white/60" htmlFor="reverse-asin">
                Discover ranking keywords (Reverse ASIN/ISBN)
              </label>
              <div className="mt-3 flex flex-col gap-3 sm:flex-row">
                <input
                  id="reverse-asin"
                  value={reverseInput}
                  onChange={(event) => setReverseInput(event.target.value)}
                  placeholder="Paste competitor ASIN or ISBN"
                  className="w-full rounded-full border border-white/10 bg-night/60 px-5 py-3 text-sm text-white placeholder:text-white/40 focus:border-aurora-400 focus:outline-none"
                />
                <button
                  type="button"
                  onClick={handleReverseSearch}
                  className="rounded-full border border-aurora-400 px-6 py-3 text-sm font-semibold text-aurora-200 transition hover:bg-aurora-500/10"
                >
                  Scan
                </button>
              </div>
              {reverseError && <p className="mt-2 text-xs text-rose-300">{reverseError}</p>}
            </div>
            <div className="flex items-center gap-2 text-xs text-white/60" aria-live="polite">
              <Sparkles className={`h-4 w-4 ${loading ? "animate-spin" : "text-aurora-400"}`} />
              {loading ? "Scanning competitor landscape…" : "Updated competitor intelligence"}
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
                <StatPill icon={<TrendingUp className="h-4 w-4 text-aurora-300" />} label="est. monthly sales" value={totals.totalSales.toLocaleString()} />
                <StatPill icon={<BarChart3 className="h-4 w-4 text-aurora-300" />} label="est. monthly royalties" value={formatCurrency(totals.totalRoyalties, totals.currencySymbol)} />
                <StatPill icon={<Star className="h-4 w-4 text-aurora-300" />} label="avg. rating" value={totals.avgRating.toFixed(1)} />
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              {[
                {
                  title: "Sales, Rank & Royalties",
                  description:
                    "Track historical sales velocity, royalties and BSR trendlines to understand launch cadence and evergreen potential.",
                  icon: <TrendingUp className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Keyword ranking insights",
                  description:
                    "Spot high-performing keywords, evaluate competition and cross-reference with your own listing for coverage gaps.",
                  icon: <Search className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Review monitoring",
                  description:
                    "Monitor new reviews, sentiment shifts and reader objections. Filter the feed by time horizon and star rating.",
                  icon: <Star className="h-5 w-5 text-aurora-300" />
                },
                {
                  title: "Organised tracking",
                  description:
                    "Slice the landscape by format, ranking, sales or royalties. Export a full CSV for deeper modelling or outreach.",
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
            <Filter className="h-4 w-4 text-aurora-300" />
            <span className="uppercase tracking-wide">Format</span>
            <div className="flex overflow-hidden rounded-full border border-white/10 text-white/70">
              <button
                type="button"
                onClick={() => setFormatFilter("all")}
                className={`px-3 py-1.5 transition ${formatFilter === "all" ? "bg-aurora-500/10 text-aurora-100" : "hover:bg-white/5"}`}
              >
                All
              </button>
              {FORMATS.map((format) => (
                <button
                  key={format}
                  type="button"
                  onClick={() => setFormatFilter(format)}
                  className={`px-3 py-1.5 transition ${formatFilter === format ? "bg-aurora-500/10 text-aurora-100" : "hover:bg-white/5"}`}
                >
                  {format}
                </button>
              ))}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4 text-aurora-300" />
            <span className="uppercase tracking-wide">Time range</span>
            <div className="flex overflow-hidden rounded-full border border-white/10 text-white/70">
              {TIME_RANGES.map((range) => (
                <button
                  key={range.value}
                  type="button"
                  onClick={() => setTimeRange(range)}
                  className={`px-3 py-1.5 transition ${timeRange.value === range.value ? "bg-aurora-500/10 text-aurora-100" : "hover:bg-white/5"}`}
                >
                  {range.label}
                </button>
              ))}
            </div>
          </div>
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
          {filteredCompetitors.map((book) => {
            const filteredReviewFeed = filteredReviews(book.reviews);
            const latestRank = book.rankSeries[book.rankSeries.length - 1];
            return (
              <article key={book.asin} className="flex flex-col gap-5 rounded-3xl border border-white/10 bg-black/40 p-6">
                <div className="flex flex-col gap-4 sm:flex-row">
                  <div className="relative h-48 w-full overflow-hidden rounded-2xl sm:w-40">
                    <img src={book.cover} alt={`${book.title} cover art`} className="h-full w-full object-cover" />
                    {book.isIndie && (
                      <span className="absolute left-3 top-3 rounded-full bg-emerald-500/80 px-3 py-1 text-xs font-semibold text-emerald-50">
                        Indie spotlight
                      </span>
                    )}
                  </div>
                  <div className="flex flex-1 flex-col gap-3 text-sm text-white/80">
                    <div>
                      <h3 className="text-lg font-semibold text-white">{book.title}</h3>
                      <p className="text-xs text-white/50">ASIN {book.asin}</p>
                    </div>
                    <div className="flex flex-wrap gap-2 text-xs text-white/60">
                      <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">{book.format}</span>
                      <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">Price {book.price}</span>
                      <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">Rating {book.rating.toFixed(1)}★</span>
                      <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">Reviews {book.reviewCount.toLocaleString()}</span>
                      <span className="rounded-full bg-white/5 px-3 py-1 text-white/70">{book.bestSellerRank}</span>
                    </div>
                    <div className="grid gap-3 text-xs text-white/70 sm:grid-cols-3">
                      <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                        <div className="flex items-center justify-between text-white">
                          <span className="font-semibold">Sales</span>
                          <span className="text-aurora-200">{formatChange(book.salesMomentum)}</span>
                        </div>
                        <p className="mt-2 text-lg font-semibold text-white">{book.monthlySales.toLocaleString()}</p>
                        <TrendSparkline series={book.salesSeries} color="#67e8f9" ariaLabel="Estimated sales trend" />
                      </div>
                      <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                        <div className="flex items-center justify-between text-white">
                          <span className="font-semibold">Royalties</span>
                          <span className="text-aurora-200">{formatChange(book.salesMomentum)}</span>
                        </div>
                        <p className="mt-2 text-lg font-semibold text-white">{formatCurrency(book.monthlyRoyalties, book.currencySymbol)}</p>
                        <TrendSparkline series={book.royaltySeries} color="#c4b5fd" ariaLabel="Estimated royalty trend" />
                      </div>
                      <div className="rounded-2xl border border-white/10 bg-night/60 p-3">
                        <div className="flex items-center justify-between text-white">
                          <span className="font-semibold">BSR</span>
                          <span className="text-aurora-200">{formatChange(book.rankMomentum)}</span>
                        </div>
                        <p className="mt-2 text-lg font-semibold text-white">#{Math.round(latestRank)}</p>
                        <TrendSparkline series={book.rankSeries} color="#fbcfe8" ariaLabel="Best seller rank trend" />
                      </div>
                    </div>
                    <a
                      href={book.url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex w-max items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-200 transition hover:border-aurora-400"
                    >
                      View listing <ExternalLink className="h-3.5 w-3.5" />
                    </a>
                  </div>
                </div>

                <div className="grid gap-4 lg:grid-cols-2">
                  <section className="rounded-2xl border border-white/10 bg-night/60 p-4">
                    <header className="flex items-center justify-between text-xs uppercase tracking-wide text-white/60">
                      <span>Top ranking keywords</span>
                      <span>Search volume / Competition / Rank</span>
                    </header>
                    <div className="mt-3 space-y-3 text-sm text-white/80">
                      {book.keywords.map((row) => (
                        <div key={`${book.asin}-${row.keyword}`} className="rounded-xl border border-white/5 bg-black/40 p-3">
                          <div className="flex items-center justify-between text-white">
                            <span className="font-semibold">{row.keyword}</span>
                            <span className="text-xs text-aurora-200">#{row.rankPosition}</span>
                          </div>
                          <dl className="mt-2 grid grid-cols-2 gap-2 text-xs text-white/60">
                            <div>
                              <dt className="uppercase tracking-wide">Search volume</dt>
                              <dd className="text-white">{row.searchVolume.toLocaleString()}</dd>
                            </div>
                            <div>
                              <dt className="uppercase tracking-wide">Competition</dt>
                              <dd className="text-white">{row.competitionScore.toFixed(2)}</dd>
                            </div>
                            <div>
                              <dt className="uppercase tracking-wide">Opportunity</dt>
                              <dd className="text-white">{row.opportunityScore}</dd>
                            </div>
                            <div>
                              <dt className="uppercase tracking-wide">Traffic share</dt>
                              <dd className="text-white">{row.trafficShare}%</dd>
                            </div>
                          </dl>
                        </div>
                      ))}
                    </div>
                  </section>
                  <section className="rounded-2xl border border-white/10 bg-night/60 p-4">
                    <header className="flex flex-wrap items-center justify-between gap-2 text-xs uppercase tracking-wide text-white/60">
                      <span>Recent reviews</span>
                      <div className="flex items-center gap-2 text-[11px] normal-case text-white/70">
                        <span>Rating:</span>
                        {RATING_FILTERS.map((option) => (
                          <button
                            key={option.value}
                            type="button"
                            onClick={() => setRatingFilter(option.value)}
                            className={`rounded-full px-2.5 py-1 transition ${
                              ratingFilter === option.value ? "bg-aurora-500/20 text-aurora-100" : "hover:bg-white/10"
                            }`}
                          >
                            {option.label}
                          </button>
                        ))}
                      </div>
                    </header>
                    <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-white/60">
                      <span>Showing {filteredReviewFeed.length} reviews from the last {reviewWindowLabel}.</span>
                      <div className="ml-auto flex items-center gap-1">
                        {REVIEW_WINDOWS.map((option) => (
                          <button
                            key={option.value}
                            type="button"
                            onClick={() => setReviewWindow(option)}
                            className={`rounded-full px-2.5 py-1 transition ${
                              reviewWindow.value === option.value ? "bg-aurora-500/20 text-aurora-100" : "hover:bg-white/10"
                            }`}
                          >
                            {option.label}
                          </button>
                        ))}
                      </div>
                    </div>
                    <div className="mt-3 space-y-3 text-sm text-white/80">
                      {filteredReviewFeed.slice(0, 3).map((review) => (
                        <div key={review.id} className="rounded-xl border border-white/5 bg-black/40 p-3">
                          <div className="flex items-center justify-between text-xs text-white/60">
                            <span>{review.reviewer}</span>
                            <span>{review.date.toLocaleDateString()}</span>
                          </div>
                          <div className="mt-1 flex items-center gap-2 text-xs text-amber-200">
                            {"★".repeat(review.rating)}
                            {"☆".repeat(5 - review.rating)}
                          </div>
                          <h4 className="mt-2 text-sm font-semibold text-white">{review.headline}</h4>
                          <p className="mt-1 text-sm text-white/70">{review.body}</p>
                        </div>
                      ))}
                      {filteredReviewFeed.length === 0 && (
                        <p className="text-xs text-white/60">No reviews in the selected window and rating filter.</p>
                      )}
                    </div>
                    <div className="mt-4 rounded-xl border border-white/5 bg-black/40 p-3 text-xs text-white/60">
                      <span className="font-semibold text-white">Review momentum:</span> {formatChange(book.reviewMomentum)} change in review velocity.
                    </div>
                  </section>
                </div>
              </article>
            );
          })}
        </div>
        <section className="mt-12 space-y-6 rounded-3xl border border-white/10 bg-black/40 p-8">
          <header className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h3 className="text-2xl font-semibold text-white">Reverse keyword results</h3>
              <p className="mt-1 text-sm text-white/70">
                Analyse the keyword footprint for competitor ASINs/ISBNs. Use the relevance column to map terms to your own
                catalogue positioning.
              </p>
            </div>
            {reverseResults.length > 0 && (
              <span className="text-xs uppercase tracking-wide text-white/60">{reverseResults.length} matching keywords</span>
            )}
          </header>
          <div className="overflow-hidden rounded-2xl border border-white/10">
            {reverseResults.length > 0 ? (
              <table className="min-w-full text-left text-sm text-white/80">
                <thead className="bg-white/5 text-xs uppercase tracking-wide text-white/60">
                  <tr>
                    <th className="px-6 py-3">Keyword</th>
                    <th className="px-6 py-3">Search volume</th>
                    <th className="px-6 py-3">Opportunity</th>
                    <th className="px-6 py-3">Competition</th>
                    <th className="px-6 py-3">Avg reviews</th>
                    <th className="px-6 py-3">Relevance</th>
                  </tr>
                </thead>
                <tbody>
                  {reverseResults.map((row) => (
                    <tr key={row.keyword} className="border-t border-white/5">
                      <td className="px-6 py-3 font-semibold text-white">{row.keyword}</td>
                      <td className="px-6 py-3">{row.searchVolume.toLocaleString()}</td>
                      <td className="px-6 py-3">{row.opportunityScore}</td>
                      <td className="px-6 py-3">{row.competitionLevel.toFixed(2)}</td>
                      <td className="px-6 py-3">{row.avgReviews.toLocaleString()}</td>
                      <td className="px-6 py-3">
                        <span className="inline-flex items-center gap-2 rounded-full bg-aurora-500/10 px-3 py-1 text-xs font-semibold text-aurora-100">
                          {row.relevance}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="px-6 py-10 text-sm text-white/60">
                Enter an ASIN or ISBN-10 above to reveal ranking keywords for a competitor listing.
              </div>
            )}
          </div>
        </section>

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
              Publisher and Publisher Pro subscribers can run unlimited competitor and reverse keyword scans, export full data
              sets and access AI-powered listing blueprints.
            </p>
            <div className="grid gap-3 text-sm">
              <div className="flex items-start gap-2">
                <Sparkles className="mt-0.5 h-4 w-4 text-aurora-200" />
                <span>Unlimited seed keyword and ASIN lookups</span>
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
                  Live scrapes are triggered whenever you run a search. Background refreshes keep trending data points updated so you can monitor shifts in BSR and reviews.
                </dd>
              </div>
              <div>
                <dt className="font-semibold text-white">Can I export the charts?</dt>
                <dd className="mt-1 text-white/70">
                  Use the export controls above to download CSV files. Publisher Pro subscribers also unlock Excel and PDF chart exports directly from this dashboard.
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
