package ui

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/url"
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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/umarmf343/Umar-kdp-product-api/internal/scraper"
)

const (
	requestTimeout = 30 * time.Second
	tutorialURL    = "https://www.youtube.com/results?search_query=RankBeam+tutorial"
)

var activeService *scraper.Service

// Run initialises and displays the desktop application.
func Run() {
	application := app.NewWithID("rankbeam")
	application.Settings().SetTheme(theme.LightTheme())

	if lifecycle := application.Lifecycle(); lifecycle != nil {
		lifecycle.SetOnStopped(func() {
			if activeService != nil {
				activeService.Close()
			}
		})
	}

	window := application.NewWindow("RankBeam")
	window.Resize(fyne.NewSize(1024, 720))
	window.SetMaster()

	client, licenseKey, licenseError := enforceLicense()
	if licenseError != "" {
		renderLicenseFailure(window, client, licenseError)
		window.ShowAndRun()
		return
	}

	loadMainApplication(window, licenseKey)
	window.ShowAndRun()
}

func loadMainApplication(window fyne.Window, licenseKey string) {
	if window == nil {
		return
	}

	if activeService != nil {
		activeService.Close()
	}

	service := scraper.NewService(25*time.Second, 25)
	activeService = service

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
	tabs.SetTabLocation(container.TabLocationTop)

	tutorialButton := widget.NewButtonWithIcon("Tutorial", tutorialIcon, func() {
		parsed, err := url.Parse(tutorialURL)
		if err != nil {
			dialog.ShowError(fmt.Errorf("open tutorial: %w", err), window)
			return
		}

		if app := fyne.CurrentApp(); app != nil {
			if err := app.OpenURL(parsed); err != nil {
				dialog.ShowError(fmt.Errorf("open tutorial: %w", err), window)
			}
		}
	})
	tutorialButton.Importance = widget.HighImportance

	topBar := container.NewPadded(container.NewHBox(layout.NewSpacer(), tutorialButton))

	window.SetContent(container.NewBorder(topBar, nil, nil, nil, tabs))

	title := "RankBeam"
	if licenseKey != "" {
		title = fmt.Sprintf("%s — License %s", title, summarizeKey(licenseKey))
	}
	window.SetTitle(title)
}

func buildProductLookupTab(window fyne.Window, service *scraper.Service, countries []string, result binding.String) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B08N5WRWNW")

	countrySelect := widget.NewSelect(countries, nil)
	if len(countries) > 0 {
		countrySelect.SetSelected(countries[0])
	}

	resultLabel := widget.NewLabelWithData(result)
	resultLabel.Wrapping = fyne.TextWrapWord

	fetch := func() {
		asin := strings.TrimSpace(asinEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if asin == "" {
			dialog.ShowInformation("Product Lookup", "Enter a valid ASIN to continue.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Product Lookup", "Select the Amazon marketplace you wish to query.", window)
			return
		}

		progress := dialog.NewProgressInfinite("Fetching Product", fmt.Sprintf("Looking up %s on %s…", strings.ToUpper(asin), strings.ToUpper(country)), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			details, err := service.FetchProduct(ctx, asin, country)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(result, fmt.Sprintf("Unable to fetch product: %v", err))
					return
				}
				safeSet(result, formatProductDetails(details))
			})
		}()
	}

	form := widget.NewForm(
		widget.NewFormItem("ASIN", asinEntry),
		widget.NewFormItem("Marketplace", countrySelect),
	)
	form.SubmitText = "Fetch Product"
	form.OnSubmit = fetch

	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		container.NewVScroll(container.NewPadded(resultLabel)),
	)

	return content
}

func buildKeywordResearchTab(window fyne.Window, service *scraper.Service, countries []string, keywordResult, categoryResult, bestsellerResult binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("children's book about space")

	countrySelect := widget.NewSelect(countries, nil)
	if len(countries) > 0 {
		countrySelect.SetSelected(countries[0])
	}

	minVolumeEntry := widget.NewEntry()
	minVolumeEntry.SetPlaceHolder("0")
	maxCompetitionEntry := widget.NewEntry()
	maxCompetitionEntry.SetPlaceHolder("1.0")
	maxDensityEntry := widget.NewEntry()
	maxDensityEntry.SetPlaceHolder("100")

	maxRankEntry := widget.NewEntry()
	maxRankEntry.SetPlaceHolder("50000")
	indieOnlyCheck := widget.NewCheck("Indie authors only", nil)

	keywordLabel := widget.NewLabelWithData(keywordResult)
	keywordLabel.Wrapping = fyne.TextWrapWord
	categoryLabel := widget.NewLabelWithData(categoryResult)
	categoryLabel.Wrapping = fyne.TextWrapWord
	bestsellerLabel := widget.NewLabelWithData(bestsellerResult)
	bestsellerLabel.Wrapping = fyne.TextWrapWord

	fetchKeywords := func() {
		seed := strings.TrimSpace(keywordEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if seed == "" {
			dialog.ShowInformation("Keyword Research", "Enter a seed keyword to continue.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Keyword Research", "Select a marketplace before running research.", window)
			return
		}

		filters, err := parseKeywordFilter(minVolumeEntry.Text, maxCompetitionEntry.Text, maxDensityEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		progress := dialog.NewProgressInfinite("Keyword Research", fmt.Sprintf("Collecting ideas for \"%s\"…", seed), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			insights, err := service.KeywordSuggestions(ctx, seed, country, filters)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(keywordResult, fmt.Sprintf("Unable to fetch keyword suggestions: %v", err))
					return
				}
				safeSet(keywordResult, formatKeywordInsights(insights))
			})
		}()
	}

	fetchCategories := func() {
		seed := strings.TrimSpace(keywordEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if seed == "" {
			dialog.ShowInformation("Category Insights", "Enter a seed keyword before analysing categories.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Category Insights", "Select a marketplace before analysing categories.", window)
			return
		}

		progress := dialog.NewProgressInfinite("Category Insights", "Discovering high performing categories…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			trends, err := service.FetchCategoryTrends(ctx, seed, country)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(categoryResult, fmt.Sprintf("Unable to fetch category insights: %v", err))
					return
				}
				safeSet(categoryResult, formatCategoryTrends(trends))
			})
		}()
	}

	fetchBestsellers := func() {
		seed := strings.TrimSpace(keywordEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if seed == "" {
			dialog.ShowInformation("Bestseller Snapshot", "Enter a keyword to analyse bestseller listings.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Bestseller Snapshot", "Select a marketplace before analysing bestsellers.", window)
			return
		}

		filter, err := parseBestsellerFilter(maxRankEntry.Text, indieOnlyCheck.Checked)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		progress := dialog.NewProgressInfinite("Bestseller Snapshot", fmt.Sprintf("Reviewing top results for \"%s\"…", seed), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			products, err := service.BestsellerAnalysis(ctx, seed, country, filter)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(bestsellerResult, fmt.Sprintf("Unable to analyse bestsellers: %v", err))
					return
				}
				safeSet(bestsellerResult, formatBestsellerProducts(products))
			})
		}()
	}

	keywordActions := container.NewHBox(
		widget.NewButton("Fetch Keyword Suggestions", fetchKeywords),
		widget.NewButton("Category Insights", fetchCategories),
		widget.NewButton("Bestseller Snapshot", fetchBestsellers),
	)

	keywordOutputs := container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle("Keyword Suggestions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keywordLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Category Intelligence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		categoryLabel,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Bestseller Analysis", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bestsellerLabel,
	))

	filtersForm := widget.NewForm(
		widget.NewFormItem("Min Search Volume", minVolumeEntry),
		widget.NewFormItem("Max Competition", maxCompetitionEntry),
		widget.NewFormItem("Max Title Density", maxDensityEntry),
		widget.NewFormItem("Max BSR", maxRankEntry),
		widget.NewFormItem("Filters", indieOnlyCheck),
	)

	return container.NewBorder(filtersForm, nil, nil, nil, container.NewVBox(
		widget.NewLabelWithStyle("Keyword Research Toolkit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keywordEntry,
		countrySelect,
		keywordActions,
		layout.NewSpacer(),
		keywordOutputs,
	))
}

func buildCompetitiveTab(window fyne.Window, service *scraper.Service, countries []string, reverseResult, campaignResult binding.String) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B0C1234XYZ")

	countrySelect := widget.NewSelect(countries, nil)
	if len(countries) > 0 {
		countrySelect.SetSelected(countries[0])
	}

	minVolumeEntry := widget.NewEntry()
	minVolumeEntry.SetPlaceHolder("0")
	maxCompetitionEntry := widget.NewEntry()
	maxCompetitionEntry.SetPlaceHolder("1.0")
	maxDensityEntry := widget.NewEntry()
	maxDensityEntry.SetPlaceHolder("100")

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Book title or product headline")
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Paste your blurb or key selling points here…")
	competitorEntry := widget.NewMultiLineEntry()
	competitorEntry.SetPlaceHolder("Comma separated competitor keywords or ASIN phrases")

	reverseLabel := widget.NewLabelWithData(reverseResult)
	reverseLabel.Wrapping = fyne.TextWrapWord
	campaignLabel := widget.NewLabelWithData(campaignResult)
	campaignLabel.Wrapping = fyne.TextWrapWord

	reverseButton := widget.NewButton("Run Reverse ASIN", func() {
		asin := strings.TrimSpace(asinEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if asin == "" {
			dialog.ShowInformation("Reverse ASIN", "Provide an ASIN to analyse.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Reverse ASIN", "Select a marketplace before analysing.", window)
			return
		}

		filters, err := parseKeywordFilter(minVolumeEntry.Text, maxCompetitionEntry.Text, maxDensityEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		progress := dialog.NewProgressInfinite("Reverse ASIN", fmt.Sprintf("Investigating %s…", strings.ToUpper(asin)), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			insights, err := service.ReverseASINSearch(ctx, asin, country, filters)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(reverseResult, fmt.Sprintf("Unable to run reverse ASIN: %v", err))
					return
				}
				safeSet(reverseResult, formatKeywordInsights(insights))
			})
		}()
	})

	campaignButton := widget.NewButton("Generate Campaign Keywords", func() {
		country := strings.TrimSpace(countrySelect.Selected)
		if country == "" {
			dialog.ShowInformation("Campaign Builder", "Select a marketplace to generate keywords.", window)
			return
		}

		competitors, err := parseCSVKeywords(competitorEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("unable to parse competitor keywords: %w", err), window)
			return
		}

		progress := dialog.NewProgressInfinite("Campaign Builder", "Composing Amazon Ads keyword list…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			keywords, err := service.GenerateAMSKeywords(ctx, titleEntry.Text, descriptionEntry.Text, competitors, country)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(campaignResult, fmt.Sprintf("Unable to generate campaign keywords: %v", err))
					return
				}
				safeSet(campaignResult, formatCampaignKeywords(keywords))
			})
		}()
	})

	reverseSection := container.NewVBox(
		widget.NewLabelWithStyle("Reverse ASIN Intelligence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		asinEntry,
		container.NewGridWithColumns(3, minVolumeEntry, maxCompetitionEntry, maxDensityEntry),
		reverseButton,
		widget.NewSeparator(),
		container.NewVScroll(reverseLabel),
	)

	campaignSection := container.NewVBox(
		widget.NewLabelWithStyle("Amazon Ads Planner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		titleEntry,
		descriptionEntry,
		competitorEntry,
		campaignButton,
		widget.NewSeparator(),
		container.NewVScroll(campaignLabel),
	)

	sidebar := widget.NewForm(
		widget.NewFormItem("Marketplace", countrySelect),
	)

	return container.NewBorder(sidebar, nil, nil, nil, container.NewVBox(reverseSection, widget.NewSeparator(), campaignSection))
}

func buildInternationalTab(window fyne.Window, service *scraper.Service, countries []string, result binding.String) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("mindfulness journal")

	countryGroup := widget.NewCheckGroup(countries, nil)
	defaults := defaultInternationalSelection(countries)
	if len(defaults) > 0 {
		countryGroup.SetSelected(defaults)
	}

	resultLabel := widget.NewLabelWithData(result)
	resultLabel.Wrapping = fyne.TextWrapWord

	fetch := func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		if keyword == "" {
			dialog.ShowInformation("International Research", "Enter a seed keyword to continue.", window)
			return
		}

		selected := countryGroup.Selected
		if len(selected) == 0 {
			dialog.ShowInformation("International Research", "Select at least one marketplace to analyse.", window)
			return
		}

		progress := dialog.NewProgressInfinite("International Research", "Localising your keyword list…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			keywords, err := service.InternationalKeywords(ctx, keyword, selected)

			queueOnMain(window, func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(result, fmt.Sprintf("Unable to fetch international keywords: %v", err))
					return
				}
				safeSet(result, formatInternationalKeywords(keywords))
			})
		}()
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("International Keyword Expansion", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keywordEntry,
		countryGroup,
		widget.NewButton("Generate Suggestions", fetch),
		widget.NewSeparator(),
		container.NewVScroll(resultLabel),
	)
}

func formatProductDetails(details *scraper.ProductDetails) string {
	if details == nil {
		return "No product details were returned."
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%s (%s)\n", fallback(details.Title, "Unknown Title"), strings.ToUpper(fallback(details.ASIN, "N/A")))
	if details.Price != "" || details.Currency != "" {
		fmt.Fprintf(builder, "Price: %s %s\n", fallback(details.Currency, "$"), fallback(details.Price, "0.00"))
	}
	if details.Rating != "" || details.ReviewCount != "" {
		fmt.Fprintf(builder, "Rating: %s (%s reviews)\n", fallback(details.Rating, "N/A"), fallback(details.ReviewCount, "0"))
	}
	if details.Availability != "" {
		fmt.Fprintf(builder, "Availability: %s\n", details.Availability)
	}
	if details.DeliveryMessage != "" {
		fmt.Fprintf(builder, "Delivery: %s\n", details.DeliveryMessage)
	}
	if details.Brand != "" {
		fmt.Fprintf(builder, "Brand: %s\n", details.Brand)
	}
	if details.Publisher != "" {
		fmt.Fprintf(builder, "Publisher: %s\n", details.Publisher)
	}
	if details.PublicationDate != "" {
		fmt.Fprintf(builder, "Publication Date: %s\n", details.PublicationDate)
	}
	if details.PrintLength != "" {
		fmt.Fprintf(builder, "Length: %s\n", details.PrintLength)
	}
	if details.Dimensions != "" {
		fmt.Fprintf(builder, "Dimensions: %s\n", details.Dimensions)
	}
	if details.Language != "" {
		fmt.Fprintf(builder, "Language: %s\n", details.Language)
	}
	if details.ISBN10 != "" || details.ISBN13 != "" {
		fmt.Fprintf(builder, "ISBN-10: %s | ISBN-13: %s\n", fallback(details.ISBN10, "N/A"), fallback(details.ISBN13, "N/A"))
	}
	if details.TitleDensity > 0 {
		fmt.Fprintf(builder, "Title Density: %.2f\n", details.TitleDensity)
	}
	if len(details.BestSellerRanks) > 0 {
		sort.SliceStable(details.BestSellerRanks, func(i, j int) bool {
			return details.BestSellerRanks[i].Rank < details.BestSellerRanks[j].Rank
		})
		builder.WriteString("\nBestseller Ranks:\n")
		for _, rank := range details.BestSellerRanks {
			fmt.Fprintf(builder, "  • #%d in %s\n", rank.Rank, rank.Category)
		}
	}
	if !details.FetchedAt.IsZero() {
		fmt.Fprintf(builder, "\nLast Checked: %s\n", details.FetchedAt.Local().Format(time.RFC1123))
	}
	if details.URL != "" {
		fmt.Fprintf(builder, "Listing URL: %s\n", details.URL)
	}
	if details.IsIndependent {
		builder.WriteString("Publisher Type: Independent\n")
	}

	output := strings.TrimSpace(builder.String())
	if output == "" {
		return "No product details were returned."
	}
	return output
}

func formatKeywordInsights(insights []scraper.KeywordInsight) string {
	if len(insights) == 0 {
		return "No keyword insights available. Try broadening your filters."
	}

	sorted := append([]scraper.KeywordInsight(nil), insights...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SearchVolume == sorted[j].SearchVolume {
			return sorted[i].Keyword < sorted[j].Keyword
		}
		return sorted[i].SearchVolume > sorted[j].SearchVolume
	})

	builder := &strings.Builder{}
	for index, insight := range sorted {
		fmt.Fprintf(builder, "%d. %s\n", index+1, insight.Keyword)
		fmt.Fprintf(builder, "   Search Volume: %d\n", insight.SearchVolume)
		fmt.Fprintf(builder, "   Competition: %.2f | Relevancy: %.2f | Title Density: %.2f\n\n", insight.CompetitionScore, insight.RelevancyScore, insight.TitleDensity)
	}

	return strings.TrimSpace(builder.String())
}

func formatCategoryTrends(trends []scraper.CategoryTrend) string {
	if len(trends) == 0 {
		return "No category signals discovered."
	}

	sorted := append([]scraper.CategoryTrend(nil), trends...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Rank < sorted[j].Rank })

	builder := &strings.Builder{}
	for _, trend := range sorted {
		fmt.Fprintf(builder, "%s (Top #%d)\n", trend.Category, trend.Rank)
		if trend.Momentum != "" {
			fmt.Fprintf(builder, "   Momentum: %s\n", trend.Momentum)
		}
		if trend.Notes != "" {
			fmt.Fprintf(builder, "   Notes: %s\n", trend.Notes)
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

func formatBestsellerProducts(products []scraper.BestsellerProduct) string {
	if len(products) == 0 {
		return "No bestseller data available for the current keyword."
	}

	sorted := append([]scraper.BestsellerProduct(nil), products...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Rank < sorted[j].Rank })

	builder := &strings.Builder{}
	for _, product := range sorted {
		fmt.Fprintf(builder, "#%d %s (%s)\n", product.Rank, product.Title, product.ASIN)
		if product.Price != "" {
			fmt.Fprintf(builder, "   Price: %s\n", product.Price)
		}
		if product.Rating != "" || product.ReviewCount != "" {
			fmt.Fprintf(builder, "   Rating: %s (%s reviews)\n", fallback(product.Rating, "N/A"), fallback(product.ReviewCount, "0"))
		}
		if product.BestSeller > 0 && product.Category != "" {
			fmt.Fprintf(builder, "   Category: #%d in %s\n", product.BestSeller, product.Category)
		}
		if product.Publisher != "" {
			fmt.Fprintf(builder, "   Publisher: %s\n", product.Publisher)
		}
		if product.IsIndie {
			builder.WriteString("   Independent Author Highlight\n")
		}
		if product.TitleDensity > 0 {
			fmt.Fprintf(builder, "   Title Density: %.2f\n", product.TitleDensity)
		}
		if product.URL != "" {
			fmt.Fprintf(builder, "   URL: %s\n", product.URL)
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

func formatCampaignKeywords(keywords []string) string {
	if len(keywords) == 0 {
		return "No campaign keywords generated. Provide more metadata or competitor insights."
	}

	flagged := scraper.FlagIllegalKeywords(keywords)
	flaggedSet := map[string]struct{}{}
	for _, keyword := range flagged {
		flaggedSet[strings.ToLower(keyword)] = struct{}{}
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "Generated %d keyword(s):\n", len(keywords))
	for index, keyword := range keywords {
		marker := ""
		if _, exists := flaggedSet[strings.ToLower(keyword)]; exists {
			marker = " ⚠️"
		}
		fmt.Fprintf(builder, "%d. %s%s\n", index+1, keyword, marker)
	}

	if len(flagged) > 0 {
		builder.WriteString("\n⚠️ Keywords flagged for compliance review:\n")
		for _, keyword := range flagged {
			fmt.Fprintf(builder, "   • %s\n", keyword)
		}
	}

	return strings.TrimSpace(builder.String())
}

func formatInternationalKeywords(keywords []scraper.InternationalKeyword) string {
	if len(keywords) == 0 {
		return "No international opportunities found yet. Try selecting more marketplaces."
	}

	sorted := append([]scraper.InternationalKeyword(nil), keywords...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].CountryCode == sorted[j].CountryCode {
			return sorted[i].SearchVolume > sorted[j].SearchVolume
		}
		return sorted[i].CountryCode < sorted[j].CountryCode
	})

	builder := &strings.Builder{}
	current := ""
	for _, keyword := range sorted {
		if keyword.CountryCode != current {
			if current != "" {
				builder.WriteString("\n")
			}
			name := keyword.CountryName
			if name == "" {
				name = keyword.CountryCode
			}
			fmt.Fprintf(builder, "%s (%s)\n", name, strings.ToUpper(keyword.CountryCode))
			current = keyword.CountryCode
		}
		fmt.Fprintf(builder, "  • %s — volume %d\n", keyword.Keyword, keyword.SearchVolume)
	}

	return strings.TrimSpace(builder.String())
}

func parseKeywordFilter(minVolume, maxCompetition, maxDensity string) (scraper.KeywordFilter, error) {
	filter := scraper.KeywordFilter{}

	if trimmed := strings.TrimSpace(minVolume); trimmed != "" {
		value, err := strconv.Atoi(trimmed)
		if err != nil {
			return filter, fmt.Errorf("invalid minimum search volume: %w", err)
		}
		if value < 0 {
			return filter, errors.New("minimum search volume cannot be negative")
		}
		filter.MinSearchVolume = value
	}

	if trimmed := strings.TrimSpace(maxCompetition); trimmed != "" {
		value, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return filter, fmt.Errorf("invalid maximum competition score: %w", err)
		}
		if value < 0 {
			return filter, errors.New("maximum competition score cannot be negative")
		}
		filter.MaxCompetitionScore = value
	}

	if trimmed := strings.TrimSpace(maxDensity); trimmed != "" {
		value, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return filter, fmt.Errorf("invalid maximum title density: %w", err)
		}
		if value < 0 {
			return filter, errors.New("maximum title density cannot be negative")
		}
		filter.MaxTitleDensity = value
	}

	return filter, nil
}

func parseBestsellerFilter(maxRank string, indieOnly bool) (scraper.BestsellerFilter, error) {
	filter := scraper.BestsellerFilter{IndependentOnly: indieOnly}
	if trimmed := strings.TrimSpace(maxRank); trimmed != "" {
		value, err := strconv.Atoi(trimmed)
		if err != nil {
			return filter, fmt.Errorf("invalid max bestseller rank: %w", err)
		}
		if value < 0 {
			return filter, errors.New("max bestseller rank cannot be negative")
		}
		filter.MaxBestSellerRank = value
	}
	return filter, nil
}

func parseCSVKeywords(input string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = -1

	keywords := []string{}
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, field := range record {
			token := strings.TrimSpace(field)
			if token != "" {
				keywords = append(keywords, token)
			}
		}
	}
	return keywords, nil
}

func fallback(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return value
}

func safeSet(target binding.String, value string) {
	if err := target.Set(value); err != nil {
		fyne.LogError("failed to update binding", err)
	}
}

type queueableWindow interface {
	QueueEvent(func())
}

func queueOnMain(win fyne.Window, fn func()) {
	if fn == nil {
		return
	}

	if win != nil {
		if q, ok := win.(queueableWindow); ok {
			q.QueueEvent(fn)
			return
		}
	}

	if app := fyne.CurrentApp(); app != nil {
		for _, candidate := range app.Driver().AllWindows() {
			if q, ok := candidate.(queueableWindow); ok {
				q.QueueEvent(fn)
				return
			}
		}
	}

	fn()
}

func defaultInternationalSelection(countries []string) []string {
	preferred := []string{"US", "UK", "CA", "DE", "AU"}
	selection := []string{}
	seen := map[string]struct{}{}
	for _, code := range countries {
		upper := strings.ToUpper(code)
		for _, pref := range preferred {
			if upper == strings.ToUpper(pref) {
				if _, exists := seen[upper]; !exists {
					selection = append(selection, code)
					seen[upper] = struct{}{}
				}
			}
		}
	}
	if len(selection) == 0 && len(countries) > 0 {
		limit := len(countries)
		if limit > 3 {
			limit = 3
		}
		selection = append(selection, countries[:limit]...)
	}
	return selection
}
