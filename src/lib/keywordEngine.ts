import { COUNTRIES, resolveCountry } from "../data/countries";

export type KeywordInsight = {
  keyword: string;
  searchVolume: number;
  competitionScore: number;
  relevancyScore: number;
  titleDensity: number;
  competitors: number;
  avgReviews: number;
  avgPrice: number;
  avgAge: number;
  opportunityScore: number;
  demandScore: number;
};

export type CategoryTrend = {
  category: string;
  rank: number;
  momentum: "Rising" | "Steady" | "Watch";
  notes: string;
};

export type CompetitorResult = {
  rank: number;
  title: string;
  asin: string;
  price: string;
  rating: number;
  reviewCount: number;
  bestSellerRank: string;
  url: string;
  cover: string;
  isIndie: boolean;
};

export type InternationalKeyword = {
  countryCode: string;
  countryName: string;
  keyword: string;
  searchVolume: number;
};

export type GrowthSignal = {
  label: string;
  score: number;
  description: string;
};

const keywordTemplates = [
  "%s",
  "%s guide",
  "%s workbook",
  "%s planner",
  "%s template",
  "%s checklist",
  "%s ideas",
  "best %s",
  "%s for beginners",
  "%s for kids",
  "how to %s",
  "%s marketing",
  "%s strategy",
  "%s 2024",
  "%s masterclass",
  "simple %s",
  "%s success",
  "%s toolkit",
  "%s prompts",
  "%s workbook for adults"
];

const categoryFallbacks = [
  "Self-Help",
  "Business & Money",
  "Education & Teaching",
  "Computers & Technology",
  "Health, Fitness & Dieting",
  "Parenting & Relationships",
  "Reference",
  "Crafts, Hobbies & Home",
  "Teen & Young Adult",
  "Children's Books"
];

const headlineAdjectives = [
  "Ultimate",
  "Essential",
  "Complete",
  "Practical",
  "Comprehensive",
  "Hands-On",
  "Step-by-Step",
  "No-Fluff",
  "Insider",
  "Rapid"
];

const headlineNouns = [
  "Blueprint",
  "Playbook",
  "Roadmap",
  "Workbook",
  "Guide",
  "Toolkit",
  "Accelerator",
  "Framework",
  "Mastery",
  "Bootcamp"
];

export function stableHash(input: string): number {
  let hash = 2166136261;
  for (let i = 0; i < input.length; i += 1) {
    hash ^= input.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}

export function stableFloat(key: string): number {
  return (stableHash(key) % 10000) / 10000;
}

function titleize(value: string): string {
  return value
    .split(/\s+/)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function estimateTitleMatches(keyword: string): { contain: number; exact: number } {
  const words = keyword.split(/\s+/);
  const contain = Math.max(1, Math.round(words.length * 1.6 + stableFloat(`contain-${keyword}`) * 6));
  const exact = Math.max(0, Math.round(words.length * 0.8 + stableFloat(`exact-${keyword}`) * 3));
  return { contain, exact };
}

export function generateKeywordInsights(seed: string, limit = 25): KeywordInsight[] {
  const normalized = seed.trim().toLowerCase() || "amazon publishing";
  const seen = new Set<string>();
  const variants: string[] = [];

  const addVariant = (value: string) => {
    const cleaned = value.replace(/\s+/g, " ").trim();
    if (!cleaned) return;
    const key = cleaned.toLowerCase();
    if (seen.has(key)) return;
    seen.add(key);
    variants.push(cleaned);
  };

  keywordTemplates.forEach((template) => addVariant(template.replace("%s", normalized)));
  ["kdp", "kindle", "amazon"].forEach((prefix) => addVariant(`${prefix} ${normalized}`));

  if (variants.length > limit) {
    variants.length = limit;
  }

  return variants.map((phrase, index) => {
    const weight = stableFloat(`kw-${phrase}-${index}`);
    const baseSearch = 650 + Math.round(weight * 450) - index * 23;
    const searchVolume = Math.max(120 + index * 5, baseSearch);
    const { contain, exact } = estimateTitleMatches(phrase);
    const relevancy = Math.round(Math.max(0.5, 0.92 - index * 0.035 + (1 - weight) * 0.18) * 100) / 100;
    const competitors = Math.max(15, Math.round(32 + contain * 2 + weight * 60 - index * 2));
    const avgReviews = Math.round(40 + weight * 380 + contain * 5 - index * 6);
    const avgPrice = Number.parseFloat((11 + weight * 16 + index * 0.35).toFixed(2));
    const avgAge = Math.round(8 + weight * 30 + index * 0.8);
    const competitionScore = Number.parseFloat(Math.max(1.1, contain / 9 + weight * 2.1).toFixed(2));
    const demandScore = Math.min(100, Math.round(searchVolume / 35 + weight * 12));
    const opportunityScore = Math.round(Math.min(100, demandScore * ((11 - competitionScore) / 10)));

    return {
      keyword: phrase,
      searchVolume,
      competitionScore,
      relevancyScore: relevancy,
      titleDensity: Number.parseFloat(exact.toFixed(2)),
      competitors,
      avgReviews: Math.max(18, avgReviews),
      avgPrice,
      avgAge,
      opportunityScore,
      demandScore
    };
  });
}

export function generateCategoryTrends(seed: string): CategoryTrend[] {
  const normalized = seed.trim() || "amazon publishing";
  return categoryFallbacks.map((category, index) => {
    const weight = stableFloat(`cat-${normalized}-${index}`);
    let momentum: CategoryTrend["momentum"] = "Watch";
    if (weight > 0.66) momentum = "Rising";
    else if (weight > 0.33) momentum = "Steady";

    return {
      category,
      rank: index + 1,
      momentum,
      notes: `Estimated from ${titleize(normalized)} search demand patterns`
    };
  });
}

export function generateHeadlineIdeas(seed: string): string[] {
  const normalized = titleize(seed.trim() || "Amazon Publishing");
  const ideas = new Set<string>();
  for (let i = 0; i < headlineAdjectives.length; i += 1) {
    const adjective = headlineAdjectives[i];
    const noun = headlineNouns[(i + normalized.length) % headlineNouns.length];
    ideas.add(`${adjective} ${normalized} ${noun}`);
  }
  return Array.from(ideas).slice(0, 8);
}

const sampleCovers = [
  "https://images.unsplash.com/photo-1498050108023-c5249f4df085?auto=format&fit=crop&w=360&q=80"
];

export function generateCompetitors(seed: string, country: string): CompetitorResult[] {
  const normalized = seed.trim() || "Amazon Publishing";
  const cfg = resolveCountry(country);
  const currency = cfg.currency === "USD" ? "$" : cfg.currency === "GBP" ? "£" : cfg.currency === "EUR" ? "€" : "";

  return Array.from({ length: 8 }).map((_, index) => {
    const hash = stableFloat(`competitor-${normalized}-${index}-${country}`);
    const priceValue = Math.round((12 + hash * 18 + index * 0.4) * 100) / 100;
    const rating = Math.round((4 + hash * 1) * 10) / 10;
    const reviewCount = Math.round(120 + hash * 900 + index * 35);
    const asin = `RB${(stableHash(`${normalized}-${index}`) % 900000 + 100000).toString(16).toUpperCase()}`.slice(0, 10);
    const category = categoryFallbacks[index % categoryFallbacks.length];
    const indie = hash > 0.55;

    return {
      rank: index + 1,
      title: `${titleize(normalized)} ${category} Secrets`,
      asin,
      price: currency ? `${currency}${priceValue.toFixed(2)}` : `${priceValue.toFixed(2)} ${cfg.currency}`,
      rating: Math.min(5, rating),
      reviewCount,
      bestSellerRank: `#${Math.round(5 + hash * 40)} in ${category}`,
      url: `https://${cfg.host}/dp/${asin}`,
      cover: sampleCovers[index % sampleCovers.length],
      isIndie: indie
    };
  });
}

export function generateInternationalKeywords(seed: string): InternationalKeyword[] {
  const normalized = seed.trim().toLowerCase() || "amazon publishing";
  return COUNTRIES.slice(0, 8).map((country, index) => {
    const weight = stableFloat(`intl-${normalized}-${country.code}-${index}`);
    const keyword = `${normalized} ${country.label.split(" ")[0].toLowerCase()}`;
    return {
      countryCode: country.code,
      countryName: country.label,
      keyword,
      searchVolume: Math.round(300 + weight * 800)
    };
  });
}

export function generateGrowthSignals(seed: string): GrowthSignal[] {
  const normalized = seed.trim() || "Amazon Publishing";
  return [
    {
      label: "Title Opportunity",
      score: Math.round((1 - stableFloat(`density-${normalized}`)) * 100),
      description: "Lower title density compared with adjacent niches"
    },
    {
      label: "Advertising Demand",
      score: Math.round((0.4 + stableFloat(`ads-${normalized}`) * 0.6) * 100),
      description: "Estimated demand based on sponsored placement counts"
    },
    {
      label: "Conversion Momentum",
      score: Math.round((0.55 + stableFloat(`momentum-${normalized}`) * 0.4) * 100),
      description: "Blended indicator using ratings, reviews and pricing"
    }
  ];
}
