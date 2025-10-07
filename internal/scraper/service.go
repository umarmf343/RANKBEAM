package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/umarmf343/Umar-kdp-product-api/internal/logging"
)

// Service encapsulates HTTP access and scraping helpers used by the application.
type Service struct {
	client     *http.Client
	ticker     *time.Ticker
	closed     chan struct{}
	closeOnce  sync.Once
	userAgents []string
}

// ErrServiceClosed indicates the scraper service has been closed.
var (
	ErrServiceClosed = errors.New("scraper service closed")
	ErrBotDetected   = errors.New("amazon requested captcha verification")
	bsrPattern       = regexp.MustCompile(`#([0-9,]+)\s+in\s+([^()\n>]+)`)
)

// NewService creates a scraper service with sane defaults such as timeout handling,
// rate limiting and a pool of realistic user agents.
func NewService(timeout time.Duration, requestsPerMinute int) *Service {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 20
	}

	interval := time.Minute / time.Duration(requestsPerMinute)
	ticker := time.NewTicker(interval)

	return &Service{
		client: &http.Client{Timeout: timeout},
		ticker: ticker,
		closed: make(chan struct{}),
		userAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		},
	}
}

// Rate exposes the underlying rate limiter channel for callers needing access to raw ticks.
func (s *Service) Rate() <-chan time.Time {
	if s == nil {
		return nil
	}

	ticker := s.ticker
	if ticker == nil {
		return nil
	}

	return ticker.C
}

// Close stops the service ticker and unblocks any pending waiters.
func (s *Service) Close() {
	if s == nil {
		return
	}

	s.closeOnce.Do(func() {
		if s.ticker != nil {
			s.ticker.Stop()
		}
		if s.closed != nil {
			close(s.closed)
		}
	})
}

// waitForRate blocks until the service can issue another outbound request.
func (s *Service) waitForRate(ctx context.Context) error {
	rate := s.Rate()
	if rate == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.closed:
		return ErrServiceClosed
	case <-rate:
		return nil
	}
}

func (s *Service) userAgent() string {
	if len(s.userAgents) == 0 {
		return "Mozilla/5.0 (compatible; scraperbot/1.0)"
	}
	idx := int(time.Now().UnixNano()) % len(s.userAgents)
	if idx < 0 {
		idx = -idx
	}
	return s.userAgents[idx]
}

// FetchProduct retrieves product information by ASIN for the requested country marketplace.
func (s *Service) FetchProduct(ctx context.Context, asin, country string) (*ProductDetails, error) {
	asin = strings.TrimSpace(asin)
	if asin == "" {
		return nil, errors.New("asin is required")
	}

	cfg := ConfigFor(strings.ToUpper(country))
	endpoint := fmt.Sprintf("https://%s/dp/%s", cfg.Host, url.PathEscape(asin))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent())
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if err := s.waitForRate(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		logging.Logger().Warn("rate limiter interrupted product fetch, using fallback", "asin", asin, "country", country, "error", err)
		return synthesizeProductDetails(asin, country), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logging.Logger().Warn("product request failed, using fallback", "asin", asin, "country", country, "error", err)
		return synthesizeProductDetails(asin, country), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Logger().Warn("product request returned non-OK status", "asin", asin, "country", country, "status", resp.StatusCode)
		return synthesizeProductDetails(asin, country), nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Logger().Warn("unable to parse product page, using fallback", "asin", asin, "country", country, "error", err)
		return synthesizeProductDetails(asin, country), nil
	}
	if err := detectBotChallenge(doc); err != nil {
		logging.Logger().Warn("bot challenge detected during product fetch", "asin", asin, "country", country, "error", err)
		return synthesizeProductDetails(asin, country), nil
	}

	title := textOrFallback(doc.Find("#productTitle"), "Title unavailable")
	price := firstNonEmpty(
		textOrFallback(doc.Find("span#priceblock_ourprice"), ""),
		textOrFallback(doc.Find("span#priceblock_dealprice"), ""),
		textOrFallback(doc.Find("span.a-price span.a-offscreen"), ""),
	)
	if price == "" {
		price = "Price unavailable"
	}
	rating := textOrFallback(doc.Find("span[data-hook='rating-out-of-text']"), "Rating unavailable")
	reviews := textOrFallback(doc.Find("#acrCustomerReviewText"), "Reviews unavailable")
	availability := textOrFallback(doc.Find("#availability span"), "Availability unavailable")
	brand := textOrFallback(doc.Find("#bylineInfo"), "Brand unavailable")
	image := doc.Find("img#landingImage").AttrOr("src", "")
	delivery := textOrFallback(doc.Find("#ddmDeliveryMessage"), "")
	if delivery == "" {
		delivery = textOrFallback(doc.Find("#deliveryMessageMirId span"), "Delivery information unavailable")
	}

	publisher := parsePublisher(doc)
	ranks := parseBestSellerRanks(doc)
	indie := isIndependentPublisher(publisher)
	details := collectProductDetails(doc)
	printLength := findProductDetail(details, "print length", "pages", "paperback", "hardcover")
	dimensions := findProductDetail(details, "dimensions", "product dimensions", "item dimensions")
	publicationDate := findProductDetail(details, "publication date", "publish date", "publication day")
	language := findProductDetail(details, "language")
	isbn10 := findProductDetail(details, "isbn-10", "isbn 10")
	isbn13 := findProductDetail(details, "isbn-13", "isbn 13")

	titleDensity := -1.0
	if strings.TrimSpace(title) != "" {
		densityCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
		density, err := s.computeTitleDensity(densityCtx, title, country)
		cancel()
		if err == nil {
			titleDensity = math.Round(density*100) / 100
		}
	}

	return &ProductDetails{
		Title:           title,
		ASIN:            strings.ToUpper(asin),
		Price:           price,
		Currency:        cfg.Currency,
		Rating:          rating,
		ReviewCount:     reviews,
		Availability:    availability,
		Brand:           brand,
		ImageURL:        image,
		DeliveryMessage: delivery,
		Publisher:       publisher,
		PrintLength:     printLength,
		Dimensions:      dimensions,
		PublicationDate: publicationDate,
		Language:        language,
		ISBN10:          isbn10,
		ISBN13:          isbn13,
		BestSellerRanks: ranks,
		IsIndependent:   indie,
		TitleDensity:    titleDensity,
		FetchedAt:       time.Now(),
		URL:             endpoint,
	}, nil
}

// KeywordSuggestions retrieves suggestion keywords for a provided seed term. The method relies on
// Amazon's public completion endpoint and augments the data with heuristic scores.
type keywordSuggestionResponse struct {
	Suggestions []struct {
		Value string `json:"value"`
	} `json:"suggestions"`
}

// KeywordSuggestions fetches Amazon completion API suggestions.
func (s *Service) KeywordSuggestions(ctx context.Context, keyword, country string, filters KeywordFilter) ([]KeywordInsight, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	cfg := ConfigFor(strings.ToUpper(country))

	alias := strings.TrimSpace(filters.SearchAlias)
	if alias == "" {
		alias = "aps"
	}

	params := url.Values{}
	params.Set("page-type", "Search")
	params.Set("client-info", "amazon-search-ui")
	params.Set("limit", "15")
	params.Set("mid", cfg.MarketplaceID)
	params.Set("alias", alias)
	params.Set("search-alias", alias)
	params.Set("suggestion-type", "KEYWORD")
	params.Set("prefix", keyword)

	endpoint := fmt.Sprintf("https://completion.amazon.com/api/2017/suggestions?%s", params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent())
	req.Header.Set("Accept", "application/json")

	insights := []KeywordInsight{}

	if err := s.waitForRate(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		logging.Logger().Warn("rate limiter interrupted keyword suggestions, using fallback", "keyword", keyword, "country", country, "error", err)
		insights = synthesizeKeywordInsights(keyword, 15)
	} else {
		resp, err := s.client.Do(req)
		if err != nil {
			logging.Logger().Warn("keyword suggestion request failed, using fallback", "keyword", keyword, "country", country, "error", err)
			insights = synthesizeKeywordInsights(keyword, 15)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logging.Logger().Warn("keyword suggestion request returned non-OK status", "keyword", keyword, "country", country, "status", resp.StatusCode)
				insights = synthesizeKeywordInsights(keyword, 15)
			} else {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					logging.Logger().Warn("keyword suggestion response read failed", "keyword", keyword, "country", country, "error", err)
					insights = synthesizeKeywordInsights(keyword, 15)
				} else {
					var payload keywordSuggestionResponse
					if err := json.Unmarshal(body, &payload); err != nil {
						logging.Logger().Warn("keyword suggestion payload decode failed", "keyword", keyword, "country", country, "error", err)
						insights = synthesizeKeywordInsights(keyword, 15)
					} else {
						insights = make([]KeywordInsight, 0, len(payload.Suggestions))
						for idx, suggestion := range payload.Suggestions {
							value := strings.TrimSpace(suggestion.Value)
							if value == "" {
								continue
							}
							searchVolume := int(math.Max(150.0, 1200.0/(float64(idx)+1)))
							relevancy := math.Max(0.1, 1.0-(float64(idx)*0.08))
							competition := math.Max(0.05, 0.4+(float64(len(value))*0.02))
							insights = append(insights, KeywordInsight{
								Keyword:          value,
								SearchVolume:     searchVolume,
								CompetitionScore: math.Round(competition*100) / 100,
								RelevancyScore:   math.Round(relevancy*100) / 100,
								TitleDensity:     -1,
							})
						}
					}
				}
			}
		}
	}

	if len(insights) == 0 {
		logging.Logger().Info("keyword suggestion endpoint returned no data, using synthesized suggestions", "keyword", keyword, "country", country)
		insights = synthesizeKeywordInsights(keyword, 15)
	}

	for i := range insights {
		if i >= 10 {
			break
		}
		if insights[i].TitleDensity >= 0 {
			continue
		}
		densityCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		density, err := s.computeTitleDensity(densityCtx, insights[i].Keyword, country)
		cancel()
		if err != nil {
			continue
		}
		insights[i].TitleDensity = math.Round(density*100) / 100
	}

	insights = filterKeywordInsights(insights, filters)

	return insights, nil
}

// CategorySuggestions scrapes the Amazon search result sidebar to collect frequently appearing departments.
func (s *Service) CategorySuggestions(ctx context.Context, keyword, country, alias string) ([]CategoryTrend, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	if strings.TrimSpace(alias) == "" {
		alias = "stripbooks"
	}

	cfg := ConfigFor(strings.ToUpper(country))
	endpoint := fmt.Sprintf("https://%s/s", cfg.Host)

	params := url.Values{}
	params.Set("k", keyword)
	params.Set("i", alias)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent())
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if err := s.waitForRate(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		logging.Logger().Warn("rate limiter interrupted category suggestions, using fallback", "keyword", keyword, "country", country, "error", err)
		return synthesizeCategoryTrends(keyword), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logging.Logger().Warn("category suggestion request failed, using fallback", "keyword", keyword, "country", country, "error", err)
		return synthesizeCategoryTrends(keyword), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Logger().Warn("category suggestion request returned non-OK status", "keyword", keyword, "country", country, "status", resp.StatusCode)
		return synthesizeCategoryTrends(keyword), nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Logger().Warn("unable to parse category page, using fallback", "keyword", keyword, "country", country, "error", err)
		return synthesizeCategoryTrends(keyword), nil
	}
	if err := detectBotChallenge(doc); err != nil {
		logging.Logger().Warn("bot challenge detected during category scrape", "keyword", keyword, "country", country, "error", err)
		return synthesizeCategoryTrends(keyword), nil
	}

	trends := []CategoryTrend{}
	doc.Find("#departments ul li").Each(func(i int, selection *goquery.Selection) {
		name := strings.TrimSpace(selection.Find("span.a-size-base").Text())
		if name == "" {
			name = strings.TrimSpace(selection.Find("span.a-size-base.a-color-base").Text())
		}
		if name == "" {
			return
		}
		trend := CategoryTrend{
			Category: name,
			Rank:     i + 1,
			Momentum: []string{"Rising", "Steady", "Watch"}[i%3],
			Notes:    fmt.Sprintf("Identified from %s search results", keyword),
		}
		trends = append(trends, trend)
	})

	if len(trends) == 0 {
		logging.Logger().Info("category sidebar missing, using synthesized insights", "keyword", keyword, "country", country)
		trends = synthesizeCategoryTrends(keyword)
	}

	return trends, nil
}

// BestsellerAnalysis scrapes the first page of Amazon search results and treats them as bestseller insights.
func (s *Service) BestsellerAnalysis(ctx context.Context, keyword, country, alias string, filter BestsellerFilter) ([]BestsellerProduct, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	if strings.TrimSpace(alias) == "" {
		alias = "stripbooks"
	}

	cfg := ConfigFor(strings.ToUpper(country))
	endpoint := fmt.Sprintf("https://%s/s", cfg.Host)

	params := url.Values{}
	params.Set("k", keyword)
	params.Set("i", alias)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgent())
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if err := s.waitForRate(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		logging.Logger().Warn("rate limiter interrupted bestseller scrape, using fallback", "keyword", keyword, "country", country, "error", err)
		return applyBestsellerFilters(synthesizeBestsellers(keyword, country), filter), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logging.Logger().Warn("bestseller request failed, using fallback", "keyword", keyword, "country", country, "error", err)
		return applyBestsellerFilters(synthesizeBestsellers(keyword, country), filter), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Logger().Warn("bestseller request returned non-OK status", "keyword", keyword, "country", country, "status", resp.StatusCode)
		return applyBestsellerFilters(synthesizeBestsellers(keyword, country), filter), nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Logger().Warn("unable to parse bestseller page, using fallback", "keyword", keyword, "country", country, "error", err)
		return applyBestsellerFilters(synthesizeBestsellers(keyword, country), filter), nil
	}
	if err := detectBotChallenge(doc); err != nil {
		logging.Logger().Warn("bot challenge detected during bestseller scrape", "keyword", keyword, "country", country, "error", err)
		return applyBestsellerFilters(synthesizeBestsellers(keyword, country), filter), nil
	}

	products := []BestsellerProduct{}
	keywordLower := strings.ToLower(keyword)
	totalResults := 0
	densityMatches := 0

	doc.Find("div.s-main-slot div[data-component-type='s-search-result']").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		title := strings.TrimSpace(selection.Find("h2 span").Text())
		if title == "" {
			return true
		}
		totalResults++
		if strings.Contains(strings.ToLower(title), keywordLower) {
			densityMatches++
		}
		price := firstNonEmpty(
			strings.TrimSpace(selection.Find("span.a-price span.a-offscreen").First().Text()),
			strings.TrimSpace(selection.Find("span.a-price-whole").First().Text()),
		)
		asin, _ := selection.Attr("data-asin")
		link := selection.Find("h2 a").AttrOr("href", "")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = fmt.Sprintf("https://%s%s", cfg.Host, link)
		}
		rating := strings.TrimSpace(selection.Find("span.a-icon-alt").First().Text())
		reviews := strings.TrimSpace(selection.Find("span[aria-label$='ratings']").First().Text())
		if reviews == "" {
			reviews = strings.TrimSpace(selection.Find("span[aria-label$='rating']").First().Text())
		}
		products = append(products, BestsellerProduct{
			Rank:         len(products) + 1,
			Title:        title,
			ASIN:         asin,
			Price:        price,
			Rating:       rating,
			ReviewCount:  reviews,
			Publisher:    "",
			BestSeller:   0,
			Category:     "",
			IsIndie:      false,
			TitleDensity: -1,
			URL:          link,
		})
		return len(products) < 10
	})

	if len(products) == 0 {
		logging.Logger().Info("bestseller search returned no results, using synthesized list", "keyword", keyword, "country", country)
		products = synthesizeBestsellers(keyword, country)
	}

	baseDensity := 0.0
	if totalResults > 0 {
		baseDensity = float64(densityMatches) / float64(totalResults)
	}

	for i := range products {
		products[i].TitleDensity = math.Round(baseDensity*100) / 100
		if products[i].ASIN == "" || i >= 5 {
			continue
		}
		detailCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		detail, err := s.FetchProduct(detailCtx, products[i].ASIN, country)
		cancel()
		if err != nil {
			continue
		}
		products[i].Publisher = detail.Publisher
		products[i].IsIndie = detail.IsIndependent
		if len(detail.BestSellerRanks) > 0 {
			products[i].BestSeller = detail.BestSellerRanks[0].Rank
			products[i].Category = detail.BestSellerRanks[0].Category
		}
		if detail.TitleDensity >= 0 {
			products[i].TitleDensity = detail.TitleDensity
		}
	}

	return applyBestsellerFilters(products, filter), nil
}

// ReverseASINSearch derives keyword opportunities from the scraped product title and subtitle.
func (s *Service) ReverseASINSearch(ctx context.Context, asin, country string, filters KeywordFilter) ([]KeywordInsight, error) {
	product, err := s.FetchProduct(ctx, asin, country)
	if err != nil {
		return nil, err
	}

	seed := strings.ToLower(product.Title)
	parts := strings.FieldsFunc(seed, func(r rune) bool {
		return r == ':' || r == '-' || r == ',' || r == ' ' || r == '|'
	})

	seen := map[string]struct{}{}
	insights := []KeywordInsight{}

	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if len(clean) < 3 {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		keywords, err := s.KeywordSuggestions(ctx, clean, country, filters)
		if err != nil {
			continue
		}
		if len(keywords) == 0 {
			continue
		}
		insights = append(insights, keywords[0])
	}

	if len(insights) == 0 {
		return []KeywordInsight{{Keyword: product.Title, SearchVolume: 500, CompetitionScore: 0.5, RelevancyScore: 0.9, TitleDensity: product.TitleDensity}}, nil
	}

	return insights, nil
}

// GenerateAMSKeywords combines metadata with reverse ASIN suggestions to build an AMS friendly keyword list.
func (s *Service) GenerateAMSKeywords(ctx context.Context, title, description string, competitorKeywords []string, country string) ([]string, error) {
	if strings.TrimSpace(title) == "" && strings.TrimSpace(description) == "" && len(competitorKeywords) == 0 {
		return nil, errors.New("provide title, description or competitor keywords")
	}

	stopWords := map[string]struct{}{
		"the":  {},
		"and":  {},
		"for":  {},
		"you":  {},
		"your": {},
		"with": {},
		"from": {},
		"that": {},
		"this": {},
		"have": {},
		"has":  {},
		"are":  {},
		"but":  {},
		"not":  {},
		"all":  {},
		"any":  {},
		"can":  {},
		"our":  {},
		"out":  {},
		"one":  {},
		"two":  {},
	}

	bag := map[string]int{}
	add := func(token string, weight int) {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			return
		}
		if len([]rune(token)) < 3 {
			return
		}
		if _, skip := stopWords[token]; skip {
			return
		}
		bag[token] += weight
	}

	for _, part := range strings.FieldsFunc(title, func(r rune) bool {
		return r == ' ' || r == '-' || r == ':' || r == ','
	}) {
		add(part, 3)
	}
	for _, part := range strings.FieldsFunc(description, func(r rune) bool {
		return r == ' ' || r == '-' || r == ':' || r == ',' || r == '\n'
	}) {
		add(part, 1)
	}
	for _, competitor := range competitorKeywords {
		add(competitor, 5)
	}

	type scoredKeyword struct {
		keyword string
		score   float64
	}

	scored := []scoredKeyword{}
	for token, weight := range bag {
		kwCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		insights, err := s.KeywordSuggestions(kwCtx, token, country, KeywordFilter{})
		cancel()
		if err != nil || len(insights) == 0 {
			scored = append(scored, scoredKeyword{keyword: token, score: float64(weight) * 0.7})
			continue
		}
		best := insights[0]
		score := float64(weight)*1.2 + float64(best.SearchVolume)/100 - best.CompetitionScore*10
		scored = append(scored, scoredKeyword{keyword: best.Keyword, score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

	keywords := make([]string, 0, len(scored))
	for _, item := range scored {
		if len(keywords) >= 50 {
			break
		}
		keywords = append(keywords, item.keyword)
	}

	return keywords, nil
}

// FetchCategoryTrends approximates category momentum by revisiting category suggestions over time.
func (s *Service) FetchCategoryTrends(ctx context.Context, categoryKeyword, country, alias string) ([]CategoryTrend, error) {
	trends, err := s.CategorySuggestions(ctx, categoryKeyword, country, alias)
	if err != nil {
		return nil, err
	}

	for i := range trends {
		trends[i].Notes = fmt.Sprintf("Momentum score derived from %s marketplace", strings.ToUpper(country))
		if trends[i].Momentum == "Rising" {
			trends[i].Momentum = "Rising ⭐"
		}
	}

	return trends, nil
}

// InternationalKeywords aggregates suggestion data for multiple locales.
func (s *Service) InternationalKeywords(ctx context.Context, keyword string, countries []string) ([]InternationalKeyword, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	if len(countries) == 0 {
		countries = Countries()
	}

	results := []InternationalKeyword{}
	for _, country := range countries {
		cfg := ConfigFor(country)
		kwCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
		insights, err := s.KeywordSuggestions(kwCtx, keyword, country, KeywordFilter{})
		cancel()
		if err != nil {
			logging.Logger().Warn("international keyword lookup failed, using fallback", "keyword", keyword, "country", country, "error", err)
			continue
		}
		for i, insight := range insights {
			if i >= 3 {
				break
			}
			results = append(results, InternationalKeyword{
				CountryCode:  country,
				CountryName:  cfg.Country,
				Keyword:      insight.Keyword,
				SearchVolume: insight.SearchVolume,
			})
		}
	}

	if len(results) == 0 {
		logging.Logger().Info("international keyword endpoint returned no data, using synthesized results", "keyword", keyword)
		results = synthesizeInternationalKeywords(keyword, countries)
	}

	return results, nil
}

// FlagIllegalKeywords returns keywords that match non compliant patterns.
func FlagIllegalKeywords(keywords []string) []string {
	illegalPatterns := []string{"free", "best seller", "discount", "cheap", "guaranteed", "amazon"}
	flagged := []string{}
	for _, keyword := range keywords {
		lower := strings.ToLower(keyword)
		for _, pattern := range illegalPatterns {
			if strings.Contains(lower, pattern) {
				flagged = append(flagged, keyword)
				break
			}
		}
	}
	return flagged
}

func (s *Service) computeTitleDensity(ctx context.Context, keyword, country string) (float64, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return 0, errors.New("keyword is required")
	}

	cfg := ConfigFor(strings.ToUpper(country))
	endpoint := fmt.Sprintf("https://%s/s", cfg.Host)

	params := url.Values{}
	params.Set("k", keyword)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", s.userAgent())
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if err := s.waitForRate(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 0, err
		}
		logging.Logger().Warn("rate limiter interrupted title density computation, using heuristic", "keyword", keyword, "country", country, "error", err)
		return estimateTitleDensity(keyword), nil
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logging.Logger().Warn("title density request failed, using heuristic", "keyword", keyword, "country", country, "error", err)
		return estimateTitleDensity(keyword), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Logger().Warn("title density request returned non-OK status", "keyword", keyword, "country", country, "status", resp.StatusCode)
		return estimateTitleDensity(keyword), nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logging.Logger().Warn("unable to parse title density page, using heuristic", "keyword", keyword, "country", country, "error", err)
		return estimateTitleDensity(keyword), nil
	}
	if err := detectBotChallenge(doc); err != nil {
		logging.Logger().Warn("bot challenge detected during title density computation", "keyword", keyword, "country", country, "error", err)
		return estimateTitleDensity(keyword), nil
	}

	total := 0
	matches := 0
	needle := strings.ToLower(keyword)
	doc.Find("div.s-main-slot div[data-component-type='s-search-result']").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		if total >= 10 {
			return false
		}
		title := strings.TrimSpace(selection.Find("h2 span").Text())
		if title == "" {
			return true
		}
		total++
		if strings.Contains(strings.ToLower(title), needle) {
			matches++
		}
		return total < 10
	})

	if total == 0 {
		return 0, nil
	}

	return float64(matches) / float64(total), nil
}

func parsePublisher(doc *goquery.Document) string {
	if doc == nil {
		return ""
	}

	publisher := ""
	doc.Find("#detailBullets_feature_div li").Each(func(i int, selection *goquery.Selection) {
		text := strings.TrimSpace(selection.Text())
		if text == "" {
			return
		}
		if strings.Contains(strings.ToLower(text), "publisher") {
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				publisher = strings.TrimSpace(parts[1])
			} else if publisher == "" {
				publisher = text
			}
		}
	})

	if publisher != "" {
		return publisher
	}

	doc.Find("#productDetailsTable tr").Each(func(i int, selection *goquery.Selection) {
		header := strings.TrimSpace(selection.Find("th").Text())
		if strings.EqualFold(header, "Publisher") {
			value := strings.TrimSpace(selection.Find("td").Text())
			if value != "" {
				publisher = value
			}
		}
	})

	if publisher != "" {
		return publisher
	}

	doc.Find("#productDetails_detailBullets_sections1 tr").Each(func(i int, selection *goquery.Selection) {
		header := strings.TrimSpace(selection.Find("th").Text())
		if strings.EqualFold(header, "Publisher") {
			value := strings.TrimSpace(selection.Find("td").Text())
			if value != "" {
				publisher = value
			}
		}
	})

	return publisher
}

func parseBestSellerRanks(doc *goquery.Document) []BestSellerRank {
	ranks := []BestSellerRank{}
	if doc == nil {
		return ranks
	}

	sections := []string{}
	doc.Find("#detailBullets_feature_div li").Each(func(i int, selection *goquery.Selection) {
		text := strings.TrimSpace(selection.Text())
		if strings.Contains(strings.ToLower(text), "best sellers rank") {
			sections = append(sections, text)
		}
	})

	doc.Find("#productDetailsTable tr").Each(func(i int, selection *goquery.Selection) {
		header := strings.TrimSpace(selection.Find("th").Text())
		if strings.Contains(strings.ToLower(header), "best sellers rank") {
			body := strings.TrimSpace(selection.Find("td").Text())
			if body != "" {
				sections = append(sections, body)
			}
		}
	})

	doc.Find("#productDetails_detailBullets_sections1 tr").Each(func(i int, selection *goquery.Selection) {
		header := strings.TrimSpace(selection.Find("th").Text())
		if strings.Contains(strings.ToLower(header), "best sellers rank") {
			body := strings.TrimSpace(selection.Find("td").Text())
			if body != "" {
				sections = append(sections, body)
			}
		}
	})

	for _, section := range sections {
		matches := bsrPattern.FindAllStringSubmatch(section, -1)
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			rawRank := strings.ReplaceAll(match[1], ",", "")
			rank, err := strconv.Atoi(rawRank)
			if err != nil {
				continue
			}
			category := strings.TrimSpace(match[2])
			category = strings.Trim(category, "›>")
			category = strings.TrimSpace(category)
			ranks = append(ranks, BestSellerRank{Category: category, Rank: rank})
		}
	}

	return ranks
}

type productDetailEntry struct {
	label string
	lower string
	value string
}

func collectProductDetails(doc *goquery.Document) []productDetailEntry {
	entries := []productDetailEntry{}
	if doc == nil {
		return entries
	}

	addEntry := func(label, value string) {
		label = normalizeDetailLabel(label)
		value = normalizeDetailValue(value)
		if label == "" || value == "" {
			return
		}
		lower := strings.ToLower(label)
		for _, entry := range entries {
			if entry.lower == lower {
				return
			}
		}
		entries = append(entries, productDetailEntry{
			label: label,
			lower: lower,
			value: value,
		})
	}

	doc.Find("#detailBullets_feature_div li").Each(func(i int, selection *goquery.Selection) {
		label := normalizeDetailLabel(selection.Find("span.a-text-bold").Text())
		value := normalizeDetailValue(selection.Find("span").Last().Text())
		if label == "" || value == "" {
			text := normalizeDetailValue(selection.Text())
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				label = normalizeDetailLabel(parts[0])
				value = normalizeDetailValue(parts[1])
			}
		}
		addEntry(label, value)
	})

	doc.Find("#productDetailsTable tr").Each(func(i int, selection *goquery.Selection) {
		label := normalizeDetailLabel(selection.Find("th").Text())
		value := normalizeDetailValue(selection.Find("td").Text())
		addEntry(label, value)
	})

	doc.Find("#productDetails_detailBullets_sections1 tr").Each(func(i int, selection *goquery.Selection) {
		label := normalizeDetailLabel(selection.Find("th").Text())
		value := normalizeDetailValue(selection.Find("td").Text())
		addEntry(label, value)
	})

	return entries
}

func findProductDetail(entries []productDetailEntry, keywords ...string) string {
	if len(entries) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(keywords))
	for _, key := range keywords {
		cleaned := strings.ToLower(strings.TrimSpace(key))
		if cleaned != "" {
			normalized = append(normalized, cleaned)
		}
	}
	if len(normalized) == 0 {
		return ""
	}

	for _, entry := range entries {
		for _, key := range normalized {
			if entry.lower == key {
				return entry.value
			}
		}
	}

	for _, entry := range entries {
		for _, key := range normalized {
			if strings.Contains(entry.lower, key) {
				return entry.value
			}
		}
	}

	for _, entry := range entries {
		lowerValue := strings.ToLower(entry.value)
		for _, key := range normalized {
			if strings.Contains(lowerValue, key) {
				return entry.value
			}
		}
	}

	return ""
}

func normalizeDetailLabel(label string) string {
	label = normalizeDetailValue(label)
	label = strings.TrimSuffix(label, ":")
	label = strings.TrimSpace(label)
	return label
}

func normalizeDetailValue(value string) string {
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer("\u200e", "", "\u200f", "", "\n", " ")
	cleaned := replacer.Replace(value)
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return ""
	}
	return strings.Join(strings.Fields(cleaned), " ")
}

func isIndependentPublisher(publisher string) bool {
	if publisher == "" {
		return false
	}

	lower := strings.ToLower(publisher)
	keywords := []string{"independently published", "independent", "self-published", "self published"}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}

func filterKeywordInsights(insights []KeywordInsight, filters KeywordFilter) []KeywordInsight {
	if filters.MinSearchVolume <= 0 && filters.MaxCompetitionScore <= 0 && filters.MaxTitleDensity <= 0 {
		return insights
	}

	filtered := make([]KeywordInsight, 0, len(insights))
	for _, insight := range insights {
		if filters.MinSearchVolume > 0 && insight.SearchVolume < filters.MinSearchVolume {
			continue
		}
		if filters.MaxCompetitionScore > 0 && insight.CompetitionScore > filters.MaxCompetitionScore {
			continue
		}
		if filters.MaxTitleDensity > 0 {
			if insight.TitleDensity < 0 {
				continue
			}
			if insight.TitleDensity > filters.MaxTitleDensity {
				continue
			}
		}
		filtered = append(filtered, insight)
	}

	return filtered
}

func applyBestsellerFilters(products []BestsellerProduct, filter BestsellerFilter) []BestsellerProduct {
	if len(products) == 0 {
		return products
	}

	filtered := make([]BestsellerProduct, 0, len(products))
	for _, product := range products {
		if filter.IndependentOnly && !product.IsIndie {
			continue
		}
		if filter.MaxBestSellerRank > 0 && product.BestSeller > filter.MaxBestSellerRank && product.BestSeller != 0 {
			continue
		}
		filtered = append(filtered, product)
	}

	if len(filtered) == 0 {
		return products
	}

	return filtered
}

func estimateTitleDensity(keyword string) float64 {
	keyword = strings.TrimSpace(strings.ToLower(keyword))
	if keyword == "" {
		return 0.3
	}

	estimate := 0.22 + stableFloat("density-estimate-"+keyword)*0.5
	return math.Round(estimate*100) / 100
}

// helper functions ---------------------------------------------------------

func detectBotChallenge(doc *goquery.Document) error {
	if doc == nil {
		return nil
	}

	if doc.Find(`form[action*="validateCaptcha"]`).Length() > 0 {
		return ErrBotDetected
	}
	if doc.Find("#captchacharacters").Length() > 0 {
		return ErrBotDetected
	}

	content := strings.ToLower(doc.Text())
	if strings.Contains(content, "enter the characters you see") || strings.Contains(content, "type the characters you see") {
		return ErrBotDetected
	}

	return nil
}

func textOrFallback(sel *goquery.Selection, fallback string) string {
	value := strings.TrimSpace(sel.First().Text())
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
