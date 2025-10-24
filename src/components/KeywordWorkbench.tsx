import { COUNTRIES } from "@/data/countries";
import { useRankBeamStore } from "@/lib/state";
import { Search, Sparkles } from "lucide-react";
import { useCallback, useEffect, useState, type KeyboardEvent } from "react";

export function KeywordWorkbench() {
  const { keyword, country, updateKeyword, updateCountry, refresh, loading, dataSource, error } = useRankBeamStore();
  const [localKeyword, setLocalKeyword] = useState(keyword);

  const handleSearch = useCallback(() => {
    const trimmedKeyword = localKeyword.trim();

    if (!trimmedKeyword) {
      return;
    }

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

  useEffect(() => {
    setLocalKeyword(keyword);
  }, [keyword]);

  return (
    <section id="keywords" className="border-b border-white/5 bg-night">
      <div className="mx-auto flex max-w-6xl flex-col gap-10 px-6 py-16 lg:flex-row">
        <div className="flex-1 space-y-6">
          <h2 className="font-display text-3xl font-semibold text-white">Keyword Intelligence Workbench</h2>
          <p className="max-w-xl text-sm text-white/70">
            Combine RankBeam's keyword engine with live marketplace signals. Start with a seed idea, choose your
            storefront and instantly uncover suggested phrases, search volume, competition and listing density metrics.
          </p>
          <div className="rounded-3xl border border-white/10 bg-black/40 p-6 shadow-[0_40px_120px_-60px_rgba(76,102,241,0.4)]">
            <label
              className="block text-xs font-semibold uppercase tracking-wide text-white/60"
              htmlFor="rankbeam-seed-keyword"
            >
              Seed keyword
            </label>
            <div className="mt-3 flex flex-col gap-3 sm:flex-row">
              <div className="relative flex-1">
                <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-white/40" />
                <input
                  id="rankbeam-seed-keyword"
                  value={localKeyword}
                  onChange={(event) => setLocalKeyword(event.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="e.g. mindfulness journal"
                  className="w-full rounded-full border border-white/10 bg-night/60 px-10 py-3 text-sm text-white placeholder:text-white/40 focus:border-aurora-400 focus:outline-none"
                />
              </div>
              <div className="flex flex-col gap-1">
                <label className="sr-only" htmlFor="rankbeam-marketplace">
                  Marketplace
                </label>
                <select
                  id="rankbeam-marketplace"
                  value={country.code}
                  onChange={(event) => updateCountry(event.target.value)}
                  className="rounded-full border border-white/10 bg-night/80 px-4 py-3 text-sm text-white focus:border-aurora-400 focus:outline-none"
                >
                  {COUNTRIES.map((item) => (
                    <option key={item.code} value={item.code}>
                      {item.label}
                    </option>
                  ))}
                </select>
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
              <Sparkles
                className={`h-4 w-4 ${
                  loading ? "animate-spin" : dataSource === "error" ? "text-rose-400" : "text-aurora-400"
                }`}
              />
              {loading
                ? "Refreshing keyword intelligence…"
                : dataSource === "error"
                ? error ?? "Unable to load live keyword intelligence"
                : "Insights updated in real-time"}
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            {[
              {
                title: "Audience Intent",
                description: "Understand how readers search and which angles drive discovery."
              },
              {
                title: "Listing Blueprint",
                description: "Optimize titles, subtitles and bullets for algorithm relevance."
              },
              {
                title: "Competitive Radar",
                description: "Track independent publishers and emerging best-sellers."
              },
              {
                title: "Global Expansion",
                description: "Localize keyword clusters across 15 marketplaces in one click."
              }
            ].map((card) => (
              <div key={card.title} className="rounded-2xl border border-white/10 bg-black/30 p-5">
                <h3 className="text-sm font-semibold text-white">{card.title}</h3>
                <p className="mt-2 text-sm text-white/60">{card.description}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
