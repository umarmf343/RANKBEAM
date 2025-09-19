package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
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

	fetchInFlight := false
	var updateButton func()
	fetchButton := widget.NewButton("Fetch Product", func() {
		asin := strings.ToUpper(strings.TrimSpace(asinEntry.Text))
		if asin != asinEntry.Text {
			asinEntry.SetText(asin)
		}
		if !isValidASIN(asin) {
			output.Set("Enter a valid 10-character ASIN to fetch product details.")
			return
		}

		country := countrySelect.Selected
		fetchInFlight = true
		fetchButton.Disable()
		output.Set(fmt.Sprintf("Fetching product %s from %s...", asin, country))

		go func() {
			defer fyne.CurrentApp().Driver().RunOnMain(func() {
				fetchInFlight = false
				updateButton()
			})

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			product, err := service.FetchProduct(ctx, asin, country)
			if err != nil {
				output.Set(fmt.Sprintf("Error: %v", err))
				return
			}
			output.Set(formatProductDetails(product))
		}()
	})
	fetchButton.Disable()

	updateButton = func() {
		updateProductButtonState(asinEntry, fetchButton, fetchInFlight)
	}
	attachTrimHandler(asinEntry, updateButton)
	updateButton()

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

	researchInFlight := false
	var updateResearch func()
	fetchButton := widget.NewButton("Run Research", func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		if keyword != keywordEntry.Text {
			keywordEntry.SetText(keyword)
		}
		if keyword == "" {
			keywordOutput.Set("Enter a keyword to run research.")
			categoryOutput.Set("Enter a keyword to run research.")
			bestsellerOutput.Set("Enter a keyword to run research.")
			return
		}

		country := countrySelect.Selected
		researchInFlight = true
		fetchButton.Disable()
		keywordOutput.Set(fmt.Sprintf("Fetching keyword suggestions for %s...", keyword))

		go func() {
			defer fyne.CurrentApp().Driver().RunOnMain(func() {
				researchInFlight = false
				updateResearch()
			})

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			suggestions, err := service.KeywordSuggestions(ctx, keyword, country)
			if err != nil {
				keywordOutput.Set(fmt.Sprintf("Error: %v", err))
			} else {
				keywordOutput.Set(formatKeywordInsights(keyword, suggestions))
			}

			categories, err := service.CategorySuggestions(ctx, keyword, country)
			if err != nil {
				categoryOutput.Set(fmt.Sprintf("Error: %v", err))
			} else {
				categoryOutput.Set(formatCategoryTrends(categories))
			}

			bestsellers, err := service.BestsellerAnalysis(ctx, keyword, country)
			if err != nil {
				bestsellerOutput.Set(fmt.Sprintf("Error: %v", err))
			} else {
				bestsellerOutput.Set(formatBestsellerProducts(bestsellers))
			}
		}()
	})
	fetchButton.Disable()

	updateResearch = func() {
		updateResearchButtonState(keywordEntry, fetchButton, researchInFlight)
	}
	attachTrimHandler(keywordEntry, updateResearch)
	updateResearch()

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

	reverseInFlight := false
	var updateReverse func()
	reverseButton := widget.NewButton("Reverse ASIN Search", func() {
		asin := strings.ToUpper(strings.TrimSpace(reverseAsinEntry.Text))
		if asin != reverseAsinEntry.Text {
			reverseAsinEntry.SetText(asin)
		}
		if !isValidASIN(asin) {
			reverseOutput.Set("Enter a valid ASIN to run a reverse search.")
			return
		}

		country := countrySelect.Selected
		reverseInFlight = true
		reverseButton.Disable()
		reverseOutput.Set(fmt.Sprintf("Running reverse ASIN search for %s...", asin))

		go func() {
			defer fyne.CurrentApp().Driver().RunOnMain(func() {
				reverseInFlight = false
				updateReverse()
			})

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			insights, err := service.ReverseASINSearch(ctx, asin, country)
			if err != nil {
				reverseOutput.Set(fmt.Sprintf("Error: %v", err))
				return
			}
			reverseOutput.Set(formatKeywordInsights(fmt.Sprintf("ASIN %s", asin), insights))
		}()
	})
	reverseButton.Disable()

	campaignInFlight := false
	var updateCampaign func()
	campaignButton := widget.NewButton("Generate AMS Keywords", func() {
		title := strings.TrimSpace(titleEntry.Text)
		if title != titleEntry.Text {
			titleEntry.SetText(title)
		}
		description := strings.TrimSpace(descriptionEntry.Text)
		if description == "" {
			campaignOutput.Set("Provide a title and description to generate AMS keywords.")
			return
		}

		country := countrySelect.Selected
		competitors := cleanCompetitorKeywords(competitorKeywordsEntry.Text)
		campaignInFlight = true
		campaignButton.Disable()
		campaignOutput.Set("Generating keyword list...")

		go func() {
			defer fyne.CurrentApp().Driver().RunOnMain(func() {
				campaignInFlight = false
				updateCampaign()
			})

			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			keywords, err := service.GenerateAMSKeywords(ctx, title, description, competitors, country)
			if err != nil {
				campaignOutput.Set(fmt.Sprintf("Error: %v", err))
				return
			}
			flagged := scraper.FlagIllegalKeywords(keywords)
			campaignOutput.Set(formatCampaignKeywords(keywords, flagged))
		}()
	})
	campaignButton.Disable()

	updateReverse = func() {
		updateReverseButtonState(reverseAsinEntry, reverseButton, reverseInFlight)
	}
	attachTrimHandler(reverseAsinEntry, updateReverse)
	updateReverse()

	updateCampaign = func() {
		updateCampaignButtonState(titleEntry, descriptionEntry, campaignButton, campaignInFlight)
	}
	descriptionEntry.OnChanged = func(string) {
		updateCampaign()
	}
	competitorKeywordsEntry.OnChanged = func(string) {
		// no-op for trimming, but ensure button state reacts to edits if description cleared
		updateCampaign()
	}
	attachTrimHandler(titleEntry, updateCampaign)
	updateCampaign()

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

	internationalInFlight := false
	var updateInternational func()
	fetchButton := widget.NewButton("Collect International Keywords", func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		if keyword != keywordEntry.Text {
			keywordEntry.SetText(keyword)
		}
		if keyword == "" {
			output.Set("Enter a keyword to collect international keyword data.")
			return
		}

		selected := countrySelect.Selected
		if len(selected) == 0 {
			selected = countries
		}

		internationalInFlight = true
		fetchButton.Disable()
		output.Set("Collecting international keyword data...")

		go func() {
			defer fyne.CurrentApp().Driver().RunOnMain(func() {
				internationalInFlight = false
				updateInternational()
			})

			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			keywords, err := service.InternationalKeywords(ctx, keyword, selected)
			if err != nil {
				output.Set(fmt.Sprintf("Error: %v", err))
				return
			}
			output.Set(formatInternationalKeywords(keywords))
		}()
	})
	fetchButton.Disable()

	updateInternational = func() {
		updateInternationalButtonState(keywordEntry, fetchButton, internationalInFlight)
	}
	attachTrimHandler(keywordEntry, updateInternational)
	updateInternational()

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Keyword", keywordEntry),
		),
		widget.NewLabel("Select markets (leave blank for all)"),
		countrySelect,
	)

	return container.NewBorder(form, fetchButton, nil, nil, outputView)
}

func attachTrimHandler(entry *widget.Entry, update func()) {
	entry.OnChanged = func(s string) {
		trimmed := strings.TrimSpace(s)
		if !entry.MultiLine && trimmed != s {
			entry.SetText(trimmed)
			return
		}
		if update != nil {
			update()
		}
	}
	entry.OnSubmitted = func(s string) {
		trimmed := strings.TrimSpace(s)
		if trimmed != s {
			entry.SetText(trimmed)
			return
		}
		if update != nil {
			update()
		}
	}
}

func updateProductButtonState(entry *widget.Entry, button *widget.Button, inFlight bool) {
	if inFlight {
		button.Disable()
		return
	}

	asin := strings.ToUpper(strings.TrimSpace(entry.Text))
	if asin != entry.Text {
		entry.SetText(asin)
		return
	}

	if isValidASIN(asin) {
		button.Enable()
	} else {
		button.Disable()
	}
}

func updateResearchButtonState(entry *widget.Entry, button *widget.Button, inFlight bool) {
	if inFlight {
		button.Disable()
		return
	}

	keyword := strings.TrimSpace(entry.Text)
	if keyword != entry.Text {
		entry.SetText(keyword)
		return
	}

	if keyword == "" {
		button.Disable()
		return
	}
	button.Enable()
}

func updateReverseButtonState(entry *widget.Entry, button *widget.Button, inFlight bool) {
	if inFlight {
		button.Disable()
		return
	}

	asin := strings.ToUpper(strings.TrimSpace(entry.Text))
	if asin != entry.Text {
		entry.SetText(asin)
		return
	}

	if isValidASIN(asin) {
		button.Enable()
	} else {
		button.Disable()
	}
}

func updateCampaignButtonState(titleEntry, descriptionEntry *widget.Entry, button *widget.Button, inFlight bool) {
	if inFlight {
		button.Disable()
		return
	}

	title := strings.TrimSpace(titleEntry.Text)
	if title != titleEntry.Text {
		titleEntry.SetText(title)
		return
	}

	description := strings.TrimSpace(descriptionEntry.Text)
	if title == "" || description == "" {
		button.Disable()
		return
	}

	button.Enable()
}

func updateInternationalButtonState(entry *widget.Entry, button *widget.Button, inFlight bool) {
	if inFlight {
		button.Disable()
		return
	}

	keyword := strings.TrimSpace(entry.Text)
	if keyword != entry.Text {
		entry.SetText(keyword)
		return
	}

	if keyword == "" {
		button.Disable()
		return
	}

	button.Enable()
}

func cleanCompetitorKeywords(raw string) []string {
	sanitized := strings.ReplaceAll(raw, "\r", "")
	parts := strings.Split(sanitized, "\n")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}
	return results
}

var asinPattern = regexp.MustCompile(`^[A-Z0-9]{10}$`)

func isValidASIN(value string) bool {
	if value == "" {
		return false
	}
	upper := strings.ToUpper(value)
	return asinPattern.MatchString(upper)
}

func formatProductDetails(product *scraper.ProductDetails) string {
	return fmt.Sprintf(`Title: %s\nASIN: %s\nPrice: %s (%s)\nRating: %s\nReviews: %s\nAvailability: %s\nBrand: %s\nDelivery: %s\nURL: %s\nFetched: %s`,
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

func formatKeywordInsights(title string, insights []scraper.KeywordInsight) string {
	if len(insights) == 0 {
		return fmt.Sprintf("No keyword suggestions available for %s", title)
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Keyword Research for %s\n", title))
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	builder.WriteString("Keyword | Search Volume | Competition | Relevancy\n")
	builder.WriteString(strings.Repeat("-", 60))
	builder.WriteString("\n")
	for _, insight := range insights {
		builder.WriteString(fmt.Sprintf("%s | %d | %.2f | %.2f\n", insight.Keyword, insight.SearchVolume, insight.CompetitionScore, insight.RelevancyScore))
	}
	return builder.String()
}

func formatCategoryTrends(trends []scraper.CategoryTrend) string {
	builder := strings.Builder{}
	builder.WriteString("Category Intelligence\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	for _, trend := range trends {
		builder.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n", trend.Rank, trend.Category, trend.Momentum, trend.Notes))
	}
	return builder.String()
}

func formatBestsellerProducts(products []scraper.BestsellerProduct) string {
	builder := strings.Builder{}
	builder.WriteString("Bestseller Snapshot\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	for _, product := range products {
		builder.WriteString(fmt.Sprintf("#%d %s\n", product.Rank, product.Title))
		builder.WriteString(fmt.Sprintf("ASIN: %s | Price: %s | Rating: %s | Reviews: %s\n", product.ASIN, product.Price, product.Rating, product.ReviewCount))
		builder.WriteString(fmt.Sprintf("URL: %s\n\n", product.URL))
	}
	return builder.String()
}

func formatCampaignKeywords(keywords, flagged []string) string {
	if len(keywords) == 0 {
		return "Unable to generate keyword suggestions. Provide more metadata."
	}
	builder := strings.Builder{}
	builder.WriteString("Amazon Ads Keyword Portfolio\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	for i, keyword := range keywords {
		builder.WriteString(fmt.Sprintf("%02d. %s\n", i+1, keyword))
	}
	if len(flagged) > 0 {
		builder.WriteString("\n⚠️ Compliance Alerts:\n")
		for _, kw := range flagged {
			builder.WriteString(fmt.Sprintf("- %s\n", kw))
		}
	}
	return builder.String()
}

func formatInternationalKeywords(results []scraper.InternationalKeyword) string {
	builder := strings.Builder{}
	builder.WriteString("International Keyword Radar\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	for _, result := range results {
		builder.WriteString(fmt.Sprintf("%s (%s): %s [Search Volume: %d]\n", result.CountryName, result.CountryCode, result.Keyword, result.SearchVolume))
	}
	return builder.String()
}
