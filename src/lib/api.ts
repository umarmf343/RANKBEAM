import type { CompetitorResult, KeywordInsight } from "@/lib/keywordEngine";

export type KeywordIntelligencePayload = {
  keyword: string;
  country: string;
  keywords: KeywordInsight[];
  competitors: CompetitorResult[];
  suggestedKeywords: string[];
  source: "scraped";
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
  const body = await response.text();
  if (!response.ok) {
    let parsed: unknown;
    try {
      parsed = body ? JSON.parse(body) : undefined;
    } catch {
      parsed = undefined;
    }
    if (
      parsed &&
      typeof parsed === "object" &&
      "error" in parsed &&
      typeof (parsed as { error?: unknown }).error === "string"
    ) {
      const message = (parsed as { error: string }).error.trim();
      if (message.length > 0) {
        throw new Error(message);
      }
    }
    throw new Error(`Unable to refresh keyword intelligence (${response.status})`);
  }
  const payload = JSON.parse(body) as KeywordIntelligencePayload;
  return payload;
}
