import { useRankBeamStore } from "@/lib/state";
import { ExternalLink, Search, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState, type KeyboardEvent } from "react";

const KEY_FEATURES = [
  {
    title: "Sales, Rank, and Royalties Tracking",
    description:
      "Monitor the sales trends, Amazon Best Seller Rank (BSR), and estimated royalties of competitor books. This data helps you understand their market position and performance over time."
  },
  {
    title: "Keyword Ranking Insights",
    description:
      "Track the keyword rankings of competitor books to assess their visibility and discover which keywords are driving their traffic. This information can inform your own keyword strategy."
  },
  {
    title: "Review Monitoring",
    description:
      "Stay updated on new reviews and ratings for competitor books. RANKBEAM provides historical review data, allowing you to analyze feedback trends and identify areas where competitors excel or face challenges."
  },
  {
    title: "Organized Tracking",
    description:
      "Easily organize and filter tracked books by format (Kindle, paperback, audiobook) and other criteria. This feature helps you maintain a structured overview of your competitors."
  },
  {
    title: "Discover Ranking Keywords",
    description:
      "Input an ASIN or ISBN-10 to identify the keywords a book ranks for, is attempting to rank for, or is related to. This insight helps you understand which keywords are driving traffic to competitor books."
  },
  {
    title: "Comprehensive Keyword Data",
    description:
      "The tool provides up to thousands of related keywords, along with essential metrics such as search volume, opportunity score, competition level, and average reviews. This data enables you to assess the effectiveness of each keyword."
  },
  {
    title: "Advanced Sorting Capabilities",
    description:
      "Sort keywords by various metrics like search volume, opportunity score, and number of competitors. Additionally, filter keywords by specific words (e.g., \"journal\" or \"book\") using the word cloud sorting feature."
  },
  {
    title: "Unlimited Searches with Subscription Plans",
    description:
      "Subscribing to the Publisher or Publisher Pro plans grants you unlimited access to the Reverse Keyword Search Tool, enabling you to perform extensive keyword research without limitations."
  }
];

export function CompetitorShowcase() {
  const { competitors, country, keyword, updateKeyword, refresh, loading } = useRankBeamStore();
  const [localKeyword, setLocalKeyword] = useState(keyword);
  const [activeKeyword, setActiveKeyword] = useState(keyword);

  useEffect(() => {
    setLocalKeyword(keyword);
    setActiveKeyword(keyword);
  }, [keyword]);

  const handleSearch = useCallback(() => {
    const trimmedKeyword = localKeyword.trim();

    if (!trimmedKeyword) {
      return;
    }

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

  return (
    <section id="competitors" className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-8 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-2xl space-y-4">
            <div>
              <h2 className="font-display text-3xl font-semibold text-white">Competitor Swipe Gallery</h2>
              <p className="mt-2 text-sm text-white/70">
                Analyse cover patterns, pricing bands and positioning angles from the highest performing titles in the {country.label}
                marketplace. RankBeam flags indie publishers so you can spot agile competitors.
              </p>
            </div>
            <span className="inline-flex items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs uppercase tracking-wide text-white/60">
              Marketplace: {country.label}
            </span>
          </div>
          <div className="w-full max-w-md rounded-3xl border border-white/10 bg-black/40 p-6 shadow-[0_40px_120px_-60px_rgba(76,102,241,0.4)]">
            <label className="block text-xs font-semibold uppercase tracking-wide text-white/60" htmlFor="competitor-keyword">
              Keyword or ASIN search
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
            <div className="mt-4 flex items-center gap-2 text-xs text-white/60" aria-live="polite">
              <Sparkles className={`h-4 w-4 ${loading ? "animate-spin" : "text-aurora-400"}`} />
              {loading ? "Scanning competitor landscape…" : "Updated competitor intelligence"}
            </div>
          </div>
        </div>
        {activeKeyword && (
          <div className="mt-12 space-y-6 rounded-3xl border border-white/5 bg-black/40 p-8">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-aurora-200">RANKBEAM - Key Features</p>
                <h3 className="mt-2 text-2xl font-semibold text-white">
                  Competitive intelligence for "{activeKeyword}"
                </h3>
              </div>
              <p className="max-w-sm text-sm text-white/70">
                Use these capabilities to understand how rivals capture attention and where new opportunities exist in the {country.label} marketplace.
              </p>
            </div>
            <div className="grid gap-5 md:grid-cols-2">
              {KEY_FEATURES.map((feature) => (
                <article key={feature.title} className="rounded-2xl border border-white/10 bg-night/60 p-5">
                  <h4 className="text-sm font-semibold text-white">{feature.title}</h4>
                  <p className="mt-2 text-sm text-white/70">{feature.description}</p>
                </article>
              ))}
            </div>
          </div>
        )}
        <div className="mt-12 grid gap-6 md:grid-cols-2 xl:grid-cols-4">
          {competitors.map((book) => (
            <article key={book.asin} className="flex flex-col overflow-hidden rounded-3xl border border-white/10 bg-black/40">
              <div className="relative h-48 w-full overflow-hidden">
                <img
                  src={book.cover}
                  alt={`${book.title} cover art`}
                  className="h-full w-full object-cover transition duration-500 hover:scale-105"
                />
                {book.isIndie && (
                  <span className="absolute left-3 top-3 rounded-full bg-emerald-500/80 px-3 py-1 text-xs font-semibold text-emerald-50">
                    Indie spotlight
                  </span>
                )}
              </div>
              <div className="flex flex-1 flex-col gap-3 p-5 text-sm text-white/70">
                <div>
                  <h3 className="text-base font-semibold text-white">{book.title}</h3>
                  <p className="text-xs text-white/50">ASIN {book.asin}</p>
                </div>
                <dl className="space-y-1 text-xs">
                  <div className="flex justify-between">
                    <dt className="text-white/50">Best Seller Rank</dt>
                    <dd className="text-white/80">{book.bestSellerRank}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-white/50">Price</dt>
                    <dd className="text-white/80">{book.price}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-white/50">Rating</dt>
                    <dd className="text-white/80">{book.rating.toFixed(1)} ⭐</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-white/50">Reviews</dt>
                    <dd className="text-white/80">{book.reviewCount.toLocaleString()}</dd>
                  </div>
                </dl>
                <a
                  href={book.url}
                  target="_blank"
                  rel="noreferrer"
                  className="mt-auto inline-flex items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-aurora-200 transition hover:border-aurora-400"
                >
                  View listing <ExternalLink className="h-3.5 w-3.5" />
                </a>
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
