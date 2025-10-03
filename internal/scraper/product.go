package scraper

import "time"

// ProductDetails represents the key data points extracted from an Amazon product page.
type ProductDetails struct {
	Title           string
	ASIN            string
	Price           string
	Currency        string
	Rating          string
	ReviewCount     string
	Availability    string
	ImageURL        string
	Brand           string
	DeliveryMessage string
	Publisher       string
	PrintLength     string
	Dimensions      string
	PublicationDate string
	Language        string
	ISBN10          string
	ISBN13          string
	BestSellerRanks []BestSellerRank
	IsIndependent   bool
	TitleDensity    float64
	FetchedAt       time.Time
	URL             string
}

// BestSellerRank represents a single category ranking entry extracted from the
// Amazon product detail page.
type BestSellerRank struct {
	Category string
	Rank     int
}

// BestsellerProduct describes a product highlighted during bestseller analysis.
type BestsellerProduct struct {
	Rank         int
	Title        string
	ASIN         string
	Price        string
	Rating       string
	ReviewCount  string
	Publisher    string
	BestSeller   int
	Category     string
	IsIndie      bool
	TitleDensity float64
	URL          string
}

// KeywordInsight captures keyword research related metrics.
type KeywordInsight struct {
	Keyword          string
	SearchVolume     int
	CompetitionScore float64
	RelevancyScore   float64
	TitleDensity     float64
}

// CategoryTrend contains aggregated category insights.
type CategoryTrend struct {
	Category string
	Rank     int
	Momentum string
	Notes    string
}

// InternationalKeyword describes keyword suggestions for specific markets.
type InternationalKeyword struct {
	CountryCode  string
	CountryName  string
	Keyword      string
	SearchVolume int
}

// KeywordFilter defines optional threshold filters applied to keyword metrics.
type KeywordFilter struct {
	MinSearchVolume     int
	MaxCompetitionScore float64
	MaxTitleDensity     float64
	SearchAlias         string
}

// BestsellerFilter describes filters for bestseller search results.
type BestsellerFilter struct {
	MaxBestSellerRank int
	IndependentOnly   bool
}
