import { ArrowRight, BarChart3, Globe2 } from "lucide-react";

export function Hero() {
  return (
    <section
      id="top"
      className="relative overflow-hidden border-b border-white/5 bg-[radial-gradient(circle_at_top,_rgba(76,102,241,0.25),_transparent_60%)]"
    >
      <div
        className="absolute inset-0 opacity-40"
        style={{
          backgroundImage:
            "url('https://images.unsplash.com/photo-1454165205744-3b78555e5572?auto=format&fit=crop&w=1600&q=80')",
          backgroundSize: "cover",
          backgroundPosition: "center"
        }}
        aria-hidden="true"
      />
      <div className="relative mx-auto max-w-6xl px-6 py-24">
        <div className="max-w-3xl space-y-8">
          <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-black/40 px-4 py-1 text-xs font-semibold uppercase tracking-wide text-aurora-200">
            <Globe2 className="h-4 w-4" /> Global Amazon Intelligence
          </span>
          <h1 className="font-display text-4xl font-bold leading-tight text-white sm:text-5xl">
            Discover, validate and dominate profitable Amazon keywords in minutes.
          </h1>
          <p className="text-lg text-white/70">
            RankBeam combines marketplace scraping, historical demand modeling and creative ideation tools into a single
            workspace built for Amazon authors and private label sellers. Launch smarter listings with data-backed keywords
            and compelling positioning across every storefront.
          </p>
          <div className="flex flex-col gap-4 sm:flex-row">
            <a
              href="#keywords"
              className="inline-flex items-center justify-center gap-2 rounded-full bg-gradient-to-r from-aurora-400 to-aurora-600 px-6 py-3 font-semibold text-white shadow-glow transition hover:scale-[1.01]"
            >
              Run Keyword Scan <ArrowRight className="h-4 w-4" />
            </a>
            <a
              href="#platform"
              className="inline-flex items-center justify-center gap-2 rounded-full border border-white/10 px-6 py-3 font-semibold text-white/80 transition hover:text-white"
            >
              Explore capabilities
            </a>
          </div>
          <dl className="grid grid-cols-2 gap-6 pt-8 sm:grid-cols-4">
            {[
              { label: "Marketplaces", value: "15" },
              { label: "Tracked Niches", value: "27k+" },
              { label: "Growth Recipes", value: "120" },
              { label: "Keywords Generated", value: "2.8M" }
            ].map((item) => (
              <div key={item.label} className="rounded-2xl border border-white/10 bg-black/40 p-4 text-center">
                <dt className="text-xs uppercase tracking-wide text-white/60">{item.label}</dt>
                <dd className="mt-2 font-display text-2xl font-semibold text-white">{item.value}</dd>
              </div>
            ))}
          </dl>
        </div>
        <div className="mt-16 grid gap-6 sm:grid-cols-3">
          {["AI Keyword Clusters", "Title Density Scoring", "Competitor Swipe Files"].map((feature) => (
            <div key={feature} className="rounded-2xl border border-white/10 bg-black/40 p-4">
              <BarChart3 className="h-10 w-10 text-aurora-400" />
              <p className="mt-4 text-sm font-semibold text-white">{feature}</p>
              <p className="mt-2 text-sm text-white/60">
                Crafted by RankBeam's signal engine to expose hidden opportunities before your competitors notice them.
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
