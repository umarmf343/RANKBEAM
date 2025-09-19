package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/umar/amazon-product-scraper/internal/scraper"
)

func main() {
	application := app.NewWithID("amazon-product-scraper")
	application.Settings().SetTheme(theme.LightTheme())

	service := scraper.NewService(25*time.Second, 25)
	defer service.Close()

	window := application.NewWindow("Amazon Product Intelligence Suite")
	window.Resize(fyne.NewSize(1024, 720))
	window.SetMaster()

	countries := scraper.Countries()
	sort.Strings(countries)

	productBinding := binding.NewString()
	productBinding.Set("Enter an ASIN and press Fetch Product to begin.")
	keywordBinding := binding.NewString()
	keywordBinding.Set("Keyword suggestions will appear here.")
	categoryBinding := binding.NewString()
	categoryBinding.Set("Category insights will appear here.")
	bestsellerBinding := binding.NewString()
	bestsellerBinding.Set("Bestseller analysis will appear here.")
	reverseBinding := binding.NewString()
	reverseBinding.Set("Reverse ASIN insights will appear here.")
	campaignBinding := binding.NewString()
	campaignBinding.Set("Generate Amazon Ads keywords to see results here.")
	internationalBinding := binding.NewString()
	internationalBinding.Set("International keyword suggestions will appear here.")

	tabs := container.NewAppTabs(
		container.NewTabItem("Product Lookup", buildProductLookupTab(service, countries, productBinding)),
		container.NewTabItem("Keyword Research", buildKeywordResearchTab(service, countries, keywordBinding, categoryBinding, bestsellerBinding)),
		container.NewTabItem("Competitive Analysis", buildCompetitiveTab(service, countries, reverseBinding, campaignBinding)),
		container.NewTabItem("International", buildInternationalTab(service, countries, internationalBinding)),
	)

	window.SetContent(tabs)
	window.ShowAndRun()
}

func buildProductLookupTab(service *scraper.Service, countries []string, output binding.String) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B09XYZ1234")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	resultView := widget.NewMultiLineEntry()
	resultView.Wrapping = fyne.TextWrapWord
	resultView.Bind(output)
	resultView.SetMinRowsVisible(12)
	resultView.Disable()

	fetchButton := widget.NewButton("Fetch Product", func() {
		asin := strings.TrimSpace(asinEntry.Text)
		country := countrySelect.Selected
		go func() {
			output.Set(fmt.Sprintf("Fetching product %s from %s...", asin, country))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			product, err := service.FetchProduct(ctx, asin, country)
			if err != nil {
				output.Set(renderScrapeError(err))
				return
			}
			output.Set(formatProductDetails(product))
		}()
	})

	form := container.New(layout.NewFormLayout(),
		widget.NewLabelWithStyle("ASIN", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), asinEntry,
		widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
	)

	content := container.NewBorder(nil, fetchButton, nil, nil,
		container.NewVBox(form, widget.NewSeparator(), resultView),
	)

	return container.NewPadded(content)
}

func buildKeywordResearchTab(service *scraper.Service, countries []string, keywordOutput, categoryOutput, bestsellerOutput binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("Enter a seed keyword e.g. self publishing")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	keywordView := widget.NewMultiLineEntry()
	keywordView.Wrapping = fyne.TextWrapWord
	keywordView.Bind(keywordOutput)
	keywordView.Disable()

	categoryView := widget.NewMultiLineEntry()
	categoryView.Wrapping = fyne.TextWrapWord
	categoryView.Bind(categoryOutput)
	categoryView.Disable()

	bestsellerView := widget.NewMultiLineEntry()
	bestsellerView.Wrapping = fyne.TextWrapWord
	bestsellerView.Bind(bestsellerOutput)
	bestsellerView.Disable()

	fetchButton := widget.NewButton("Run Research", func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		country := countrySelect.Selected
		go func() {
			keywordOutput.Set(fmt.Sprintf("Fetching keyword suggestions for %s...", keyword))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			suggestions, err := service.KeywordSuggestions(ctx, keyword, country)
			if err != nil {
				keywordOutput.Set(renderScrapeError(err))
			} else {
				formatted, _ := formatKeywordInsights(keyword, suggestions)
				keywordOutput.Set(formatted)
			}

			categories, err := service.CategorySuggestions(ctx, keyword, country)
			if err != nil {
				categoryOutput.Set(renderScrapeError(err))
			} else {
				formatted, _ := formatCategoryTrends(categories)
				categoryOutput.Set(formatted)
			}

			bestsellers, err := service.BestsellerAnalysis(ctx, keyword, country)
			if err != nil {
				bestsellerOutput.Set(renderScrapeError(err))
			} else {
				formatted, _ := formatBestsellerProducts(bestsellers)
				bestsellerOutput.Set(formatted)
			}
		}()
	})

	grid := container.NewGridWithRows(3,
		container.NewBorder(widget.NewLabelWithStyle("Keyword Suggestions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, keywordView),
		container.NewBorder(widget.NewLabelWithStyle("Category Opportunities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, categoryView),
		container.NewBorder(widget.NewLabelWithStyle("Bestseller Snapshot", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, bestsellerView),
	)

	form := container.New(layout.NewFormLayout(),
		widget.NewLabelWithStyle("Keyword", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), keywordEntry,
		widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
	)

	return container.NewBorder(container.NewVBox(form, widget.NewSeparator()), fetchButton, nil, nil, grid)
}

func buildCompetitiveTab(service *scraper.Service, countries []string, reverseOutput, campaignOutput binding.String) fyne.CanvasObject {
	reverseAsinEntry := widget.NewEntry()
	reverseAsinEntry.SetPlaceHolder("Competitor ASIN")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Book title for AMS keyword generation")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Product description / blurb")

	competitorKeywordsEntry := widget.NewMultiLineEntry()
	competitorKeywordsEntry.SetPlaceHolder("Paste competitor keywords (one per line)")

	reverseView := widget.NewMultiLineEntry()
	reverseView.Wrapping = fyne.TextWrapWord
	reverseView.Bind(reverseOutput)
	reverseView.Disable()

	campaignView := widget.NewMultiLineEntry()
	campaignView.Wrapping = fyne.TextWrapWord
	campaignView.Bind(campaignOutput)
	campaignView.Disable()

	reverseButton := widget.NewButton("Reverse ASIN Search", func() {
		asin := strings.TrimSpace(reverseAsinEntry.Text)
		country := countrySelect.Selected
		go func() {
			reverseOutput.Set(fmt.Sprintf("Running reverse ASIN search for %s...", asin))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			insights, err := service.ReverseASINSearch(ctx, asin, country)
			if err != nil {
				reverseOutput.Set(renderScrapeError(err))
				return
			}
			formatted, _ := formatKeywordInsights(fmt.Sprintf("ASIN %s", asin), insights)
			reverseOutput.Set(formatted)
		}()
	})

	campaignButton := widget.NewButton("Generate AMS Keywords", func() {
		country := countrySelect.Selected
		competitors := strings.Split(strings.ReplaceAll(competitorKeywordsEntry.Text, "\r", ""), "\n")
		go func() {
			campaignOutput.Set("Generating keyword list...")
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			keywords, err := service.GenerateAMSKeywords(ctx, titleEntry.Text, descriptionEntry.Text, competitors, country)
			if err != nil {
				campaignOutput.Set(renderScrapeError(err))
				return
			}
			flagged := scraper.FlagIllegalKeywords(keywords)
			formatted, _ := formatCampaignKeywords(keywords, flagged)
			campaignOutput.Set(formatted)
		}()
	})

	buttonRow := container.NewHBox(reverseButton, campaignButton)

	form := container.NewVBox(
		container.New(layout.NewFormLayout(),
			widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
			widget.NewLabelWithStyle("Reverse ASIN", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), reverseAsinEntry,
		),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Amazon Ads Planner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Description", descriptionEntry),
			widget.NewFormItem("Competitor Keywords", competitorKeywordsEntry),
		),
	)

	results := container.NewHSplit(
		container.NewBorder(widget.NewLabelWithStyle("Reverse ASIN Insights", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, reverseView),
		container.NewBorder(widget.NewLabelWithStyle("Keyword Portfolio", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), nil, nil, nil, campaignView),
	)
	results.SetOffset(0.5)

	return container.NewBorder(form, buttonRow, nil, nil, results)
}

func buildInternationalTab(service *scraper.Service, countries []string, output binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("Seed keyword e.g. coloring book")

	countrySelect := widget.NewCheckGroup(countries, nil)
	countrySelect.Horizontal = true

	outputView := widget.NewMultiLineEntry()
	outputView.Wrapping = fyne.TextWrapWord
	outputView.Bind(output)
	outputView.Disable()

	fetchButton := widget.NewButton("Collect International Keywords", func() {
		selected := countrySelect.Selected
		if len(selected) == 0 {
			selected = countries
		}
		keyword := keywordEntry.Text
		go func() {
			output.Set("Collecting international keyword data...")
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			keywords, err := service.InternationalKeywords(ctx, keyword, selected)
			if err != nil {
				output.Set(renderScrapeError(err))
				return
			}
			formatted, _ := formatInternationalKeywords(keywords)
			output.Set(formatted)
		}()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Keyword", keywordEntry),
		),
		widget.NewLabel("Select markets (leave blank for all)"),
		countrySelect,
	)

	return container.NewBorder(form, fetchButton, nil, nil, outputView)
}

func renderScrapeError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, scraper.ErrBotDetected) {
		return "Amazon is asking for a captcha/robot check. Please slow down, retry later, or rotate your network/proxy."
	}
	return fmt.Sprintf("Error: %v", err)
}

func formatProductDetails(product *scraper.ProductDetails) string {
	return fmt.Sprintf(`Title: %s
ASIN: %s
Price: %s (%s)
Rating: %s
Reviews: %s
Availability: %s
Brand: %s
Delivery: %s
URL: %s
Fetched: %s`,
		product.Title,
		product.ASIN,
		product.Price,
		product.Currency,
		product.Rating,
		product.ReviewCount,
		product.Availability,
		product.Brand,
		product.DeliveryMessage,
		product.URL,
		product.FetchedAt.Format(time.RFC1123),
	)
}

func formatKeywordInsights(title string, insights []scraper.KeywordInsight) (string, string) {
	if len(insights) == 0 {
		return fmt.Sprintf("No keyword suggestions available for %s", title), ""
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Keyword Research for %s\n", title))
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	builder.WriteString("Keyword | Search Volume | Competition | Relevancy\n")
	builder.WriteString(strings.Repeat("-", 60))
	builder.WriteString("\n")

	records := [][]string{{"Keyword", "Search Volume", "Competition", "Relevancy"}}

	for _, insight := range insights {
		builder.WriteString(fmt.Sprintf("%s | %d | %.2f | %.2f\n", insight.Keyword, insight.SearchVolume, insight.CompetitionScore, insight.RelevancyScore))
		records = append(records, []string{
			insight.Keyword,
			strconv.Itoa(insight.SearchVolume),
			fmt.Sprintf("%.2f", insight.CompetitionScore),
			fmt.Sprintf("%.2f", insight.RelevancyScore),
		})
	}

	return builder.String(), csvFromRecords(records)
}

func formatCategoryTrends(trends []scraper.CategoryTrend) (string, string) {
	builder := strings.Builder{}
	builder.WriteString("Category Intelligence\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")

	records := [][]string{{"Rank", "Category", "Momentum", "Notes"}}

	for _, trend := range trends {
		builder.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", trend.Rank, trend.Category, trend.Momentum, trend.Notes))
		records = append(records, []string{
			strconv.Itoa(trend.Rank),
			trend.Category,
			trend.Momentum,
			trend.Notes,
		})
	}

	return builder.String(), csvFromRecords(records)
}

func formatBestsellerProducts(products []scraper.BestsellerProduct) (string, string) {
	builder := strings.Builder{}
	builder.WriteString("Bestseller Snapshot\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")

	records := [][]string{{"Rank", "Title", "ASIN", "Price", "Rating", "Reviews", "URL"}}

	for _, product := range products {
		builder.WriteString(fmt.Sprintf("#%d %s\n", product.Rank, product.Title))
		builder.WriteString(fmt.Sprintf("ASIN: %s | Price: %s | Rating: %s | Reviews: %s\n", product.ASIN, product.Price, product.Rating, product.ReviewCount))
		builder.WriteString(fmt.Sprintf("URL: %s\n\n", product.URL))
		records = append(records, []string{
			strconv.Itoa(product.Rank),
			product.Title,
			product.ASIN,
			product.Price,
			product.Rating,
			product.ReviewCount,
			product.URL,
		})
	}

	return builder.String(), csvFromRecords(records)
}

func formatCampaignKeywords(keywords, flagged []string) (string, string) {
	if len(keywords) == 0 {
		return "Unable to generate keyword suggestions. Provide more metadata.", ""
	}

	builder := strings.Builder{}
	builder.WriteString("Amazon Ads Keyword Portfolio\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")

	records := [][]string{{"Index", "Keyword", "Flagged"}}

	flaggedSet := make(map[string]struct{}, len(flagged))
	for _, kw := range flagged {
		trimmed := strings.TrimSpace(kw)
		if trimmed != "" {
			flaggedSet[trimmed] = struct{}{}
		}
	}

	for i, keyword := range keywords {
		builder.WriteString(fmt.Sprintf("%02d. %s\n", i+1, keyword))

		flagValue := ""
		if _, exists := flaggedSet[strings.TrimSpace(keyword)]; exists {
			flagValue = "YES"
		}

		records = append(records, []string{
			strconv.Itoa(i + 1),
			keyword,
			flagValue,
		})
	}

	if len(flagged) > 0 {
		builder.WriteString("\n⚠️ Compliance Alerts:\n")
		for _, kw := range flagged {
			builder.WriteString(fmt.Sprintf("- %s\n", kw))
		}
	}

	return builder.String(), csvFromRecords(records)
}

func formatInternationalKeywords(results []scraper.InternationalKeyword) (string, string) {
	builder := strings.Builder{}
	builder.WriteString("International Keyword Radar\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")

	records := [][]string{{"Country Code", "Country", "Keyword", "Search Volume"}}

	for _, result := range results {
		builder.WriteString(fmt.Sprintf("%s (%s): %s [Search Volume: %d]\n", result.CountryName, result.CountryCode, result.Keyword, result.SearchVolume))
		records = append(records, []string{
			result.CountryCode,
			result.CountryName,
			result.Keyword,
			strconv.Itoa(result.SearchVolume),
		})
	}

	return builder.String(), csvFromRecords(records)
}

func csvFromRecords(records [][]string) string {
	if len(records) == 0 {
		return ""
	}

	builder := &strings.Builder{}
	writer := csv.NewWriter(builder)

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return ""
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return ""
	}

	return builder.String()
}
