import { useRankBeamStore } from "@/lib/state";
import { Activity, Compass, Globe2, Target } from "lucide-react";
import { useMemo } from "react";

export function InsightSummary() {
  const { keywordInsights, competitors, internationalKeywords } = useRankBeamStore();

  const metrics = useMemo(() => {
    if (keywordInsights.length === 0) {
      return [];
    }

    const sortedByVolume = [...keywordInsights].sort((a, b) => b.searchVolume - a.searchVolume);
    const highestDemand = sortedByVolume[0];

    const lowCompetition = keywordInsights
      .filter((item) => item.competitionScore <= 4 && item.titleDensity <= 15 && item.relevancyScore >= 0.6)
      .sort((a, b) => b.opportunityScore - a.opportunityScore);

    const averageCompetition =
      keywordInsights.reduce((total, item) => total + item.competitionScore, 0) / keywordInsights.length;
    const averageDemand = Math.round(
      keywordInsights.reduce((total, item) => total + item.demandScore, 0) / keywordInsights.length
    );

    const topInternational =
      internationalKeywords.length > 0
        ? internationalKeywords.reduce((best, current) => {
            if (!best || current.searchVolume > best.searchVolume) {
              return current;
            }
            return best;
          })
        : undefined;

    const competitorAverages = competitors.reduce(
      (acc, item) => {
        acc.rating += item.rating;
        acc.reviewCount += item.reviewCount;
        const numericPrice = parseFloat(item.price.replace(/[^0-9.]/g, ""));
        if (!Number.isNaN(numericPrice)) {
          acc.priceSum += numericPrice;
          acc.priceCount += 1;
        }
        return acc;
      },
      { rating: 0, reviewCount: 0, priceSum: 0, priceCount: 0 }
    );

    const averageRating = competitors.length > 0 ? competitorAverages.rating / competitors.length : undefined;
    const averageReviews = competitors.length > 0 ? Math.round(competitorAverages.reviewCount / competitors.length) : undefined;
    const averagePrice =
      competitorAverages.priceCount > 0 ? competitorAverages.priceSum / competitorAverages.priceCount : undefined;

    return [
      {
        title: "Highest demand keyword",
        primary: highestDemand.keyword,
        stat: `${highestDemand.searchVolume.toLocaleString()} searches`,
        context: `Demand score ${highestDemand.demandScore} • Competition ${Math.round(highestDemand.competitionScore * 10)}`,
        icon: Target
      },
      {
        title: "Opportunity rich phrases",
        primary: `${lowCompetition.length} ready to launch`,
        stat: lowCompetition[0]
          ? `${lowCompetition[0].keyword} (Opportunity ${lowCompetition[0].opportunityScore})`
          : "Review filter criteria",
        context: `Avg competition ${Math.round(averageCompetition * 10)} • Avg demand ${averageDemand}`,
        icon: Compass
      },
      {
        title: "Competitor health",
        primary:
          averageRating && averageReviews
            ? `${averageRating.toFixed(1)}★ • ${averageReviews.toLocaleString()} reviews`
            : "Monitoring...",
        stat:
          averagePrice !== undefined
            ? `Median price ~$${averagePrice.toFixed(2)}`
            : "Price mix unavailable",
        context: "Balances indie launches with legacy brands",
        icon: Activity
      },
      {
        title: "International breakout",
        primary: topInternational ? topInternational.countryName : "Scanning markets",
        stat: topInternational ? `${topInternational.searchVolume.toLocaleString()} searches` : "Add more storefronts",
        context: topInternational ? `Localized query: ${topInternational.keyword}` : "Awaiting expansion data",
        icon: Globe2
      }
    ];
  }, [keywordInsights, competitors, internationalKeywords]);

  if (metrics.length === 0) {
    return null;
  }

  return (
    <section className="border-b border-white/5 bg-night" aria-label="Keyword opportunity highlights">
      <div className="mx-auto max-w-6xl px-6 py-14">
        <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="font-display text-3xl font-semibold text-white">Opportunity radar</h2>
            <p className="mt-2 max-w-2xl text-sm text-white/70">
              RankBeam synthesises keyword, competitor and international datasets into quick-glance insights so you know exactly
              where to deploy creative and advertising energy.
            </p>
          </div>
        </div>
        <div className="grid gap-4 lg:grid-cols-4">
          {metrics.map((metric) => {
            const Icon = metric.icon;
            return (
              <article
                key={metric.title}
                className="group relative overflow-hidden rounded-3xl border border-white/10 bg-black/40 p-6"
              >
                <div className="absolute inset-0 bg-gradient-to-br from-aurora-500/0 via-aurora-500/5 to-aurora-500/0 opacity-0 transition group-hover:opacity-100" />
                <div className="relative flex items-start gap-4">
                  <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-aurora-500/10 text-aurora-200">
                    <Icon className="h-5 w-5" />
                  </span>
                  <div className="space-y-2 text-white/70">
                    <p className="text-xs uppercase tracking-wide text-white/50">{metric.title}</p>
                    <p className="text-sm font-semibold text-white">{metric.primary}</p>
                    <p className="text-sm text-aurora-100">{metric.stat}</p>
                    <p className="text-xs text-white/50">{metric.context}</p>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      </div>
    </section>
  );
}
