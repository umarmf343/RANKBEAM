package scraper

import "strings"

// CountryConfig represents marketplace configuration for supported Amazon regions.
type CountryConfig struct {
	Country       string
	Currency      string
	Host          string
	MarketplaceID string
	MarketID      string
}

// countryConfigs is a curated subset of Amazon marketplaces that the application supports.
// The information is sourced from public documentation and open-source projects that interact
// with the Amazon websites. The MarketID value is used for the completion API while the
// MarketplaceID helps with other endpoints that expect the identifier.
var countryConfigs = map[string]CountryConfig{
	"US": {Country: "United States", Currency: "USD", Host: "www.amazon.com", MarketplaceID: "ATVPDKIKX0DER", MarketID: "1"},
	"CA": {Country: "Canada", Currency: "CAD", Host: "www.amazon.ca", MarketplaceID: "A2EUQ1WTGCTBG2", MarketID: "7"},
	"GB": {Country: "United Kingdom", Currency: "GBP", Host: "www.amazon.co.uk", MarketplaceID: "A1F83G8C2ARO7P", MarketID: "3"},
	"DE": {Country: "Germany", Currency: "EUR", Host: "www.amazon.de", MarketplaceID: "A1PA6795UKMFR9", MarketID: "4"},
	"FR": {Country: "France", Currency: "EUR", Host: "www.amazon.fr", MarketplaceID: "A13V1IB3VIYZZH", MarketID: "5"},
	"ES": {Country: "Spain", Currency: "EUR", Host: "www.amazon.es", MarketplaceID: "A1RKKUPIHCS9HS", MarketID: "44551"},
	"IT": {Country: "Italy", Currency: "EUR", Host: "www.amazon.it", MarketplaceID: "APJ6JRA9NG5V4", MarketID: "35691"},
	"IN": {Country: "India", Currency: "INR", Host: "www.amazon.in", MarketplaceID: "A21TJRUUN4KGV", MarketID: "44571"},
	"JP": {Country: "Japan", Currency: "JPY", Host: "www.amazon.co.jp", MarketplaceID: "A1VC38T7YXB528", MarketID: "6"},
	"AU": {Country: "Australia", Currency: "AUD", Host: "www.amazon.com.au", MarketplaceID: "A39IBJ37TRP1C6", MarketID: "111172"},
	"BR": {Country: "Brazil", Currency: "BRL", Host: "www.amazon.com.br", MarketplaceID: "A2Q3Y263D00KWC", MarketID: "193817"},
	"MX": {Country: "Mexico", Currency: "MXN", Host: "www.amazon.com.mx", MarketplaceID: "A1AM78C64UM0Y8", MarketID: "771770"},
	"AE": {Country: "United Arab Emirates", Currency: "AED", Host: "www.amazon.ae", MarketplaceID: "A2VIGQ35RCS4UG", MarketID: "162569"},
	"SG": {Country: "Singapore", Currency: "SGD", Host: "www.amazon.sg", MarketplaceID: "A19VAU5U5O7RUS", MarketID: "211896"},
}

var (
	// countryDisplayAlias maps canonical marketplace codes to their preferred
	// human-friendly variants for display inside the UI. The canonical code is
	// still used for scraping requests but we surface the alias to match what
	// merchants expect (for example "UK" instead of the ISO "GB").
	countryDisplayAlias = map[string]string{
		"GB": "UK",
	}

	// countryLookupAlias maps common aliases back to the canonical marketplace
	// code used by the scraper configuration.
	countryLookupAlias = map[string]string{
		"UK": "GB",
	}
)

// Countries returns the list of supported country codes.
func Countries() []string {
	codes := make([]string, 0, len(countryConfigs))
	for code := range countryConfigs {
		if alias, ok := countryDisplayAlias[code]; ok {
			codes = append(codes, alias)
			continue
		}
		codes = append(codes, code)
	}
	return codes
}

// ConfigFor returns the marketplace configuration for the provided ISO Alpha-2 code.
// When the country is unknown the function falls back to the United States marketplace.
func ConfigFor(country string) CountryConfig {
	normalized := strings.ToUpper(strings.TrimSpace(country))
	if canonical, ok := countryLookupAlias[normalized]; ok {
		normalized = canonical
	}
	if cfg, ok := countryConfigs[normalized]; ok {
		return cfg
	}
	return countryConfigs["US"]
}
