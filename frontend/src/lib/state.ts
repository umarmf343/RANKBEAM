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
  updateKeyword: (keyword: string) => void;
  updateCountry: (countryCode: string) => void;
  refresh: () => void;
};

const DEFAULT_KEYWORD = "low content books";
const DEFAULT_COUNTRY = resolveCountry("US");

export const useRankBeamStore = create<RankBeamState>((set, get) => ({
  keyword: DEFAULT_KEYWORD,
  country: DEFAULT_COUNTRY,
  keywordInsights: [],
  categoryTrends: [],
  internationalKeywords: [],
  competitors: [],
  growthSignals: [],
  headlineIdeas: [],
  loading: false,
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
    const { keyword, country } = get();
    set({ loading: true });

    const keywordInsights = generateKeywordInsights(keyword);
    const categoryTrends = generateCategoryTrends(keyword);
    const internationalKeywords = generateInternationalKeywords(keyword);
    const competitors = generateCompetitors(keyword, country.code);
    const growthSignals = generateGrowthSignals(keyword);
    const headlineIdeas = generateHeadlineIdeas(keyword);

    set({
      keywordInsights,
      categoryTrends,
      internationalKeywords,
      competitors,
      growthSignals,
      headlineIdeas,
      loading: false
    });
  }
}));

useRankBeamStore.getState().refresh();
