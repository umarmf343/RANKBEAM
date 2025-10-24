import { useRankBeamStore } from "@/lib/state";
import { Activity, Gauge, Radar, TrendingUp } from "lucide-react";

const icons = [Gauge, Activity, Radar, TrendingUp];

export function SignalsPanel() {
  const { categoryTrends, growthSignals } = useRankBeamStore();

  return (
    <section className="border-b border-white/5 bg-night">
      <div className="mx-auto grid max-w-6xl gap-10 px-6 py-16 lg:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded-3xl border border-white/10 bg-black/40 p-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="font-display text-2xl font-semibold text-white">Category Momentum</h2>
              <p className="mt-2 max-w-lg text-sm text-white/70">
                Blend demand signals across trending Amazon categories to prioritise expansion. Momentum scores combine
                organic rank velocity, bestseller badge frequency and review velocity.
              </p>
            </div>
          </div>
          {categoryTrends.length > 0 ? (
            <div className="mt-6 space-y-4">
              {categoryTrends.map((trend) => (
                <div
                  key={trend.category}
                  className="flex items-center justify-between rounded-2xl border border-white/5 bg-white/5 px-5 py-4"
                >
                  <div>
                    <p className="text-xs uppercase tracking-wide text-white/50">#{trend.rank}</p>
                    <p className="text-sm font-semibold text-white">{trend.category}</p>
                    <p className="text-xs text-white/60">{trend.notes}</p>
                  </div>
                  <span
                    className={`rounded-full px-4 py-2 text-xs font-semibold ${
                      trend.momentum === "Rising"
                        ? "bg-emerald-500/20 text-emerald-200"
                        : trend.momentum === "Steady"
                        ? "bg-amber-500/20 text-amber-100"
                        : "bg-white/10 text-white/70"
                    }`}
                  >
                    {trend.momentum}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <div className="mt-6 rounded-2xl border border-dashed border-white/10 bg-black/30 p-6 text-sm text-white/60">
              Run a keyword scan to populate live category signals from Amazon search results.
            </div>
          )}
        </div>
        <div className="space-y-6">
          <div className="rounded-3xl border border-white/10 bg-black/40 p-8">
            <h3 className="font-display text-xl font-semibold text-white">Growth Signals</h3>
            <p className="mt-2 text-sm text-white/60">
              AI-weighted indicators highlighting pockets of growth before they saturate.
            </p>
            {growthSignals.length > 0 ? (
              <div className="mt-6 space-y-4">
                {growthSignals.map((signal, index) => {
                  const Icon = icons[index % icons.length];
                  return (
                    <div
                      key={signal.label}
                      className="flex items-center gap-4 rounded-2xl border border-white/5 bg-white/5 p-4"
                    >
                      <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-aurora-500/10 text-aurora-200">
                        <Icon className="h-6 w-6" />
                      </span>
                      <div className="flex-1">
                        <p className="text-sm font-semibold text-white">{signal.label}</p>
                        <p className="text-xs text-white/60">{signal.description}</p>
                      </div>
                      <span className="text-xl font-semibold text-white">{signal.score}</span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="mt-6 rounded-2xl border border-dashed border-white/10 bg-black/30 p-6 text-sm text-white/60">
                Growth indicators will be generated automatically once Amazon keywords have been scraped.
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
