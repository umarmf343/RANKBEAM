package scraper

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"
)

var (
	keywordTemplates = []string{
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
	}

	categoryFallbacks = []string{
		"Self-Help",
		"Business & Money",
		"Education & Teaching",
		"Computers & Technology",
		"Health, Fitness & Dieting",
		"Parenting & Relationships",
		"Reference",
		"Crafts, Hobbies & Home",
		"Teen & Young Adult",
		"Children's Books",
	}

	headlineAdjectives = []string{
		"Ultimate",
		"Essential",
		"Complete",
		"Practical",
		"Comprehensive",
		"Hands-On",
		"Step-by-Step",
		"No-Fluff",
		"Insider",
		"Rapid",
	}

	headlineNouns = []string{
		"Blueprint",
		"Playbook",
		"Roadmap",
		"Workbook",
		"Guide",
		"Toolkit",
		"Accelerator",
		"Framework",
		"Mastery",
		"Bootcamp",
	}
)

func synthesizeKeywordInsights(seed string, limit int) []KeywordInsight {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = "amazon publishing"
	}

	lower := strings.ToLower(seed)
	seen := map[string]struct{}{}
	variants := make([]string, 0, len(keywordTemplates))

	add := func(value string) {
		cleaned := strings.Join(strings.Fields(value), " ")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
		if cleaned == "" {
			return
		}
		key := strings.ToLower(cleaned)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		variants = append(variants, cleaned)
	}

	for _, tpl := range keywordTemplates {
		formatted := fmt.Sprintf(tpl, lower)
		add(formatted)
	}

	prefixes := []string{"kdp", "kindle", "amazon"}
	for _, prefix := range prefixes {
		add(fmt.Sprintf("%s %s", prefix, lower))
	}

	if len(variants) > limit && limit > 0 {
		variants = variants[:limit]
	}

	insights := make([]KeywordInsight, 0, len(variants))
	for idx, phrase := range variants {
		weight := stableFloat(fmt.Sprintf("kw-%s-%d", phrase, idx))
		search := 650 + int(weight*450) - idx*23
		if search < 120 {
			search = 120 + idx*5
		}
		competition := math.Round((0.18+weight*0.55)*100) / 100
		relevancy := math.Round(math.Max(0.5, 0.92-float64(idx)*0.035+(1-weight)*0.18)*100) / 100
		density := math.Round((0.22+(1-weight)*0.4)*100) / 100
		insights = append(insights, KeywordInsight{
			Keyword:          phrase,
			SearchVolume:     search,
			CompetitionScore: competition,
			RelevancyScore:   relevancy,
			TitleDensity:     density,
		})
	}

	return insights
}

func synthesizeCategoryTrends(keyword string) []CategoryTrend {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		keyword = "amazon publishing"
	}

	trends := make([]CategoryTrend, 0, len(categoryFallbacks))
	for idx, name := range categoryFallbacks {
		weight := stableFloat(fmt.Sprintf("cat-%s-%d", keyword, idx))
		momentum := "Watch"
		switch {
		case weight > 0.66:
			momentum = "Rising"
		case weight > 0.33:
			momentum = "Steady"
		}
		trends = append(trends, CategoryTrend{
			Category: name,
			Rank:     idx + 1,
			Momentum: momentum,
			Notes:    fmt.Sprintf("Estimated from %s search demand patterns", titleize(keyword)),
		})
	}

	return trends
}

func synthesizeBestsellers(keyword, country string) []BestsellerProduct {
	cfg := ConfigFor(country)
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		keyword = "publishing"
	}

	products := make([]BestsellerProduct, 0, 10)
	for i := 0; i < 10; i++ {
		adjective := headlineAdjectives[i%len(headlineAdjectives)]
		noun := headlineNouns[i%len(headlineNouns)]
		title := fmt.Sprintf("%s %s: %s", adjective, noun, titleize(keyword))
		price := formatPrice(cfg.Currency, 8.99+float64(i)*0.75+stableFloat(fmt.Sprintf("price-%d-%s", i, keyword))*6.5)
		reviews := int(1200 + stableFloat(fmt.Sprintf("rev-%d-%s", i, keyword))*3200)
		rating := 4.2 + stableFloat(fmt.Sprintf("rating-%d-%s", i, keyword))*0.7
		asin := fmt.Sprintf("OFFLINE%05d", i+1)
		bestsellerRank := 350 + i*27 + int(stableFloat(fmt.Sprintf("bsr-%d-%s", i, keyword))*120)
		indie := i%3 != 0
		density := math.Round((0.24+stableFloat(fmt.Sprintf("density-%d-%s", i, keyword))*0.5)*100) / 100
		url := fmt.Sprintf("https://%s/dp/%s", cfg.Host, asin)

		products = append(products, BestsellerProduct{
			Rank:         i + 1,
			Title:        title,
			ASIN:         asin,
			Price:        price,
			Rating:       fmt.Sprintf("%.1f out of 5 stars", rating),
			ReviewCount:  fmt.Sprintf("%s ratings", humanizeNumber(reviews)),
			Publisher:    map[bool]string{true: "Independently published", false: "Curated Press"}[indie],
			BestSeller:   bestsellerRank,
			Category:     categoryFallbacks[i%len(categoryFallbacks)],
			IsIndie:      indie,
			TitleDensity: density,
			URL:          url,
		})
	}

	return products
}

func synthesizeInternationalKeywords(keyword string, countries []string) []InternationalKeyword {
	if len(countries) == 0 {
		countries = Countries()
	}
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		keyword = "amazon publishing"
	}

	results := make([]InternationalKeyword, 0, len(countries)*3)
	for _, country := range countries {
		cfg := ConfigFor(country)
		base := synthesizeKeywordInsights(fmt.Sprintf("%s %s", keyword, strings.ToLower(cfg.Country)), 3)
		sort.Slice(base, func(i, j int) bool { return base[i].SearchVolume > base[j].SearchVolume })
		for i := 0; i < len(base) && i < 3; i++ {
			results = append(results, InternationalKeyword{
				CountryCode:  country,
				CountryName:  cfg.Country,
				Keyword:      base[i].Keyword,
				SearchVolume: base[i].SearchVolume,
			})
		}
	}

	return results
}

func synthesizeProductDetails(asin, country string) *ProductDetails {
	cfg := ConfigFor(country)
	title := fmt.Sprintf("%s Research Preview", strings.ToUpper(asin))
	price := formatPrice(cfg.Currency, 12.49+stableFloat(asin)*10)
	rating := 4.1 + stableFloat("rating-"+asin)*0.8
	reviews := 240 + int(stableFloat("review-"+asin)*2600)
	density := math.Round((0.2+stableFloat("density-"+asin)*0.5)*100) / 100
	ranks := []BestSellerRank{{
		Category: categoryFallbacks[int(stableFloat("cat-"+asin)*float64(len(categoryFallbacks)))],
		Rank:     250 + int(stableFloat("rank-"+asin)*600),
	}}

	return &ProductDetails{
		Title:           title,
		ASIN:            strings.ToUpper(asin),
		Price:           price,
		Currency:        cfg.Currency,
		Rating:          fmt.Sprintf("%.1f out of 5 stars", rating),
		ReviewCount:     fmt.Sprintf("%s ratings", humanizeNumber(reviews)),
		Availability:    "Estimated stock available",
		ImageURL:        "",
		Brand:           "Offline Insight",
		DeliveryMessage: "Delivery estimate varies while offline",
		Publisher:       "Independently published",
		BestSellerRanks: ranks,
		IsIndependent:   true,
		TitleDensity:    density,
		FetchedAt:       now(),
		URL:             fmt.Sprintf("https://%s/dp/%s", cfg.Host, strings.ToUpper(asin)),
	}
}

func formatPrice(currency string, value float64) string {
	value = math.Round(value*100) / 100
	symbol := currencySymbol(currency)
	return fmt.Sprintf("%s%.2f", symbol, value)
}

func currencySymbol(currency string) string {
	switch strings.ToUpper(currency) {
	case "USD":
		return "$"
	case "CAD":
		return "C$"
	case "GBP":
		return "£"
	case "EUR":
		return "€"
	case "AUD":
		return "A$"
	case "INR":
		return "₹"
	case "JPY":
		return "¥"
	case "BRL":
		return "R$"
	case "MXN":
		return "MX$"
	case "AED":
		return "د.إ"
	case "SGD":
		return "S$"
	default:
		return ""
	}
}

func stableFloat(seed string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	return float64(h.Sum32()%1000) / 999
}

func humanizeNumber(value int) string {
	switch {
	case value >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(value)/1_000_000)
	case value >= 1_000:
		return fmt.Sprintf("%.1fK", float64(value)/1_000)
	default:
		return fmt.Sprintf("%d", value)
	}
}

func now() time.Time {
	return time.Now()
}

func titleize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
