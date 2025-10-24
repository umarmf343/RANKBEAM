import { useRankBeamStore } from "@/lib/state";
import { ExternalLink, LineChart } from "lucide-react";
import { useMemo } from "react";

export function CompetitorShowcase() {
  const { competitors, country } = useRankBeamStore();
  const summary = useMemo(() => {
    if (!competitors.length) {
      return {
        averagePrice: 0,
        averageRating: 0,
        medianReviews: 0,
        indieShare: 0
      };
    }

    const priceValues = competitors
      .map((book) => parseFloat(book.price.replace(/[^[0-9.,]/g, "").replace(/,/g, "")))
      .filter((value) => Number.isFinite(value));
    const sortedReviews = [...competitors.map((book) => book.reviewCount)].sort((a, b) => a - b);
    const mid = Math.floor(sortedReviews.length / 2);
    const medianReviews =
      sortedReviews.length % 2 === 0
        ? Math.round((sortedReviews[mid - 1] + sortedReviews[mid]) / 2)
        : sortedReviews[mid];

    return {
      averagePrice:
        priceValues.length > 0
          ? priceValues.reduce((total, value) => total + value, 0) / priceValues.length
          : 0,
      averageRating:
        competitors.reduce((total, book) => total + book.rating, 0) / Math.max(competitors.length, 1),
      medianReviews,
      indieShare: Math.round((competitors.filter((book) => book.isIndie).length / competitors.length) * 100)
    };
  }, [competitors]);

  return (
    <section id="competitors" className="border-b border-white/5 bg-night">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">Competitor Swipe Gallery</h2>
            <p className="mt-2 max-w-2xl text-sm text-white/70">
              Analyse cover patterns, pricing bands and positioning angles from the highest performing titles in the {country.label}
              marketplace. RankBeam flags indie publishers so you can spot agile competitors.
            </p>
          </div>
          <span className="inline-flex items-center gap-2 rounded-full border border-white/10 px-4 py-2 text-xs uppercase tracking-wide text-white/60">
            Marketplace: {country.label}
          </span>
        </div>
        <div className="mt-8 grid gap-4 rounded-3xl border border-white/10 bg-black/40 p-6 text-xs text-white/60 sm:grid-cols-2 xl:grid-cols-4">
          {[{
            label: "Avg. price",
            value:
              summary.averagePrice
                ? new Intl.NumberFormat("en-US", {
                    style: "currency",
                    currency: country.currency || "USD"
                  }).format(summary.averagePrice)
                : "--"
          },
          { label: "Avg. rating", value: summary.averageRating.toFixed(2) },
          { label: "Median reviews", value: summary.medianReviews.toLocaleString() },
          { label: "Indie presence", value: `${summary.indieShare}% of top titles` }].map((stat) => (
            <div key={stat.label} className="flex items-center justify-between rounded-2xl border border-white/5 bg-white/5 px-4 py-3 text-sm">
              <div>
                <p className="text-xs uppercase tracking-wide text-white/50">{stat.label}</p>
                <p className="mt-1 font-semibold text-white">{stat.value}</p>
              </div>
              <span className="flex h-9 w-9 items-center justify-center rounded-full bg-aurora-500/10 text-aurora-200">
                <LineChart className="h-4 w-4" />
              </span>
            </div>
          ))}
        </div>
        <div className="mt-8 grid gap-6 md:grid-cols-2 xl:grid-cols-4">
          {competitors.map((book) => (
            <article key={book.asin} className="flex flex-col overflow-hidden rounded-3xl border border-white/10 bg-black/40">
              <div className="relative h-48 w-full overflow-hidden">
                <img src={book.cover} alt="" className="h-full w-full object-cover transition duration-500 hover:scale-105" />
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
