package ui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/umarmf343/Umar-kdp-product-api/assets"
	"github.com/umarmf343/Umar-kdp-product-api/internal/scraper"
)

const (
	requestTimeout      = 30 * time.Second
	tutorialURL         = "https://www.youtube.com/results?search_query=RankBeam+tutorial"
	defaultScrollHeight = 280
)

func newResultScroll(content fyne.CanvasObject) *container.Scroll {
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(0, defaultScrollHeight))
	return scroll
}

var activeService *scraper.Service

// Run initialises and displays the desktop application.
func Run() {
	application := app.NewWithID("rankbeam")
	application.Settings().SetTheme(theme.LightTheme())

	if icon := assets.AppIcon(); icon != nil {
		application.SetIcon(icon)
	}

	if lifecycle := application.Lifecycle(); lifecycle != nil {
		lifecycle.SetOnStopped(func() {
			if activeService != nil {
				activeService.Close()
			}
		})
	}

	window := application.NewWindow("RankBeam")
	if icon := assets.AppIcon(); icon != nil {
		window.SetIcon(icon)
	}
	window.Resize(fyne.NewSize(1024, 720))
	window.SetMaster()

	loadMainApplication(window)
	window.ShowAndRun()
}

func loadMainApplication(window fyne.Window) {
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

	statusBinding := binding.NewString()
	statusBinding.Set("🟢 Service ready")
	quotaBinding := binding.NewString()
	quotaTracker := newQuotaTracker(25, quotaBinding)
	activityTracker := newServiceActivity(statusBinding)

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
		container.NewTabItem("Product Lookup", buildProductLookupTab(window, service, countries, productBinding, activityTracker, quotaTracker)),
		container.NewTabItem("Keyword Research", buildKeywordResearchTab(window, service, countries, keywordBinding, categoryBinding, bestsellerBinding, activityTracker, quotaTracker)),
		container.NewTabItem("Competitive Analysis", buildCompetitiveTab(window, service, countries, reverseBinding, campaignBinding, activityTracker, quotaTracker)),
		container.NewTabItem("International", buildInternationalTab(window, service, countries, internationalBinding, activityTracker, quotaTracker)),
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

	statusLabel := widget.NewLabelWithData(statusBinding)
	quotaLabel := widget.NewLabelWithData(quotaBinding)
	topBar := container.NewPadded(container.NewHBox(
		statusLabel,
		widget.NewLabel("•"),
		quotaLabel,
		layout.NewSpacer(),
		tutorialButton,
	))

	window.SetContent(container.NewBorder(topBar, nil, nil, nil, tabs))
	window.SetTitle("RankBeam")

	window.SetOnClosed(func() {
		quotaTracker.Stop()
	})
}

func buildProductLookupTab(window fyne.Window, service *scraper.Service, countries []string, result binding.String, activity *serviceActivity, quota *quotaTracker) fyne.CanvasObject {
	asinEntry := widget.NewEntry()
	asinEntry.SetPlaceHolder("B08N5WRWNW")

	countrySelect := widget.NewSelect(countries, nil)
	if len(countries) > 0 {
		countrySelect.SetSelected(countries[0])
	}

	summaryLabel := widget.NewLabelWithData(result)
	summaryLabel.Wrapping = fyne.TextWrapWord

	validationHint := canvas.NewText("ASINs are 10 characters (A-Z, 0-9).", theme.DisabledColor())
	validationHint.Alignment = fyne.TextAlignLeading

	var lastProduct *scraper.ProductDetails

	productCards := container.NewVBox(summaryLabel)

	copyButton := widget.NewButtonWithIcon("Copy Summary", theme.ContentCopyIcon(), func() {
		if lastProduct == nil {
			return
		}
		copyToClipboard(window, formatProductDetails(lastProduct))
	})
	jsonButton := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if lastProduct == nil {
			return
		}
		exportJSON(window, "product.json", lastProduct)
	})
	csvButton := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if lastProduct == nil {
			return
		}
		exportProductCSV(window, lastProduct)
	})

	controlRow := container.NewHBox(copyButton, jsonButton, csvButton)
	controlRow.Hide()

	asinEntry.OnChanged = func(value string) {
		cleaned := strings.ToUpper(strings.TrimSpace(value))
		if value != cleaned {
			asinEntry.SetText(cleaned)
			return
		}
		switch {
		case cleaned == "":
			validationHint.Text = "Enter an ASIN to begin."
			validationHint.Color = theme.DisabledColor()
		case isValidASIN(cleaned):
			validationHint.Text = "Looks good! Press Fetch Product when you're ready."
			validationHint.Color = theme.PrimaryColor()
		default:
			validationHint.Text = "ASINs must be 10 uppercase characters without spaces."
			validationHint.Color = theme.ErrorColor()
		}
		validationHint.Refresh()
	}

	fetch := func() {
		asin := strings.TrimSpace(strings.ToUpper(asinEntry.Text))
		country := strings.TrimSpace(countrySelect.Selected)
		if !isValidASIN(asin) {
			validationHint.Text = "ASINs must be 10 uppercase characters without spaces."
			validationHint.Color = theme.ErrorColor()
			validationHint.Refresh()
			return
		}
		if country == "" {
			dialog.ShowInformation("Product Lookup", "Select the Amazon marketplace you wish to query.", window)
			return
		}

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Fetching Product", fmt.Sprintf("Looking up %s on %s…", strings.ToUpper(asin), strings.ToUpper(country)), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			details, err := service.FetchProduct(ctx, asin, country)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(result, fmt.Sprintf("Unable to fetch product: %v", err))
					updateProductCards(productCards, summaryLabel, nil)
					controlRow.Hide()
					return
				}
				lastProduct = details
				safeSet(result, formatProductDetails(details))
				updateProductCards(productCards, summaryLabel, details)
				controlRow.Show()
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
		container.NewHBox(validationHint, layout.NewSpacer()),
		widget.NewSeparator(),
		controlRow,
		widget.NewSeparator(),
		newResultScroll(container.NewPadded(productCards)),
	)

	return content
}

func buildKeywordResearchTab(window fyne.Window, service *scraper.Service, countries []string, keywordResult, categoryResult, bestsellerResult binding.String, activity *serviceActivity, quota *quotaTracker) fyne.CanvasObject {
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

	keywordChart := container.NewVBox(widget.NewLabel("Keyword suggestions will appear here."))

	var lastKeywordInsights []scraper.KeywordInsight
	var lastCategoryTrends []scraper.CategoryTrend
	var lastBestsellers []scraper.BestsellerProduct

	keywordControls := container.NewHBox()
	keywordControls.Hide()
	keywordCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastKeywordInsights) == 0 {
			return
		}
		copyToClipboard(window, formatKeywordInsights(lastKeywordInsights))
	})
	keywordJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastKeywordInsights) == 0 {
			return
		}
		exportJSON(window, "keywords.json", lastKeywordInsights)
	})
	keywordCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastKeywordInsights) == 0 {
			return
		}
		exportKeywordCSV(window, lastKeywordInsights)
	})
	keywordControls.Objects = []fyne.CanvasObject{keywordCopy, keywordJSON, keywordCSV}

	categoryControls := container.NewHBox()
	categoryControls.Hide()
	categoryCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastCategoryTrends) == 0 {
			return
		}
		copyToClipboard(window, formatCategoryTrends(lastCategoryTrends))
	})
	categoryJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastCategoryTrends) == 0 {
			return
		}
		exportJSON(window, "categories.json", lastCategoryTrends)
	})
	categoryCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastCategoryTrends) == 0 {
			return
		}
		exportCategoryCSV(window, lastCategoryTrends)
	})
	categoryControls.Objects = []fyne.CanvasObject{categoryCopy, categoryJSON, categoryCSV}

	bestsellerControls := container.NewHBox()
	bestsellerControls.Hide()
	bestsellerCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastBestsellers) == 0 {
			return
		}
		copyToClipboard(window, formatBestsellerProducts(lastBestsellers))
	})
	bestsellerJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastBestsellers) == 0 {
			return
		}
		exportJSON(window, "bestsellers.json", lastBestsellers)
	})
	bestsellerCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastBestsellers) == 0 {
			return
		}
		exportBestsellerCSV(window, lastBestsellers)
	})
	bestsellerControls.Objects = []fyne.CanvasObject{bestsellerCopy, bestsellerJSON, bestsellerCSV}

	keywordInfoAction := newInfoButton("Generates keyword ideas with volume, competition and relevancy scores from Amazon auto-complete data.")
	categoryInfoAction := newInfoButton("Highlights categories where the seed term is trending so you can position listings effectively.")
	bestsellerInfoAction := newInfoButton("Summarises top selling books for the keyword to benchmark pricing, reviews and rank metrics.")

	keywordInfoHeader := newInfoButton("Generates keyword ideas with volume, competition and relevancy scores from Amazon auto-complete data.")
	categoryInfoHeader := newInfoButton("Highlights categories where the seed term is trending so you can position listings effectively.")
	bestsellerInfoHeader := newInfoButton("Summarises top selling books for the keyword to benchmark pricing, reviews and rank metrics.")

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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Keyword Research", fmt.Sprintf("Collecting ideas for \"%s\"…", seed), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			insights, err := service.KeywordSuggestions(ctx, seed, country, filters)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(keywordResult, fmt.Sprintf("Unable to fetch keyword suggestions: %v", err))
					lastKeywordInsights = nil
					keywordControls.Hide()
					keywordChart.Objects = []fyne.CanvasObject{widget.NewLabel("No keyword suggestions available yet.")}
					keywordChart.Refresh()
					return
				}
				lastKeywordInsights = insights
				keywordControls.Show()
				keywordControls.Refresh()
				safeSet(keywordResult, formatKeywordInsights(insights))
				updateKeywordChart(keywordChart, insights)
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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Category Insights", "Discovering high performing categories…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			trends, err := service.FetchCategoryTrends(ctx, seed, country)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(categoryResult, fmt.Sprintf("Unable to fetch category insights: %v", err))
					lastCategoryTrends = nil
					categoryControls.Hide()
					return
				}
				lastCategoryTrends = trends
				categoryControls.Show()
				categoryControls.Refresh()
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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Bestseller Snapshot", fmt.Sprintf("Reviewing top results for \"%s\"…", seed), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			products, err := service.BestsellerAnalysis(ctx, seed, country, filter)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(bestsellerResult, fmt.Sprintf("Unable to analyse bestsellers: %v", err))
					lastBestsellers = nil
					bestsellerControls.Hide()
					return
				}
				lastBestsellers = products
				bestsellerControls.Show()
				bestsellerControls.Refresh()
				safeSet(bestsellerResult, formatBestsellerProducts(products))
			})
		}()
	}

	keywordButton := widget.NewButton("Fetch Keyword Suggestions", fetchKeywords)
	categoryButton := widget.NewButton("Category Insights", fetchCategories)
	bestsellerButton := widget.NewButton("Bestseller Snapshot", fetchBestsellers)

	actionGrid := container.NewGridWithColumns(3,
		container.NewHBox(keywordButton, keywordInfoAction),
		container.NewHBox(categoryButton, categoryInfoAction),
		container.NewHBox(bestsellerButton, bestsellerInfoAction),
	)

	advancedFilters := widget.NewAccordion(
		widget.NewAccordionItem("Keyword Filters", container.NewVBox(
			widget.NewLabel("Fine-tune search suggestions"),
			container.NewGridWithColumns(3, minVolumeEntry, maxCompetitionEntry, maxDensityEntry),
		)),
		widget.NewAccordionItem("Bestseller Filters", container.NewVBox(
			widget.NewLabel("Limit bestseller results"),
			container.NewGridWithColumns(2, maxRankEntry, indieOnlyCheck),
		)),
	)
	advancedFilters.MultiOpen = true

	keywordOutputs := newResultScroll(container.NewVBox(
		container.NewVBox(
			container.NewHBox(widget.NewLabelWithStyle("Keyword Suggestions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), keywordInfoHeader),
			keywordControls,
			keywordChart,
			widget.NewSeparator(),
			keywordLabel,
		),
		widget.NewSeparator(),
		container.NewVBox(
			container.NewHBox(widget.NewLabelWithStyle("Category Intelligence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), categoryInfoHeader),
			categoryControls,
			categoryLabel,
		),
		widget.NewSeparator(),
		container.NewVBox(
			container.NewHBox(widget.NewLabelWithStyle("Bestseller Analysis", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), bestsellerInfoHeader),
			bestsellerControls,
			bestsellerLabel,
		),
	))

	form := widget.NewForm(
		widget.NewFormItem("Seed Keyword", keywordEntry),
		widget.NewFormItem("Marketplace", countrySelect),
	)
	form.SubmitText = "Fetch Suggestions"
	form.OnSubmit = fetchKeywords

	return container.NewVBox(
		widget.NewLabelWithStyle("Keyword Research Toolkit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		form,
		advancedFilters,
		widget.NewSeparator(),
		actionGrid,
		widget.NewSeparator(),
		keywordOutputs,
	)
}

func buildCompetitiveTab(window fyne.Window, service *scraper.Service, countries []string, reverseResult, campaignResult binding.String, activity *serviceActivity, quota *quotaTracker) fyne.CanvasObject {
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

	var lastReverseInsights []scraper.KeywordInsight
	var lastCampaignKeywords []string

	reverseControls := container.NewHBox()
	reverseControls.Hide()
	reverseCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastReverseInsights) == 0 {
			return
		}
		copyToClipboard(window, formatKeywordInsights(lastReverseInsights))
	})
	reverseJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastReverseInsights) == 0 {
			return
		}
		exportJSON(window, "reverse-asin.json", lastReverseInsights)
	})
	reverseCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastReverseInsights) == 0 {
			return
		}
		exportKeywordCSV(window, lastReverseInsights)
	})
	reverseControls.Objects = []fyne.CanvasObject{reverseCopy, reverseJSON, reverseCSV}

	campaignControls := container.NewHBox()
	campaignControls.Hide()
	campaignCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastCampaignKeywords) == 0 {
			return
		}
		copyToClipboard(window, formatCampaignKeywords(lastCampaignKeywords))
	})
	campaignJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastCampaignKeywords) == 0 {
			return
		}
		exportJSON(window, "campaign-keywords.json", lastCampaignKeywords)
	})
	campaignCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastCampaignKeywords) == 0 {
			return
		}
		exportCampaignCSV(window, lastCampaignKeywords)
	})
	campaignControls.Objects = []fyne.CanvasObject{campaignCopy, campaignJSON, campaignCSV}

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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Reverse ASIN", fmt.Sprintf("Investigating %s…", strings.ToUpper(asin)), window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			insights, err := service.ReverseASINSearch(ctx, asin, country, filters)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(reverseResult, fmt.Sprintf("Unable to run reverse ASIN: %v", err))
					lastReverseInsights = nil
					reverseControls.Hide()
					return
				}
				lastReverseInsights = insights
				reverseControls.Show()
				reverseControls.Refresh()
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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("Campaign Builder", "Composing Amazon Ads keyword list…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			keywords, err := service.GenerateAMSKeywords(ctx, titleEntry.Text, descriptionEntry.Text, competitors, country)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(campaignResult, fmt.Sprintf("Unable to generate campaign keywords: %v", err))
					lastCampaignKeywords = nil
					campaignControls.Hide()
					return
				}
				lastCampaignKeywords = keywords
				campaignControls.Show()
				campaignControls.Refresh()
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
		reverseControls,
		newResultScroll(container.NewPadded(reverseLabel)),
	)

	campaignSection := container.NewVBox(
		widget.NewLabelWithStyle("Amazon Ads Planner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		titleEntry,
		descriptionEntry,
		competitorEntry,
		campaignButton,
		widget.NewSeparator(),
		campaignControls,
		newResultScroll(container.NewPadded(campaignLabel)),
	)

	sidebar := widget.NewForm(
		widget.NewFormItem("Marketplace", countrySelect),
	)

	return container.NewBorder(sidebar, nil, nil, nil, container.NewVBox(reverseSection, widget.NewSeparator(), campaignSection))
}

func buildInternationalTab(window fyne.Window, service *scraper.Service, countries []string, result binding.String, activity *serviceActivity, quota *quotaTracker) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("mindfulness journal")

	countryGroup := widget.NewCheckGroup(countries, nil)
	defaults := defaultInternationalSelection(countries)
	if len(defaults) > 0 {
		countryGroup.SetSelected(defaults)
	}

	resultLabel := widget.NewLabelWithData(result)
	resultLabel.Wrapping = fyne.TextWrapWord

	var lastInternational []scraper.InternationalKeyword

	resultControls := container.NewHBox()
	resultControls.Hide()
	copyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastInternational) == 0 {
			return
		}
		copyToClipboard(window, formatInternationalKeywords(lastInternational))
	})
	jsonButton := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastInternational) == 0 {
			return
		}
		exportJSON(window, "international-keywords.json", lastInternational)
	})
	csvButton := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastInternational) == 0 {
			return
		}
		exportInternationalCSV(window, lastInternational)
	})
	resultControls.Objects = []fyne.CanvasObject{copyButton, jsonButton, csvButton}

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

		activity.Start()
		quota.Use()

		progress := dialog.NewProgressInfinite("International Research", "Localising your keyword list…", window)
		progress.Show()

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			keywords, err := service.InternationalKeywords(ctx, keyword, selected)

			queueOnMain(window, func() {
				activity.Done()
				progress.Hide()
				if err != nil {
					dialog.ShowError(err, window)
					safeSet(result, fmt.Sprintf("Unable to fetch international keywords: %v", err))
					lastInternational = nil
					resultControls.Hide()
					return
				}
				lastInternational = keywords
				resultControls.Show()
				resultControls.Refresh()
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
		resultControls,
		newResultScroll(container.NewPadded(resultLabel)),
	)
}

func updateProductCards(target *fyne.Container, summary *widget.Label, details *scraper.ProductDetails) {
	if target == nil {
		return
	}

	target.Objects = nil
	if details == nil {
		if summary != nil {
			target.Add(summary)
		}
		target.Refresh()
		return
	}

	cards := buildProductCards(details)
	if len(cards) > 0 {
		target.Add(container.NewGridWithColumns(2, cards...))
	}
	if summary != nil {
		target.Add(widget.NewSeparator())
		target.Add(summary)
	}
	target.Refresh()
}

func buildProductCards(details *scraper.ProductDetails) []fyne.CanvasObject {
	if details == nil {
		return nil
	}

	buildRows := func(values ...string) []fyne.CanvasObject {
		rows := make([]fyne.CanvasObject, 0, len(values))
		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				continue
			}
			rows = append(rows, widget.NewLabel(value))
		}
		return rows
	}

	title := fallback(details.Title, "Unknown Title")
	asin := strings.ToUpper(fallback(details.ASIN, "N/A"))

	overviewContent := container.NewVBox(buildRows(
		fmt.Sprintf("Brand: %s", fallback(details.Brand, "Unknown")),
		fmt.Sprintf("Availability: %s", fallback(details.Availability, "Unknown")),
		fmt.Sprintf("Publisher: %s", fallback(details.Publisher, "Unknown")),
		fmt.Sprintf("Delivery: %s", fallback(details.DeliveryMessage, "Not specified")),
		fmt.Sprintf("Listing URL: %s", fallback(details.URL, "Unavailable")),
	)...)

	currencyPrice := strings.TrimSpace(strings.TrimSpace(details.Currency + " " + details.Price))
	pricingContent := container.NewVBox(buildRows(
		fmt.Sprintf("Price: %s", fallback(currencyPrice, "Not available")),
		fmt.Sprintf("Rating: %s", fallback(details.Rating, "Unknown")),
		fmt.Sprintf("Reviews: %s", fallback(details.ReviewCount, "Unknown")),
		fmt.Sprintf("Title Density: %s", formatFloat(details.TitleDensity)),
		fmt.Sprintf("Independent Publisher: %s", boolToString(details.IsIndependent)),
	)...)

	specsContent := container.NewVBox(buildRows(
		fmt.Sprintf("Publication Date: %s", fallback(details.PublicationDate, "Unknown")),
		fmt.Sprintf("Print Length: %s", fallback(details.PrintLength, "Unknown")),
		fmt.Sprintf("Dimensions: %s", fallback(details.Dimensions, "Unknown")),
		fmt.Sprintf("Language: %s", fallback(details.Language, "Unknown")),
		fmt.Sprintf("ISBN-10: %s", fallback(details.ISBN10, "N/A")),
		fmt.Sprintf("ISBN-13: %s", fallback(details.ISBN13, "N/A")),
	)...)

	cards := []fyne.CanvasObject{
		widget.NewCard("Listing Snapshot", fmt.Sprintf("ASIN %s", asin), overviewContent),
		widget.NewCard("Pricing & Reviews", title, pricingContent),
		widget.NewCard("Product Specs", "Key attributes", specsContent),
	}

	if len(details.BestSellerRanks) > 0 {
		ranksContent := container.NewVBox()
		sorted := append([]scraper.BestSellerRank(nil), details.BestSellerRanks...)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Rank < sorted[j].Rank })
		for _, rank := range sorted {
			ranksContent.Add(widget.NewLabel(fmt.Sprintf("#%d in %s", rank.Rank, rank.Category)))
		}
		cards = append(cards, widget.NewCard("Bestseller Signals", "Top category placements", ranksContent))
	}

	return cards
}

func updateKeywordChart(chart *fyne.Container, insights []scraper.KeywordInsight) {
	if chart == nil {
		return
	}

	chart.Objects = nil
	if len(insights) == 0 {
		chart.Add(widget.NewLabel("No keyword suggestions available yet."))
		chart.Refresh()
		return
	}

	cloned := append([]scraper.KeywordInsight(nil), insights...)
	sort.SliceStable(cloned, func(i, j int) bool {
		if cloned[i].SearchVolume == cloned[j].SearchVolume {
			return cloned[i].Keyword < cloned[j].Keyword
		}
		return cloned[i].SearchVolume > cloned[j].SearchVolume
	})

	limit := len(cloned)
	if limit > 15 {
		limit = 15
	}
	maxVolume := cloned[0].SearchVolume
	if maxVolume <= 0 {
		maxVolume = 1
	}

	baseWidth := float32(280)
	for index := 0; index < limit; index++ {
		insight := cloned[index]
		ratio := float32(insight.SearchVolume) / float32(maxVolume)
		width := baseWidth * ratio
		if width < 6 {
			width = 6
		}

		bar := canvas.NewRectangle(theme.PrimaryColor())
		bar.SetMinSize(fyne.NewSize(width, 12))
		bar.CornerRadius = 6

		background := canvas.NewRectangle(theme.DisabledColor())
		background.SetMinSize(fyne.NewSize(baseWidth, 12))
		background.CornerRadius = 6

		chart.Add(container.NewVBox(
			container.NewHBox(
				widget.NewLabel(fmt.Sprintf("%d. %s", index+1, insight.Keyword)),
				layout.NewSpacer(),
				widget.NewLabel(fmt.Sprintf("%d", insight.SearchVolume)),
			),
			container.NewMax(background, bar),
			widget.NewLabel(fmt.Sprintf("Competition %.2f • Relevancy %.2f • Title Density %.2f", insight.CompetitionScore, insight.RelevancyScore, insight.TitleDensity)),
		))
	}
	chart.Refresh()
}

func newInfoButton(tooltip string) *widget.Button {
	button := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {})
	button.Importance = widget.LowImportance
	button.SetTooltip(tooltip)
	return button
}

func copyToClipboard(window fyne.Window, content string) {
	if window == nil {
		return
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return
	}
	if clip := window.Clipboard(); clip != nil {
		clip.SetContent(trimmed)
	}
}

func exportJSON(window fyne.Window, fileName string, data interface{}) {
	if window == nil || data == nil {
		return
	}

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			dialog.ShowError(fmt.Errorf("failed to export JSON: %w", err), window)
		}
	}, window)
	saveDialog.SetFileName(fileName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	saveDialog.Show()
}

func exportCSV(window fyne.Window, fileName string, headers []string, rows [][]string) {
	if window == nil || len(rows) == 0 {
		return
	}

	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		csvWriter := csv.NewWriter(writer)
		if len(headers) > 0 {
			if err := csvWriter.Write(headers); err != nil {
				dialog.ShowError(fmt.Errorf("failed to write CSV header: %w", err), window)
				return
			}
		}
		for _, row := range rows {
			if err := csvWriter.Write(row); err != nil {
				dialog.ShowError(fmt.Errorf("failed to write CSV row: %w", err), window)
				return
			}
		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			dialog.ShowError(fmt.Errorf("failed to finalise CSV: %w", err), window)
		}
	}, window)
	saveDialog.SetFileName(fileName)
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	saveDialog.Show()
}

func exportProductCSV(window fyne.Window, details *scraper.ProductDetails) {
	if details == nil {
		return
	}
	ranks := make([]string, 0, len(details.BestSellerRanks))
	for _, rank := range details.BestSellerRanks {
		ranks = append(ranks, fmt.Sprintf("#%d in %s", rank.Rank, rank.Category))
	}

	row := []string{
		strings.ToUpper(fallback(details.ASIN, "")),
		fallback(details.Title, ""),
		fallback(details.Price, ""),
		fallback(details.Currency, ""),
		fallback(details.Rating, ""),
		fallback(details.ReviewCount, ""),
		fallback(details.Availability, ""),
		fallback(details.Brand, ""),
		fallback(details.Publisher, ""),
		fallback(details.PublicationDate, ""),
		fallback(details.PrintLength, ""),
		fallback(details.Dimensions, ""),
		fallback(details.Language, ""),
		formatFloat(details.TitleDensity),
		boolToString(details.IsIndependent),
		fallback(details.URL, ""),
		strings.Join(ranks, " | "),
	}

	exportCSV(window, "product.csv", []string{
		"ASIN", "Title", "Price", "Currency", "Rating", "ReviewCount", "Availability", "Brand", "Publisher", "PublicationDate", "PrintLength", "Dimensions", "Language", "TitleDensity", "Independent", "URL", "BestsellerRanks",
	}, [][]string{row})
}

func exportKeywordCSV(window fyne.Window, insights []scraper.KeywordInsight) {
	if len(insights) == 0 {
		return
	}
	rows := make([][]string, 0, len(insights))
	for _, insight := range insights {
		rows = append(rows, []string{
			insight.Keyword,
			strconv.Itoa(insight.SearchVolume),
			formatFloat(insight.CompetitionScore),
			formatFloat(insight.RelevancyScore),
			formatFloat(insight.TitleDensity),
		})
	}
	exportCSV(window, "keywords.csv", []string{"Keyword", "SearchVolume", "Competition", "Relevancy", "TitleDensity"}, rows)
}

func exportCategoryCSV(window fyne.Window, trends []scraper.CategoryTrend) {
	if len(trends) == 0 {
		return
	}
	rows := make([][]string, 0, len(trends))
	for _, trend := range trends {
		rows = append(rows, []string{
			trend.Category,
			strconv.Itoa(trend.Rank),
			trend.Momentum,
			trend.Notes,
		})
	}
	exportCSV(window, "categories.csv", []string{"Category", "Rank", "Momentum", "Notes"}, rows)
}

func exportBestsellerCSV(window fyne.Window, products []scraper.BestsellerProduct) {
	if len(products) == 0 {
		return
	}
	rows := make([][]string, 0, len(products))
	for _, product := range products {
		rows = append(rows, []string{
			strconv.Itoa(product.Rank),
			product.ASIN,
			product.Title,
			product.Price,
			product.Rating,
			product.ReviewCount,
			product.Category,
			strconv.Itoa(product.BestSeller),
			product.Publisher,
			formatFloat(product.TitleDensity),
			boolToString(product.IsIndie),
			product.URL,
		})
	}
	exportCSV(window, "bestsellers.csv", []string{"Rank", "ASIN", "Title", "Price", "Rating", "Reviews", "Category", "BSR", "Publisher", "TitleDensity", "Independent", "URL"}, rows)
}

func exportCampaignCSV(window fyne.Window, keywords []string) {
	if len(keywords) == 0 {
		return
	}
	rows := make([][]string, 0, len(keywords))
	for index, keyword := range keywords {
		rows = append(rows, []string{strconv.Itoa(index + 1), keyword})
	}
	exportCSV(window, "campaign-keywords.csv", []string{"Position", "Keyword"}, rows)
}

func exportInternationalCSV(window fyne.Window, keywords []scraper.InternationalKeyword) {
	if len(keywords) == 0 {
		return
	}
	rows := make([][]string, 0, len(keywords))
	for _, keyword := range keywords {
		rows = append(rows, []string{
			keyword.CountryName,
			strings.ToUpper(keyword.CountryCode),
			keyword.Keyword,
			strconv.Itoa(keyword.SearchVolume),
		})
	}
	exportCSV(window, "international-keywords.csv", []string{"Country", "Code", "Keyword", "SearchVolume"}, rows)
}

func boolToString(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func formatFloat(value float64) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", value)
}

func isValidASIN(asin string) bool {
	if len(asin) != 10 {
		return false
	}
	for _, r := range asin {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
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

type serviceActivity struct {
	mu      sync.Mutex
	binding binding.String
	active  int
}

func newServiceActivity(binding binding.String) *serviceActivity {
	return &serviceActivity{binding: binding}
}

func (s *serviceActivity) Start() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.active++
	status := "🟡 Running request…"
	if s.active > 1 {
		status = fmt.Sprintf("🟡 Running %d requests…", s.active)
	}
	if s.binding != nil {
		_ = s.binding.Set(status)
	}
	s.mu.Unlock()
}

func (s *serviceActivity) Done() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.active > 0 {
		s.active--
	}
	status := "🟢 Service ready"
	if s.active > 0 {
		status = fmt.Sprintf("🟡 Running %d request(s)…", s.active)
	}
	if s.binding != nil {
		_ = s.binding.Set(status)
	}
	s.mu.Unlock()
}

type quotaTracker struct {
	mu        sync.Mutex
	capacity  int
	remaining int
	binding   binding.String
	ticker    *time.Ticker
	stop      chan struct{}
	stopOnce  sync.Once
}

func newQuotaTracker(capacity int, binding binding.String) *quotaTracker {
	if capacity <= 0 {
		capacity = 20
	}
	tracker := &quotaTracker{
		capacity:  capacity,
		remaining: capacity,
		binding:   binding,
		ticker:    time.NewTicker(time.Minute),
		stop:      make(chan struct{}),
	}
	tracker.updateBindingLocked()
	go tracker.loop()
	return tracker
}

func (q *quotaTracker) loop() {
	if q == nil {
		return
	}
	for {
		select {
		case <-q.stop:
			return
		case <-q.ticker.C:
			q.mu.Lock()
			q.remaining = q.capacity
			q.updateBindingLocked()
			q.mu.Unlock()
		}
	}
}

func (q *quotaTracker) Use() {
	if q == nil {
		return
	}
	q.mu.Lock()
	if q.remaining > 0 {
		q.remaining--
	}
	q.updateBindingLocked()
	q.mu.Unlock()
}

func (q *quotaTracker) Stop() {
	if q == nil {
		return
	}
	q.stopOnce.Do(func() {
		if q.ticker != nil {
			q.ticker.Stop()
		}
		if q.stop != nil {
			close(q.stop)
		}
	})
}

func (q *quotaTracker) updateBindingLocked() {
	if q == nil || q.binding == nil {
		return
	}
	status := fmt.Sprintf("Quota: %d/%d requests remaining", q.remaining, q.capacity)
	if q.remaining == 0 {
		status += " (queuing new requests)"
	}
	_ = q.binding.Set(status)
}
