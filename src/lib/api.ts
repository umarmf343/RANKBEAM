import type { CompetitorResult, KeywordInsight } from "@/lib/keywordEngine";

export type KeywordIntelligencePayload = {
  keyword: string;
  country: string;
  keywords: KeywordInsight[];
  competitors: CompetitorResult[];
  suggestedKeywords: string[];
  source: "scraped" | "fallback";
  scrapedAt: string;
};

function buildQuery(params: Record<string, string | undefined>): string {
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      searchParams.set(key, value);
    }
  });
  return searchParams.toString();
}

export async function fetchKeywordIntelligence(keyword: string, countryCode: string): Promise<KeywordIntelligencePayload> {
  const query = buildQuery({ keyword, country: countryCode });
  const response = await fetch(`/api/keywords?${query}`);
  if (!response.ok) {
    throw new Error(`Unable to refresh keyword intelligence (${response.status})`);
  }
  const payload = (await response.json()) as KeywordIntelligencePayload;
  return payload;
}
