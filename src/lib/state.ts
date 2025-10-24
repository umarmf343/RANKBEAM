import { create } from "zustand";
import type {
  CategoryTrend,
  CompetitorResult,
  GrowthSignal,
  InternationalKeyword,
  KeywordInsight
} from "@/lib/keywordEngine";
import {
  generateCategoryTrends,
  generateCompetitors,
  generateGrowthSignals,
  generateHeadlineIdeas,
  generateInternationalKeywords,
  generateKeywordInsights
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
  dataSource: "scraped" | "fallback";
  lastUpdated?: string;
  updateKeyword: (keyword: string) => void;
  updateCountry: (countryCode: string) => void;
  refresh: () => void;
};

const DEFAULT_KEYWORD = "low content books";
const DEFAULT_COUNTRY = resolveCountry("US");
let refreshHandle: ReturnType<typeof setTimeout> | undefined;

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
  dataSource: "fallback",
  updateKeyword: (keyword) => {
    set({ keyword });
    get().refresh();
  },
  updateCountry: (countryCode) => {
    const country = resolveCountry(countryCode);
    set({ country });
    get().refresh();
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
          const categoryTrends = generateCategoryTrends(currentKeyword);
          const internationalKeywords = generateInternationalKeywords(currentKeyword);
          const growthSignals = generateGrowthSignals(currentKeyword);
          const headlineIdeas = generateHeadlineIdeas(currentKeyword);
          const competitors =
            payload.competitors && payload.competitors.length > 0
              ? payload.competitors
              : generateCompetitors(currentKeyword, currentCountry.code);

          set({
            keywordInsights: payload.keywords,
            categoryTrends,
            internationalKeywords,
            competitors,
            growthSignals,
            headlineIdeas,
            suggestedKeywords: payload.suggestedKeywords,
            dataSource: payload.source,
            lastUpdated: payload.scrapedAt,
            loading: false
          });
        } catch (error) {
          const keywordInsights = generateKeywordInsights(currentKeyword);
          const categoryTrends = generateCategoryTrends(currentKeyword);
          const internationalKeywords = generateInternationalKeywords(currentKeyword);
          const competitors = generateCompetitors(currentKeyword, currentCountry.code);
          const growthSignals = generateGrowthSignals(currentKeyword);
          const headlineIdeas = generateHeadlineIdeas(currentKeyword);
          set({
            keywordInsights,
            categoryTrends,
            internationalKeywords,
            competitors,
            growthSignals,
            headlineIdeas,
            suggestedKeywords: keywordInsights.slice(1, 12).map((row) => row.keyword),
            dataSource: "fallback",
            lastUpdated: new Date().toISOString(),
            loading: false
          });
        } finally {
          refreshHandle = undefined;
        }
      })();
    }, 200);
  }
}));

useRankBeamStore.getState().refresh();
