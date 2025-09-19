package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
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
		container.NewTabItem("Product Lookup", buildProductLookupTab(window, service, countries, productBinding)),
		container.NewTabItem("Keyword Research", buildKeywordResearchTab(window, service, countries, keywordBinding, categoryBinding, bestsellerBinding)),
		container.NewTabItem("Competitive Analysis", buildCompetitiveTab(window, service, countries, reverseBinding, campaignBinding)),
		container.NewTabItem("International", buildInternationalTab(window, service, countries, internationalBinding)),
	)

	window.SetContent(tabs)
	window.ShowAndRun()
}

const exportTimestampFormat = "20060102-150405"

func saveBindingToFile(win fyne.Window, text binding.String, csv binding.String, defaultName string) {
	current, err := text.Get()
	if err != nil {
		dialog.ShowError(err, win)
		return
	}
	if strings.TrimSpace(current) == "" {
		dialog.ShowInformation("No Data", "There is nothing to export yet.", win)
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		content := current
		if csv != nil {
			ext := strings.ToLower(filepath.Ext(writer.URI().Name()))
			if ext == ".csv" {
				csvContent, csvErr := csv.Get()
				if csvErr != nil {
					dialog.ShowError(csvErr, win)
					return
				}
				if strings.TrimSpace(csvContent) != "" {
					content = csvContent
				}
			}
		}

		if _, writeErr := writer.Write([]byte(content)); writeErr != nil {
			dialog.ShowError(writeErr, win)
		}
	}, win)

	if csv != nil {
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".csv"}))
	} else {
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
	}
	fileDialog.SetFileName(defaultName)
	fileDialog.Show()
}

func sanitizeFilenamePart(part string) string {
	part = strings.TrimSpace(part)
	if part == "" {
		return ""
	}

	var builder strings.Builder
	lastUnderscore := false
	for _, r := range part {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastUnderscore = false
		case r == ' ' || r == '-' || r == '_':
			if !lastUnderscore && builder.Len() > 0 {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	sanitized := builder.String()
	sanitized = strings.Trim(sanitized, "_")
	return sanitized
}

func buildDefaultFilename(parts []string, fallback, ext string) string {
	sanitizedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if sanitized := sanitizeFilenamePart(part); sanitized != "" {
			sanitizedParts = append(sanitizedParts, sanitized)
		}
	}
	if len(sanitizedParts) == 0 {
		fallbackSanitized := sanitizeFilenamePart(fallback)
		if fallbackSanitized != "" {
			sanitizedParts = append(sanitizedParts, fallbackSanitized)
		} else {
			sanitizedParts = append(sanitizedParts, "export")
		}
	}

	base := strings.Join(sanitizedParts, "_")
	return fmt.Sprintf("%s_%s%s", base, time.Now().Format(exportTimestampFormat), ext)
}

func csvFromRecords(records [][]string) string {
	if len(records) == 0 {
		return ""
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.WriteAll(records); err != nil {
		fyne.LogError("failed to create CSV", err)
		return ""
	}
	return builder.String()
}

func buildProductLookupTab(win fyne.Window, service *scraper.Service, countries []string, output binding.String) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B09XYZ1234")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	resultView := widget.NewMultiLineEntry()
	resultView.Wrapping = fyne.TextWrapWord
	resultView.Bind(output)
	resultView.SetMinRowsVisible(12)
	resultView.Disable()

	saveButton := widget.NewButton("Save...", func() {
		defaultName := buildDefaultFilename([]string{asinEntry.Text, "product"}, "product_lookup", ".txt")
		saveBindingToFile(win, output, nil, defaultName)
	})

	fetchButton := widget.NewButton("Fetch Product", func() {
		asin := strings.TrimSpace(asinEntry.Text)
		country := countrySelect.Selected
		go func() {
			output.Set(fmt.Sprintf("Fetching product %s from %s...", asin, country))
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

	form := container.New(layout.NewFormLayout(),
		widget.NewLabelWithStyle("ASIN", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), asinEntry,
		widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
	)

	content := container.NewBorder(nil, fetchButton, nil, nil,
		container.NewVBox(
			form,
			widget.NewSeparator(),
			resultView,
			container.NewHBox(layout.NewSpacer(), saveButton),
		),
	)

	return container.NewPadded(content)
}

func buildKeywordResearchTab(win fyne.Window, service *scraper.Service, countries []string, keywordOutput, categoryOutput, bestsellerOutput binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("Enter a seed keyword e.g. self publishing")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	keywordView := widget.NewMultiLineEntry()
	keywordView.Wrapping = fyne.TextWrapWord
	keywordView.Bind(keywordOutput)
	keywordView.Disable()

	keywordCSV := binding.NewString()

	categoryView := widget.NewMultiLineEntry()
	categoryView.Wrapping = fyne.TextWrapWord
	categoryView.Bind(categoryOutput)
	categoryView.Disable()

	categoryCSV := binding.NewString()

	bestsellerView := widget.NewMultiLineEntry()
	bestsellerView.Wrapping = fyne.TextWrapWord
	bestsellerView.Bind(bestsellerOutput)
	bestsellerView.Disable()

	bestsellerCSV := binding.NewString()

	keywordSave := widget.NewButton("Save...", func() {
		base := strings.TrimSpace(keywordEntry.Text)
		defaultName := buildDefaultFilename([]string{base, "keywords"}, "keyword_suggestions", ".txt")
		saveBindingToFile(win, keywordOutput, keywordCSV, defaultName)
	})

	categorySave := widget.NewButton("Save...", func() {
		base := strings.TrimSpace(keywordEntry.Text)
		defaultName := buildDefaultFilename([]string{base, "categories"}, "category_opportunities", ".txt")
		saveBindingToFile(win, categoryOutput, categoryCSV, defaultName)
	})

	bestsellerSave := widget.NewButton("Save...", func() {
		base := strings.TrimSpace(keywordEntry.Text)
		defaultName := buildDefaultFilename([]string{base, "bestsellers"}, "bestseller_snapshot", ".txt")
		saveBindingToFile(win, bestsellerOutput, bestsellerCSV, defaultName)
	})

	fetchButton := widget.NewButton("Run Research", func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		country := countrySelect.Selected
		go func() {
			keywordOutput.Set(fmt.Sprintf("Fetching keyword suggestions for %s...", keyword))
			keywordCSV.Set("")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			suggestions, err := service.KeywordSuggestions(ctx, keyword, country)
			if err != nil {
				keywordOutput.Set(fmt.Sprintf("Error: %v", err))
				keywordCSV.Set("")
			} else {
				text, csv := formatKeywordInsights(keyword, suggestions)
				keywordOutput.Set(text)
				keywordCSV.Set(csv)
			}

			categoryCSV.Set("")
			categories, err := service.CategorySuggestions(ctx, keyword, country)
			if err != nil {
				categoryOutput.Set(fmt.Sprintf("Error: %v", err))
				categoryCSV.Set("")
			} else {
				text, csv := formatCategoryTrends(categories)
				categoryOutput.Set(text)
				categoryCSV.Set(csv)
			}

			bestsellerCSV.Set("")
			bestsellers, err := service.BestsellerAnalysis(ctx, keyword, country)
			if err != nil {
				bestsellerOutput.Set(fmt.Sprintf("Error: %v", err))
				bestsellerCSV.Set("")
			} else {
				text, csv := formatBestsellerProducts(bestsellers)
				bestsellerOutput.Set(text)
				bestsellerCSV.Set(csv)
			}
		}()
	})

	grid := container.NewGridWithRows(3,
		container.NewBorder(
			widget.NewLabelWithStyle("Keyword Suggestions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(layout.NewSpacer(), keywordSave), nil, nil, keywordView,
		),
		container.NewBorder(
			widget.NewLabelWithStyle("Category Opportunities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(layout.NewSpacer(), categorySave), nil, nil, categoryView,
		),
		container.NewBorder(
			widget.NewLabelWithStyle("Bestseller Snapshot", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(layout.NewSpacer(), bestsellerSave), nil, nil, bestsellerView,
		),
	)

	form := container.New(layout.NewFormLayout(),
		widget.NewLabelWithStyle("Keyword", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), keywordEntry,
		widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
	)

	return container.NewBorder(container.NewVBox(form, widget.NewSeparator()), fetchButton, nil, nil, grid)
}

func buildCompetitiveTab(win fyne.Window, service *scraper.Service, countries []string, reverseOutput, campaignOutput binding.String) fyne.CanvasObject {
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

	reverseCSV := binding.NewString()

	campaignView := widget.NewMultiLineEntry()
	campaignView.Wrapping = fyne.TextWrapWord
	campaignView.Bind(campaignOutput)
	campaignView.Disable()

	campaignCSV := binding.NewString()

	reverseSave := widget.NewButton("Save...", func() {
		defaultName := buildDefaultFilename([]string{strings.TrimSpace(reverseAsinEntry.Text), "reverse"}, "reverse_asin", ".txt")
		saveBindingToFile(win, reverseOutput, reverseCSV, defaultName)
	})

	campaignSave := widget.NewButton("Save...", func() {
		defaultName := buildDefaultFilename([]string{strings.TrimSpace(titleEntry.Text), "campaign"}, "campaign_keywords", ".txt")
		saveBindingToFile(win, campaignOutput, campaignCSV, defaultName)
	})

	reverseButton := widget.NewButton("Reverse ASIN Search", func() {
		asin := strings.TrimSpace(reverseAsinEntry.Text)
		country := countrySelect.Selected
		go func() {
			reverseOutput.Set(fmt.Sprintf("Running reverse ASIN search for %s...", asin))
			reverseCSV.Set("")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			insights, err := service.ReverseASINSearch(ctx, asin, country)
			if err != nil {
				reverseOutput.Set(fmt.Sprintf("Error: %v", err))
				reverseCSV.Set("")
				return
			}
			text, csv := formatKeywordInsights(fmt.Sprintf("ASIN %s", asin), insights)
			reverseOutput.Set(text)
			reverseCSV.Set(csv)
		}()
	})

	campaignButton := widget.NewButton("Generate AMS Keywords", func() {
		country := countrySelect.Selected
		competitors := strings.Split(strings.ReplaceAll(competitorKeywordsEntry.Text, "\r", ""), "\n")
		go func() {
			campaignOutput.Set("Generating keyword list...")
			campaignCSV.Set("")
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			keywords, err := service.GenerateAMSKeywords(ctx, titleEntry.Text, descriptionEntry.Text, competitors, country)
			if err != nil {
				campaignOutput.Set(fmt.Sprintf("Error: %v", err))
				campaignCSV.Set("")
				return
			}
			flagged := scraper.FlagIllegalKeywords(keywords)
			text, csv := formatCampaignKeywords(keywords, flagged)
			campaignOutput.Set(text)
			campaignCSV.Set(csv)
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
		container.NewBorder(
			widget.NewLabelWithStyle("Reverse ASIN Insights", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(layout.NewSpacer(), reverseSave), nil, nil, reverseView,
		),
		container.NewBorder(
			widget.NewLabelWithStyle("Keyword Portfolio", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(layout.NewSpacer(), campaignSave), nil, nil, campaignView,
		),
	)
	results.SetOffset(0.5)

	return container.NewBorder(form, buttonRow, nil, nil, results)
}

func buildInternationalTab(win fyne.Window, service *scraper.Service, countries []string, output binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("Seed keyword e.g. coloring book")

	countrySelect := widget.NewCheckGroup(countries, nil)
	countrySelect.Horizontal = true

	outputView := widget.NewMultiLineEntry()
	outputView.Wrapping = fyne.TextWrapWord
	outputView.Bind(output)
	outputView.Disable()

	csvBinding := binding.NewString()

	saveButton := widget.NewButton("Save...", func() {
		defaultName := buildDefaultFilename([]string{strings.TrimSpace(keywordEntry.Text), "international"}, "international_keywords", ".txt")
		saveBindingToFile(win, output, csvBinding, defaultName)
	})

	fetchButton := widget.NewButton("Collect International Keywords", func() {
		selected := countrySelect.Selected
		if len(selected) == 0 {
			selected = countries
		}
		keyword := keywordEntry.Text
		go func() {
			output.Set("Collecting international keyword data...")
			csvBinding.Set("")
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			keywords, err := service.InternationalKeywords(ctx, keyword, selected)
			if err != nil {
				output.Set(fmt.Sprintf("Error: %v", err))
				csvBinding.Set("")
				return
			}
			text, csv := formatInternationalKeywords(keywords)
			output.Set(text)
			csvBinding.Set(csv)
		}()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Keyword", keywordEntry),
		),
		widget.NewLabel("Select markets (leave blank for all)"),
		countrySelect,
	)

	return container.NewBorder(form, container.NewVBox(container.NewHBox(layout.NewSpacer(), saveButton), fetchButton), nil, nil, outputView)
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

	var builder strings.Builder
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
	var builder strings.Builder
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
	var builder strings.Builder
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
	var builder strings.Builder
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
		key := strings.TrimSpace(keyword)
		if _, exists := flaggedSet[key]; exists {
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
	var builder strings.Builder
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
