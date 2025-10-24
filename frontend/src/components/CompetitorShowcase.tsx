import { useRankBeamStore } from "@/lib/state";
import { ExternalLink } from "lucide-react";

export function CompetitorShowcase() {
  const { competitors, country } = useRankBeamStore();

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
