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
	FetchedAt       time.Time
	URL             string
}

// BestsellerProduct describes a product highlighted during bestseller analysis.
type BestsellerProduct struct {
	Rank        int
	Title       string
	ASIN        string
	Price       string
	Rating      string
	ReviewCount string
	URL         string
}

// KeywordInsight captures keyword research related metrics.
type KeywordInsight struct {
	Keyword          string
	SearchVolume     int
	CompetitionScore float64
	RelevancyScore   float64
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
