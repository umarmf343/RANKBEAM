import { create } from "zustand";
import type {
  CategoryTrend,
  CompetitorResult,
  GrowthSignal,
  InternationalKeyword,
  KeywordInsight
} from "@/lib/keywordEngine";
import type { CountryConfig } from "@/data/countries";
import { resolveCountry } from "@/data/countries";
import { fetchKeywordIntelligence } from "@/lib/api";

export type RankBeamState = {
  keyword: string;
  country: CountryConfig;
  keywordInsights: KeywordInsight[];
  categoryTrends: CategoryTrend[];
  internationalKeywords: InternationalKeyword[];
  competitors: CompetitorResult[];
  growthSignals: GrowthSignal[];
  headlineIdeas: string[];
  loading: boolean;
  suggestedKeywords: string[];
  dataSource: "scraped" | "error";
  lastUpdated?: string;
  error?: string;
  updateKeyword: (keyword: string) => void;
  updateCountry: (countryCode: string) => void;
  refresh: () => void;
};

const DEFAULT_KEYWORD = "low content books";
const DEFAULT_COUNTRY = resolveCountry("US");
let refreshHandle: ReturnType<typeof setTimeout> | undefined;

function toTitleCase(value: string): string {
  return value
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function deriveCategoryTrends(_: string, rows: KeywordInsight[]): CategoryTrend[] {
  if (rows.length === 0) {
    return [];
  }

  const sorted = [...rows].sort((a, b) => b.demandScore - a.demandScore);
  return sorted.slice(0, 6).map((row, index) => {
    const momentum = row.opportunityScore >= 70 ? "Rising" : row.opportunityScore >= 55 ? "Steady" : "Watch";
    const notes = `Demand ${row.demandScore} • Competition ${Math.round(row.competitionScore * 10)} • Density ${Math.round(row.titleDensity)}`;
    return {
      category: row.keyword,
      rank: index + 1,
      momentum,
      notes
    } satisfies CategoryTrend;
  });
}

function deriveGrowthSignals(rows: KeywordInsight[]): GrowthSignal[] {
  if (rows.length === 0) {
    return [];
  }

  const highestDemand = [...rows].sort((a, b) => b.demandScore - a.demandScore)[0];
  const bestOpportunity = [...rows].sort((a, b) => b.opportunityScore - a.opportunityScore)[0];
  const lowestCompetition = [...rows].sort((a, b) => a.competitionScore - b.competitionScore)[0];
  const leanDensity = [...rows].sort((a, b) => a.titleDensity - b.titleDensity)[0];

  const signals: GrowthSignal[] = [];

  if (highestDemand) {
    signals.push({
      label: "Demand velocity",
      score: Math.min(100, highestDemand.demandScore),
      description: `${toTitleCase(highestDemand.keyword)} is pacing ${highestDemand.searchVolume.toLocaleString()} searches.`
    });
  }

  if (bestOpportunity) {
    signals.push({
      label: "Launch-ready cluster",
      score: bestOpportunity.opportunityScore,
      description: `${toTitleCase(bestOpportunity.keyword)} balances reach with low friction intent.`
    });
  }

  if (lowestCompetition) {
    const competitionScore = Math.round(lowestCompetition.competitionScore * 10);
    signals.push({
      label: "Competition gap",
      score: Math.max(10, 100 - competitionScore * 5),
      description: `${toTitleCase(lowestCompetition.keyword)} shows lighter listings (score ${competitionScore}).`
    });
  }

  if (leanDensity) {
    signals.push({
      label: "Title whitespace",
      score: Math.max(10, Math.min(100, Math.round((20 - Math.min(20, leanDensity.titleDensity)) * 5))),
      description: `${toTitleCase(leanDensity.keyword)} appears in only ${Math.round(leanDensity.titleDensity)} titles.`
    });
  }

  return signals.slice(0, 4);
}

function deriveHeadlineIdeas(seed: string, rows: KeywordInsight[]): string[] {
  const ideas = new Set<string>();
  const normalizedSeed = seed.trim();
  const anchors = rows.slice(0, 4);

  if (normalizedSeed) {
    ideas.add(`${toTitleCase(normalizedSeed)} Launch Blueprint`);
  }

  anchors.forEach((row, index) => {
    if (!row.keyword) return;
    const base = toTitleCase(row.keyword);
    if (index === 0) {
      ideas.add(`${base} Accelerator Playbook`);
    } else {
      ideas.add(`${base} Growth Kit`);
    }
  });

  if (anchors.length >= 2) {
    const combo = `${toTitleCase(anchors[0].keyword)} & ${toTitleCase(anchors[1].keyword)} Expansion Plan`;
    ideas.add(combo);
  }

  return Array.from(ideas).slice(0, 5);
}

export const useRankBeamStore = create<RankBeamState>((set, get) => ({
  keyword: DEFAULT_KEYWORD,
  country: DEFAULT_COUNTRY,
  keywordInsights: [],
  categoryTrends: [],
  internationalKeywords: [],
  competitors: [],
  growthSignals: [],
  headlineIdeas: [],
  suggestedKeywords: [],
  loading: false,
  dataSource: "scraped",
  error: undefined,
  updateKeyword: (keyword) => {
    set({ keyword });
  },
  updateCountry: (countryCode) => {
    const country = resolveCountry(countryCode);
    set({ country });
  },
  refresh: () => {
    if (refreshHandle) {
      clearTimeout(refreshHandle);
    }

    set({ loading: true });

    refreshHandle = setTimeout(() => {
      const { keyword: currentKeyword, country: currentCountry } = get();
      void (async () => {
        try {
          const payload = await fetchKeywordIntelligence(currentKeyword, currentCountry.code);
          const categoryTrends = deriveCategoryTrends(currentKeyword, payload.keywords);
          const growthSignals = deriveGrowthSignals(payload.keywords);
          const headlineIdeas = deriveHeadlineIdeas(currentKeyword, payload.keywords);
          const competitors = payload.competitors ?? [];

          set({
            keywordInsights: payload.keywords,
            categoryTrends,
            internationalKeywords: [],
            competitors,
            growthSignals,
            headlineIdeas,
            suggestedKeywords: payload.suggestedKeywords,
            dataSource: payload.source,
            lastUpdated: payload.scrapedAt,
            loading: false,
            error: undefined
          });
        } catch (error) {
          set({
            keywordInsights: [],
            categoryTrends: [],
            internationalKeywords: [],
            competitors: [],
            growthSignals: [],
            headlineIdeas: [],
            suggestedKeywords: [],
            dataSource: "error",
            lastUpdated: undefined,
            loading: false,
            error: error instanceof Error ? error.message : "Failed to retrieve keyword intelligence"
          });
        } finally {
          refreshHandle = undefined;
        }
      })();
    }, 200);
  }
}));

// Refresh is intentionally triggered by user interactions (e.g. clicking "Go") so that
// scraping does not begin until the user explicitly requests it.
