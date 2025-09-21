package ui

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
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/umar/amazon-product-scraper/internal/scraper"
)

func Run() {
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

func buildProductLookupTab(win fyne.Window, service *scraper.Service, countries []string, output binding.String) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B09XYZ1234")

	countrySelect := widget.NewSelect(countries, nil)
	countrySelect.SetSelected("US")

	resultView := widget.NewMultiLineEntry()
	resultView.Wrapping = fyne.TextWrapWord
	resultView.Bind(output)
	resultView.SetMinRowsVisible(12)
	makeEntryReadOnly(resultView)

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

	results := newResultPanel("Product Details", resultView, output, nil, win, "")

	content := container.NewBorder(form, container.NewHBox(layout.NewSpacer(), fetchButton), nil, nil,
		container.NewVBox(widget.NewSeparator(), results),
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
	makeEntryReadOnly(keywordView)

	categoryView := widget.NewMultiLineEntry()
	categoryView.Wrapping = fyne.TextWrapWord
	categoryView.Bind(categoryOutput)
	makeEntryReadOnly(categoryView)

	bestsellerView := widget.NewMultiLineEntry()
	keywordCSV := binding.NewString()
	categoryCSV := binding.NewString()
	bestsellerCSV := binding.NewString()
	bestsellerView.Wrapping = fyne.TextWrapWord
	bestsellerView.Bind(bestsellerOutput)
	makeEntryReadOnly(bestsellerView)

	keywordCSV.Set("")
	categoryCSV.Set("")
	bestsellerCSV.Set("")

	metricControls, metricFilterPanel := newMetricFilterControls()
	bestsellerControls, bestsellerFilterPanel := newBestsellerFilterControls()

	fetchButton := widget.NewButton("Run Research", func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		country := countrySelect.Selected
		filters := metricControls.keywordFilter()
		showDensity := metricControls.showDensity()
		bestsellerFilter := bestsellerControls.filter()
		showBSR := bestsellerControls.showBSR()
		go func() {
			keywordOutput.Set(fmt.Sprintf("Fetching keyword suggestions for %s...", keyword))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			suggestions, err := service.KeywordSuggestions(ctx, keyword, country, filters)
			if err != nil {
				keywordOutput.Set(renderScrapeError(err))
				keywordCSV.Set("")
			} else {
				formatted, csvData := formatKeywordInsights(keyword, suggestions, showDensity)
				keywordOutput.Set(formatted)
				keywordCSV.Set(csvData)
			}

			categories, err := service.CategorySuggestions(ctx, keyword, country)
			if err != nil {
				categoryOutput.Set(renderScrapeError(err))
				categoryCSV.Set("")
			} else {
				formatted, csvData := formatCategoryTrends(categories)
				categoryOutput.Set(formatted)
				categoryCSV.Set(csvData)
			}

			bestsellers, err := service.BestsellerAnalysis(ctx, keyword, country, bestsellerFilter)
			if err != nil {
				bestsellerOutput.Set(renderScrapeError(err))
				bestsellerCSV.Set("")
			} else {
				formatted, csvData := formatBestsellerProducts(bestsellers, showBSR)
				bestsellerOutput.Set(formatted)
				bestsellerCSV.Set(csvData)
			}
		}()
	})

	keywordResults := newResultPanel("Keyword Suggestions", keywordView, keywordOutput, keywordCSV, win, "keyword-suggestions.csv")
	categoryResults := newResultPanel("Category Opportunities", categoryView, categoryOutput, categoryCSV, win, "category-opportunities.csv")
	bestsellerResults := newResultPanel("Bestseller Snapshot", bestsellerView, bestsellerOutput, bestsellerCSV, win, "bestseller-snapshot.csv")

	grid := container.NewGridWithRows(3,
		keywordResults,
		categoryResults,
		bestsellerResults,
	)

	form := container.New(layout.NewFormLayout(),
		widget.NewLabelWithStyle("Keyword", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), keywordEntry,
		widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
	)

	controls := container.NewVBox(
		form,
		widget.NewSeparator(),
		metricFilterPanel,
		widget.NewSeparator(),
		bestsellerFilterPanel,
	)

	return container.NewBorder(controls, fetchButton, nil, nil, grid)
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
	makeEntryReadOnly(reverseView)

	campaignView := widget.NewMultiLineEntry()
	reverseCSV := binding.NewString()
	campaignCSV := binding.NewString()
	reverseCSV.Set("")
	campaignCSV.Set("")
	campaignView.Wrapping = fyne.TextWrapWord
	campaignView.Bind(campaignOutput)
	makeEntryReadOnly(campaignView)

	metricControls, metricPanel := newMetricFilterControls()

	reverseButton := widget.NewButton("Reverse ASIN Search", func() {
		asin := strings.TrimSpace(reverseAsinEntry.Text)
		country := countrySelect.Selected
		filters := metricControls.keywordFilter()
		showDensity := metricControls.showDensity()
		go func() {
			reverseOutput.Set(fmt.Sprintf("Running reverse ASIN search for %s...", asin))
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			insights, err := service.ReverseASINSearch(ctx, asin, country, filters)
			if err != nil {
				reverseOutput.Set(renderScrapeError(err))
				reverseCSV.Set("")
				return
			}
			formatted, csvData := formatKeywordInsights(fmt.Sprintf("ASIN %s", asin), insights, showDensity)
			reverseOutput.Set(formatted)
			reverseCSV.Set(csvData)
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
				campaignCSV.Set("")
				return
			}
			flagged := scraper.FlagIllegalKeywords(keywords)
			formatted, csvData := formatCampaignKeywords(keywords, flagged)
			campaignOutput.Set(formatted)
			campaignCSV.Set(csvData)
		}()
	})

	buttonRow := container.NewHBox(reverseButton, campaignButton)

	form := container.NewVBox(
		container.New(layout.NewFormLayout(),
			widget.NewLabelWithStyle("Country", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), countrySelect,
			widget.NewLabelWithStyle("Reverse ASIN", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), reverseAsinEntry,
		),
		widget.NewSeparator(),
		metricPanel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Amazon Ads Planner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Description", descriptionEntry),
			widget.NewFormItem("Competitor Keywords", competitorKeywordsEntry),
		),
	)

	reverseResults := newResultPanel("Reverse ASIN Insights", reverseView, reverseOutput, reverseCSV, win, "reverse-asin.csv")
	campaignResults := newResultPanel("Keyword Portfolio", campaignView, campaignOutput, campaignCSV, win, "ams-keywords.csv")

	results := container.NewHSplit(
		reverseResults,
		campaignResults,
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
	makeEntryReadOnly(outputView)

	csvBinding := binding.NewString()
	csvBinding.Set("")

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
				csvBinding.Set("")
				return
			}
			formatted, csvData := formatInternationalKeywords(keywords)
			output.Set(formatted)
			csvBinding.Set(csvData)
		}()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Keyword", keywordEntry),
		),
		widget.NewLabel("Select markets (leave blank for all)"),
		countrySelect,
	)

	results := newResultPanel("International Keyword Radar", outputView, output, csvBinding, win, "international-keywords.csv")

	return container.NewBorder(form, container.NewHBox(layout.NewSpacer(), fetchButton), nil, nil, container.NewVBox(widget.NewSeparator(), results))
}

func newResultPanel(title string, view *widget.Entry, textData binding.String, csvData binding.String, win fyne.Window, fileName string) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	actions := []fyne.CanvasObject{layout.NewSpacer(), newCopyButton(win, textData)}
	if csvData != nil {
		actions = append(actions, newExportCSVButton(win, csvData, fileName))
	}

	return container.NewBorder(label, container.NewHBox(actions...), nil, nil, view)
}

// makeEntryReadOnly prevents the user from editing result fields while still
// allowing programmatic updates through data bindings. Older versions of Fyne
// do not expose a SetReadOnly API, so we disable the entry to achieve the same
// effect.
func makeEntryReadOnly(entry *widget.Entry) {
	entry.Disable()
}

func newCopyButton(win fyne.Window, data binding.String) *widget.Button {
	return widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if win == nil || data == nil {
			return
		}
		value, err := data.Get()
		if err != nil || strings.TrimSpace(value) == "" {
			dialog.ShowInformation("Nothing to copy", "Run a search to generate results before copying.", win)
			return
		}
		win.Clipboard().SetContent(value)
	})
}

func newExportCSVButton(win fyne.Window, data binding.String, fileName string) *widget.Button {
	return widget.NewButtonWithIcon("Export CSV", theme.DocumentSaveIcon(), func() {
		if win == nil || data == nil {
			return
		}
		csvValue, err := data.Get()
		if err != nil || strings.TrimSpace(csvValue) == "" {
			dialog.ShowInformation("Nothing to export", "There is no CSV data available yet. Run a search first.", win)
			return
		}
		name := strings.TrimSpace(fileName)
		if name == "" {
			name = "results.csv"
		}
		save := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			if _, err := writer.Write([]byte(csvValue)); err != nil {
				dialog.ShowError(err, win)
			}
		}, win)
		save.SetFileName(name)
		save.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		save.Show()
	})
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
	if product == nil {
		return "No product details available"
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "Title: %s\n", product.Title)
	fmt.Fprintf(builder, "ASIN: %s\n", product.ASIN)
	fmt.Fprintf(builder, "Price: %s (%s)\n", product.Price, product.Currency)
	fmt.Fprintf(builder, "Rating: %s\n", product.Rating)
	fmt.Fprintf(builder, "Reviews: %s\n", product.ReviewCount)
	fmt.Fprintf(builder, "Availability: %s\n", product.Availability)
	fmt.Fprintf(builder, "Brand: %s\n", product.Brand)
	fmt.Fprintf(builder, "Delivery: %s\n", product.DeliveryMessage)
	if product.Publisher != "" {
		fmt.Fprintf(builder, "Publisher: %s\n", product.Publisher)
	} else {
		builder.WriteString("Publisher: Not available\n")
	}
	fmt.Fprintf(builder, "Independent Publisher: %t\n", product.IsIndependent)
	if strings.TrimSpace(product.PrintLength) != "" {
		fmt.Fprintf(builder, "Print Length: %s\n", product.PrintLength)
	} else {
		builder.WriteString("Print Length: Not available\n")
	}
	if strings.TrimSpace(product.Dimensions) != "" {
		fmt.Fprintf(builder, "Dimensions: %s\n", product.Dimensions)
	} else {
		builder.WriteString("Dimensions: Not available\n")
	}
	if strings.TrimSpace(product.PublicationDate) != "" {
		fmt.Fprintf(builder, "Publication Date: %s\n", product.PublicationDate)
	} else {
		builder.WriteString("Publication Date: Not available\n")
	}
	if strings.TrimSpace(product.Language) != "" {
		fmt.Fprintf(builder, "Language: %s\n", product.Language)
	} else {
		builder.WriteString("Language: Not available\n")
	}
	if strings.TrimSpace(product.ISBN10) != "" {
		fmt.Fprintf(builder, "ISBN-10: %s\n", product.ISBN10)
	} else {
		builder.WriteString("ISBN-10: Not available\n")
	}
	if strings.TrimSpace(product.ISBN13) != "" {
		fmt.Fprintf(builder, "ISBN-13: %s\n", product.ISBN13)
	} else {
		builder.WriteString("ISBN-13: Not available\n")
	}
	if len(product.BestSellerRanks) > 0 {
		builder.WriteString("Best Seller Ranks:\n")
		for _, rank := range product.BestSellerRanks {
			fmt.Fprintf(builder, "- #%d in %s\n", rank.Rank, rank.Category)
		}
	} else {
		builder.WriteString("Best Seller Ranks: Not available\n")
	}
	if product.TitleDensity >= 0 {
		fmt.Fprintf(builder, "Title Density: %.2f\n", product.TitleDensity)
	} else {
		builder.WriteString("Title Density: N/A\n")
	}
	fmt.Fprintf(builder, "URL: %s\n", product.URL)
	fmt.Fprintf(builder, "Fetched: %s", product.FetchedAt.Format(time.RFC1123))

	return builder.String()
}

func formatKeywordInsights(title string, insights []scraper.KeywordInsight, showDensity bool) (string, string) {
	if len(insights) == 0 {
		return fmt.Sprintf("No keyword suggestions available for %s — try relaxing your filters.", title), ""
	}

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Keyword Research for %s\n", title))
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")
	builder.WriteString("Legend: 🟢 Strong | 🟡 Moderate | 🔴 Weak\n\n")

	header := "Keyword | Search Volume | Competition | Relevancy"
	if showDensity {
		header += " | Title Density"
	}
	builder.WriteString(header + "\n")
	builder.WriteString(strings.Repeat("-", len(header)) + "\n")

	records := [][]string{{"Keyword", "Search Volume", "Competition", "Relevancy"}}
	if showDensity {
		records[0] = append(records[0], "Title Density")
	}

	for _, insight := range insights {
		volumeSignal := searchVolumeSignal(insight.SearchVolume)
		competitionSignal := competitionSignal(insight.CompetitionScore)
		volumeDisplay := fmt.Sprintf("%d %s", insight.SearchVolume, volumeSignal.String())
		competitionDisplay := fmt.Sprintf("%.2f %s", insight.CompetitionScore, competitionSignal.String())
		row := []string{
			insight.Keyword,
			volumeDisplay,
			competitionDisplay,
			fmt.Sprintf("%.2f", insight.RelevancyScore),
		}
		builder.WriteString(strings.Join([]string{insight.Keyword, volumeDisplay, competitionDisplay, fmt.Sprintf("%.2f", insight.RelevancyScore)}, " | "))
		if showDensity {
			density := densityString(insight.TitleDensity)
			builder.WriteString(fmt.Sprintf(" | %s", density))
			row = append(row, density)
		}
		builder.WriteString("\n")
		records = append(records, row)
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

func formatBestsellerProducts(products []scraper.BestsellerProduct, showBSR bool) (string, string) {
	builder := strings.Builder{}
	builder.WriteString("Bestseller Snapshot\n")
	builder.WriteString(strings.Repeat("=", 40))
	builder.WriteString("\n")

	headers := []string{"Rank", "Title", "ASIN", "Price", "Rating", "Reviews"}
	if showBSR {
		headers = append(headers, "Best Seller Rank", "Category")
	}
	headers = append(headers, "Publisher", "Independent", "Title Density", "URL")

	records := [][]string{headers}

	for _, product := range products {
		builder.WriteString(fmt.Sprintf("#%d %s\n", product.Rank, product.Title))
		builder.WriteString(fmt.Sprintf("ASIN: %s | Price: %s | Rating: %s | Reviews: %s\n", product.ASIN, product.Price, product.Rating, product.ReviewCount))
		if showBSR {
			if product.BestSeller > 0 {
				builder.WriteString(fmt.Sprintf("Best Seller Rank: #%d (%s)\n", product.BestSeller, product.Category))
			} else {
				builder.WriteString("Best Seller Rank: Not available\n")
			}
		}
		publisher := product.Publisher
		if strings.TrimSpace(publisher) == "" {
			publisher = "Unknown"
		}
		builder.WriteString(fmt.Sprintf("Publisher: %s | Independent: %t\n", publisher, product.IsIndie))
		builder.WriteString(fmt.Sprintf("Title Density: %s\n", densityString(product.TitleDensity)))
		builder.WriteString(fmt.Sprintf("URL: %s\n\n", product.URL))

		row := []string{
			strconv.Itoa(product.Rank),
			product.Title,
			product.ASIN,
			product.Price,
			product.Rating,
			product.ReviewCount,
		}
		if showBSR {
			bsrValue := ""
			if product.BestSeller > 0 {
				bsrValue = strconv.Itoa(product.BestSeller)
			}
			row = append(row, bsrValue, product.Category)
		}
		row = append(row, publisher, fmt.Sprintf("%t", product.IsIndie), densityString(product.TitleDensity), product.URL)
		records = append(records, row)
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

type metricFilterControls struct {
	searchCheck      *widget.Check
	searchEntry      *widget.Entry
	competitionCheck *widget.Check
	competitionEntry *widget.Entry
	densityCheck     *widget.Check
	densityEntry     *widget.Entry
}

func newMetricFilterControls() (*metricFilterControls, fyne.CanvasObject) {
	controls := &metricFilterControls{}

	controls.searchEntry = widget.NewEntry()
	controls.searchEntry.SetText("500")
	controls.searchEntry.SetPlaceHolder("500")
	controls.searchEntry.Disable()
	controls.searchCheck = widget.NewCheck("Search volume ≥", func(checked bool) {
		if checked {
			controls.searchEntry.Enable()
		} else {
			controls.searchEntry.Disable()
		}
	})

	controls.competitionEntry = widget.NewEntry()
	controls.competitionEntry.SetText("0.60")
	controls.competitionEntry.SetPlaceHolder("0.60")
	controls.competitionEntry.Disable()
	controls.competitionCheck = widget.NewCheck("Competition ≤", func(checked bool) {
		if checked {
			controls.competitionEntry.Enable()
		} else {
			controls.competitionEntry.Disable()
		}
	})

	controls.densityEntry = widget.NewEntry()
	controls.densityEntry.SetText("0.40")
	controls.densityEntry.SetPlaceHolder("0.40")
	controls.densityEntry.Disable()
	controls.densityCheck = widget.NewCheck("Title density ≤", func(checked bool) {
		if checked {
			controls.densityEntry.Enable()
		} else {
			controls.densityEntry.Disable()
		}
	})

	panel := container.NewVBox(
		widget.NewLabelWithStyle("Metric Filters", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(controls.searchCheck, controls.searchEntry),
		container.NewHBox(controls.competitionCheck, controls.competitionEntry),
		container.NewHBox(controls.densityCheck, controls.densityEntry),
	)

	return controls, panel
}

func (c *metricFilterControls) keywordFilter() scraper.KeywordFilter {
	filter := scraper.KeywordFilter{}
	if c.searchCheck.Checked {
		filter.MinSearchVolume = parsePositiveInt(c.searchEntry.Text, 500)
	}
	if c.competitionCheck.Checked {
		filter.MaxCompetitionScore = parsePositiveFloat(c.competitionEntry.Text, 0.6)
	}
	if c.densityCheck.Checked {
		filter.MaxTitleDensity = parsePositiveFloat(c.densityEntry.Text, 0.4)
	}
	return filter
}

func (c *metricFilterControls) showDensity() bool {
	return c != nil && c.densityCheck.Checked
}

type bestsellerFilterControls struct {
	indieCheck      *widget.Check
	bsrLimitCheck   *widget.Check
	bsrEntry        *widget.Entry
	bsrDisplayCheck *widget.Check
}

func newBestsellerFilterControls() (*bestsellerFilterControls, fyne.CanvasObject) {
	controls := &bestsellerFilterControls{}

	controls.indieCheck = widget.NewCheck("Independent publishers only", nil)

	controls.bsrEntry = widget.NewEntry()
	controls.bsrEntry.SetText("50000")
	controls.bsrEntry.SetPlaceHolder("50000")
	controls.bsrEntry.Disable()
	controls.bsrLimitCheck = widget.NewCheck("Limit BSR to ≤", func(checked bool) {
		if checked {
			controls.bsrEntry.Enable()
		} else {
			controls.bsrEntry.Disable()
		}
	})

	controls.bsrDisplayCheck = widget.NewCheck("Show BSR details", nil)
	controls.bsrDisplayCheck.Checked = true
	controls.bsrDisplayCheck.Refresh()

	panel := container.NewVBox(
		widget.NewLabelWithStyle("Bestseller Filters", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		controls.indieCheck,
		container.NewHBox(controls.bsrLimitCheck, controls.bsrEntry),
		controls.bsrDisplayCheck,
	)

	return controls, panel
}

func (c *bestsellerFilterControls) filter() scraper.BestsellerFilter {
	filter := scraper.BestsellerFilter{}
	if c.indieCheck.Checked {
		filter.IndependentOnly = true
	}
	if c.bsrLimitCheck.Checked {
		filter.MaxBestSellerRank = parsePositiveInt(c.bsrEntry.Text, 50000)
	}
	return filter
}

func (c *bestsellerFilterControls) showBSR() bool {
	if c == nil {
		return false
	}
	return c.bsrDisplayCheck.Checked || c.bsrLimitCheck.Checked
}

func parsePositiveInt(value string, fallback int) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func parsePositiveFloat(value string, fallback float64) float64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}

func densityString(value float64) string {
	if value < 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.2f", value)
}

type metricSignal struct {
	emoji string
	label string
}

func (m metricSignal) String() string {
	emoji := strings.TrimSpace(m.emoji)
	label := strings.TrimSpace(m.label)
	switch {
	case emoji != "" && label != "":
		return fmt.Sprintf("%s %s", emoji, label)
	case label != "":
		return label
	default:
		return emoji
	}
}

func searchVolumeSignal(volume int) metricSignal {
	if volume <= 0 {
		return metricSignal{label: "N/A"}
	}
	switch {
	case volume >= 1500:
		return metricSignal{emoji: "🟢", label: "High"}
	case volume >= 600:
		return metricSignal{emoji: "🟡", label: "Medium"}
	default:
		return metricSignal{emoji: "🔴", label: "Low"}
	}
}

func competitionSignal(score float64) metricSignal {
	if score < 0 {
		return metricSignal{label: "N/A"}
	}
	switch {
	case score <= 0.4:
		return metricSignal{emoji: "🟢", label: "Low"}
	case score <= 0.7:
		return metricSignal{emoji: "🟡", label: "Moderate"}
	default:
		return metricSignal{emoji: "🔴", label: "High"}
	}
}
