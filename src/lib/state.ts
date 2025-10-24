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
    if (refreshHandle) {
      clearTimeout(refreshHandle);
    }

    set({ loading: true });

    refreshHandle = setTimeout(() => {
      const { keyword: currentKeyword, country: currentCountry } = get();

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
        loading: false
      });

      refreshHandle = undefined;
    }, 200);
  }
}));

useRankBeamStore.getState().refresh();
