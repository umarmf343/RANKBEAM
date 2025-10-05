package ui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
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
	tutorialURL         = "https://www.youtube.com/results?search_query=RankBeam+tutorial"
	defaultScrollHeight = 280
)

func newResultScroll(content fyne.CanvasObject) *container.Scroll {
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(0, defaultScrollHeight))
	return scroll
}

func currentRequestTimeout() time.Duration {
	return throttleState.Timeout()
}

var (
	activeService *scraper.Service
	activeQuota   *quotaTracker
	throttleState = newThrottleConfig(30*time.Second, 25)
)

type throttleSettings struct {
	Timeout           time.Duration
	RequestsPerMinute int
}

type throttleConfig struct {
	mu       sync.RWMutex
	settings throttleSettings
}

func newThrottleConfig(timeout time.Duration, rpm int) *throttleConfig {
	cfg := &throttleConfig{}
	cfg.Update(timeout, rpm)
	return cfg
}

func (c *throttleConfig) Snapshot() throttleSettings {
	if c == nil {
		return throttleSettings{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}

func (c *throttleConfig) Timeout() time.Duration {
	return c.Snapshot().Timeout
}

func (c *throttleConfig) RequestsPerMinute() int {
	return c.Snapshot().RequestsPerMinute
}

func (c *throttleConfig) Update(timeout time.Duration, rpm int) (throttleSettings, bool) {
	if c == nil {
		return throttleSettings{}, false
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if rpm <= 0 {
		rpm = 20
	}
	c.mu.Lock()
	changed := c.settings.Timeout != timeout || c.settings.RequestsPerMinute != rpm
	c.settings = throttleSettings{Timeout: timeout, RequestsPerMinute: rpm}
	snapshot := c.settings
	c.mu.Unlock()
	return snapshot, changed
}

type keywordPreset struct {
	Name           string
	Seed           string
	Country        string
	MinVolume      string
	MaxCompetition string
	MaxDensity     string
	MaxRank        string
	IndieOnly      bool
	Format         string
}

var (
	keywordPresetLock sync.Mutex
	keywordPresets    []keywordPreset
)

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
	if activeQuota != nil {
		activeQuota.Stop()
		activeQuota = nil
	}

	throttle := throttleState.Snapshot()

	service := scraper.NewService(throttle.Timeout, throttle.RequestsPerMinute)
	activeService = service

	countries := scraper.Countries()
	sort.Strings(countries)

	statusBinding := binding.NewString()
	statusBinding.Set("🟢 Service ready")
	quotaBinding := binding.NewString()
	quotaTracker := newQuotaTracker(throttle.RequestsPerMinute, quotaBinding)
	activityTracker := newServiceActivity(statusBinding)
	activeQuota = quotaTracker

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
	settingsButton := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		showThrottleSettingsDialog(window)
	})
	settingsButton.Importance = widget.MediumImportance

	topBar := container.NewPadded(container.NewHBox(
		statusLabel,
		widget.NewLabel("•"),
		quotaLabel,
		layout.NewSpacer(),
		settingsButton,
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

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("children's book about space")

	searchFormatOptions := []string{"All Books", "Kindle", "Paperback", "Hardcover"}
	searchFormatSelect := widget.NewSelect(searchFormatOptions, nil)
	searchFormatSelect.SetSelected(searchFormatOptions[0])

	maxResultsEntry := widget.NewEntry()
	maxResultsEntry.SetPlaceHolder("15")

	searchBinding := binding.NewString()
	searchBinding.Set("Enter a keyword and press Search Catalog to capture live listings.")
	searchStatusLabel := widget.NewLabelWithData(searchBinding)
	searchStatusLabel.Wrapping = fyne.TextWrapWord

	var lastProduct *scraper.ProductDetails
	var lastSearchResults []scraper.SearchResult

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

	searchCopy := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if len(lastSearchResults) == 0 {
			return
		}
		copyToClipboard(window, formatSearchResults(lastSearchResults))
	})
	searchJSON := widget.NewButtonWithIcon("Export JSON", theme.DocumentIcon(), func() {
		if len(lastSearchResults) == 0 {
			return
		}
		exportJSON(window, "catalog-search.json", lastSearchResults)
	})
	searchCSV := widget.NewButtonWithIcon("Export CSV", theme.DocumentIcon(), func() {
		if len(lastSearchResults) == 0 {
			return
		}
		exportSearchCSV(window, lastSearchResults)
	})

	searchControls := container.NewHBox(searchCopy, searchJSON, searchCSV)
	searchControls.Hide()

	searchTableData := make([]scraper.SearchResult, 0)
	searchTable := widget.NewTable(
		func() (int, int) {
			rows := len(searchTableData)
			if rows == 0 {
				return 1, 8
			}
			return rows + 1, 8
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignLeading
			label.Wrapping = fyne.TextWrapOff
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			label, _ := object.(*widget.Label)
			if label == nil {
				return
			}

			label.Alignment = fyne.TextAlignLeading
			label.Wrapping = fyne.TextWrapOff
			label.TextStyle = fyne.TextStyle{}

			headers := []string{"Rank", "Title", "Author", "Price", "Rating", "Reviews", "Best Seller Rank", "URL"}
			if id.Row == 0 {
				if id.Col < len(headers) {
					label.TextStyle = fyne.TextStyle{Bold: true}
					if id.Col == 0 {
						label.Alignment = fyne.TextAlignCenter
					} else if id.Col >= 3 {
						label.Alignment = fyne.TextAlignCenter
					}
					label.SetText(headers[id.Col])
				}
				return
			}

			if id.Row-1 >= len(searchTableData) {
				label.SetText("")
				return
			}

			result := searchTableData[id.Row-1]
			switch id.Col {
			case 0:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(strconv.Itoa(result.Rank))
			case 1:
				label.Alignment = fyne.TextAlignLeading
				label.SetText(result.Title)
			case 2:
				label.Alignment = fyne.TextAlignLeading
				label.SetText(result.Author)
			case 3:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(result.Price)
			case 4:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(result.Rating)
			case 5:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(result.ReviewCount)
			case 6:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(result.BestSellerRank)
			case 7:
				label.Alignment = fyne.TextAlignLeading
				label.SetText(result.URL)
			}
		},
	)
	searchTable.SetColumnWidth(0, 70)
	searchTable.SetColumnWidth(1, 220)
	searchTable.SetColumnWidth(2, 180)
	searchTable.SetColumnWidth(3, 110)
	searchTable.SetColumnWidth(4, 140)
	searchTable.SetColumnWidth(5, 140)
	searchTable.SetColumnWidth(6, 180)
	searchTable.SetColumnWidth(7, 260)

	searchTableContainer := container.NewPadded(searchTable)
	searchTableContainer.Hide()
	searchStatusContainer := container.NewPadded(searchStatusLabel)
	searchResultStack := container.NewStack(searchStatusContainer, searchTableContainer)

	normalizeSearchFormat := func(selected string) string {
		trimmed := strings.TrimSpace(selected)
		for _, option := range searchFormatOptions {
			if option == trimmed {
				return option
			}
		}
		return searchFormatOptions[0]
	}

	resolveSearchSeed := func(seed, selectedFormat string) (string, string) {
		trimmed := strings.TrimSpace(seed)
		if trimmed == "" {
			return "", "stripbooks"
		}

		alias := "stripbooks"
		switch normalizeSearchFormat(selectedFormat) {
		case "Kindle":
			alias = "digital-text"
			trimmed = strings.TrimSpace(fmt.Sprintf("%s kindle edition", trimmed))
		case "Paperback":
			trimmed = strings.TrimSpace(fmt.Sprintf("%s paperback", trimmed))
		case "Hardcover":
			trimmed = strings.TrimSpace(fmt.Sprintf("%s hardcover", trimmed))
		}

		return trimmed, alias
	}

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

	fetchProduct := func() {
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

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Fetching Product", fmt.Sprintf("Looking up %s on %s…", strings.ToUpper(asin), strings.ToUpper(country)), cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			details, err := service.FetchProduct(ctx, asin, country)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if errors.Is(err, context.Canceled) {
						safeSet(result, "Request cancelled.")
					} else {
						dialog.ShowError(err, window)
						safeSet(result, fmt.Sprintf("Unable to fetch product: %v", err))
					}
					lastProduct = nil
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

	fetchSearch := func() {
		keyword := strings.TrimSpace(searchEntry.Text)
		country := strings.TrimSpace(countrySelect.Selected)
		if keyword == "" {
			dialog.ShowInformation("Catalog Search", "Enter a keyword to search.", window)
			return
		}
		if country == "" {
			dialog.ShowInformation("Catalog Search", "Select the Amazon marketplace you wish to query.", window)
			return
		}

		maxResults := 15
		if text := strings.TrimSpace(maxResultsEntry.Text); text != "" {
			value, err := strconv.Atoi(text)
			if err != nil || value <= 0 {
				dialog.ShowError(fmt.Errorf("Max results must be a positive number"), window)
				return
			}
			maxResults = value
		}

		searchSeed, alias := resolveSearchSeed(keyword, searchFormatSelect.Selected)
		if searchSeed == "" {
			dialog.ShowInformation("Catalog Search", "Provide a keyword to search.", window)
			return
		}

		activity.Start()
		quota.Use()

		safeSet(searchBinding, fmt.Sprintf("Searching for \"%s\"…", searchSeed))
		searchControls.Hide()
		searchTableContainer.Hide()
		searchStatusContainer.Show()
		searchResultStack.Refresh()
		searchTableData = searchTableData[:0]
		searchTable.Refresh()

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Catalog Search", fmt.Sprintf("Collecting listings for \"%s\"…", searchSeed), cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			results, err := service.SearchProducts(ctx, searchSeed, country, alias, maxResults)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to search catalog: %v", err)
					}
					safeSet(searchBinding, message)
					lastSearchResults = nil
					searchControls.Hide()
					searchTableData = searchTableData[:0]
					searchTable.Refresh()
					searchTableContainer.Hide()
					searchStatusContainer.Show()
					searchResultStack.Refresh()
					return
				}

				searchTableData = append(searchTableData[:0], results...)
				lastSearchResults = append(lastSearchResults[:0], results...)
				if len(searchTableData) == 0 {
					safeSet(searchBinding, "No catalog listings returned. Try another keyword.")
					searchControls.Hide()
					searchTableContainer.Hide()
					searchStatusContainer.Show()
					searchResultStack.Refresh()
					return
				}

				searchControls.Show()
				searchControls.Refresh()
				searchTableContainer.Show()
				searchTable.Refresh()
				searchStatusContainer.Hide()
				searchResultStack.Refresh()
				safeSet(searchBinding, fmt.Sprintf("Found %d listing(s). Use the export controls for deeper analysis.", len(searchTableData)))
			})
		}()
	}

	searchEntry.OnSubmitted = func(string) {
		fetchSearch()
	}

	form := widget.NewForm(
		widget.NewFormItem("ASIN", asinEntry),
		widget.NewFormItem("Marketplace", countrySelect),
	)
	form.SubmitText = "Fetch Product"
	form.OnSubmit = fetchProduct

	searchForm := widget.NewForm(
		widget.NewFormItem("Keyword", searchEntry),
		widget.NewFormItem("Format", searchFormatSelect),
		widget.NewFormItem("Max Results", maxResultsEntry),
	)
	searchForm.SubmitText = "Search Catalog"
	searchForm.OnSubmit = func() {
		fetchSearch()
	}

	content := container.NewVBox(
		form,
		container.NewHBox(validationHint, layout.NewSpacer()),
		widget.NewSeparator(),
		controlRow,
		widget.NewSeparator(),
		newResultScroll(container.NewPadded(productCards)),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Catalog Search", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		searchForm,
		searchControls,
		newResultScroll(searchResultStack),
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
	maxCompetitionEntry.SetPlaceHolder("5")
	maxDensityEntry := widget.NewEntry()
	maxDensityEntry.SetPlaceHolder("100")

	const (
		bookFormatAll         = "All Books"
		bookFormatKindle      = "Kindle"
		bookFormatPaperback   = "Paperback"
		bookFormatHardcover   = "Hardcover"
		presetCollapsedOffset = 0.02
		presetDefaultOffset   = 0.28
		presetExpandedOffset  = 0.4
	)

	formatOptions := []string{bookFormatAll, bookFormatKindle, bookFormatPaperback, bookFormatHardcover}
	formatSelect := widget.NewSelect(formatOptions, nil)
	formatSelect.SetSelected(bookFormatAll)

	normalizeFormat := func(selected string) string {
		trimmed := strings.TrimSpace(selected)
		for _, option := range formatOptions {
			if option == trimmed {
				return option
			}
		}
		return bookFormatAll
	}

	resolveBookSearch := func(seed, selectedFormat string) (string, string) {
		trimmed := strings.TrimSpace(seed)
		if trimmed == "" {
			return "", "stripbooks"
		}

		alias := "stripbooks"
		qualifier := ""

		switch selected := normalizeFormat(selectedFormat); selected {
		case bookFormatKindle:
			alias = "digital-text"
			qualifier = "kindle edition"
		case bookFormatPaperback:
			qualifier = "paperback"
		case bookFormatHardcover:
			qualifier = "hardcover"
		default:
			alias = "stripbooks"
		}

		if qualifier != "" {
			trimmed = strings.TrimSpace(fmt.Sprintf("%s %s", trimmed, qualifier))
		}

		return trimmed, alias
	}

	maxRankEntry := widget.NewEntry()
	maxRankEntry.SetPlaceHolder("50000")
	indieOnlyCheck := widget.NewCheck("Indie authors only", nil)

	keywordLabel := widget.NewLabelWithData(keywordResult)
	keywordLabel.Wrapping = fyne.TextWrapWord
	categoryLabel := widget.NewLabelWithData(categoryResult)
	categoryLabel.Wrapping = fyne.TextWrapWord
	bestsellerLabel := widget.NewLabelWithData(bestsellerResult)
	bestsellerLabel.Wrapping = fyne.TextWrapWord

	presetNameEntry := widget.NewEntry()
	presetNameEntry.SetPlaceHolder("Preset name")

	presetListData := binding.NewStringList()
	if err := presetListData.Set(listKeywordPresetNames()); err != nil {
		fyne.LogError("unable to populate preset list", err)
	}
	selectedPresetIndex := -1

	presetList := widget.NewListWithData(presetListData,
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return container.NewPadded(label)
		},
		func(data binding.DataItem, item fyne.CanvasObject) {
			bindingString, ok := data.(binding.String)
			if !ok {
				return
			}
			text, err := bindingString.Get()
			if err != nil {
				fyne.LogError("unable to read preset name", err)
				return
			}
			if padded, ok := item.(*fyne.Container); ok && len(padded.Objects) > 0 {
				if label, ok := padded.Objects[0].(*widget.Label); ok {
					label.SetText(text)
				}
			}
		})
	presetListScroll := container.NewVScroll(presetList)
	presetListScroll.SetMinSize(fyne.NewSize(220, 240))

	refreshPresetList := func(targetName string) {
		names := listKeywordPresetNames()
		if err := presetListData.Set(names); err != nil {
			fyne.LogError("unable to refresh preset names", err)
		}
		selectedPresetIndex = -1
		if strings.TrimSpace(targetName) == "" {
			return
		}
		for i, name := range names {
			if strings.EqualFold(name, targetName) {
				presetList.Select(i)
				selectedPresetIndex = i
				return
			}
		}
	}

	presetList.OnSelected = func(id widget.ListItemID) {
		preset, ok := getKeywordPreset(int(id))
		if !ok {
			return
		}
		selectedPresetIndex = int(id)
		presetNameEntry.SetText(preset.Name)
		keywordEntry.SetText(preset.Seed)
		if preset.Country != "" {
			for _, option := range countries {
				if option == preset.Country {
					countrySelect.SetSelected(preset.Country)
					break
				}
			}
		}
		minVolumeEntry.SetText(preset.MinVolume)
		maxCompetitionEntry.SetText(preset.MaxCompetition)
		maxDensityEntry.SetText(preset.MaxDensity)
		maxRankEntry.SetText(preset.MaxRank)
		indieOnlyCheck.SetChecked(preset.IndieOnly)
		formatSelect.SetSelected(normalizeFormat(preset.Format))
	}

	presetList.OnUnselected = func(id widget.ListItemID) {
		if selectedPresetIndex == int(id) {
			selectedPresetIndex = -1
		}
	}

	presetInfo := widget.NewLabel("Bookmark frequent keyword settings and reload them instantly.")
	presetInfo.Wrapping = fyne.TextWrapWord

	savePresetButton := widget.NewButtonWithIcon("Save Preset", theme.ContentAddIcon(), func() {
		name := strings.TrimSpace(presetNameEntry.Text)
		if name == "" {
			dialog.ShowInformation("Research Presets", "Enter a preset name before saving.", window)
			return
		}
		seed := strings.TrimSpace(keywordEntry.Text)
		if seed == "" {
			dialog.ShowInformation("Research Presets", "Provide a seed keyword to save.", window)
			return
		}

		preset := keywordPreset{
			Name:           name,
			Seed:           seed,
			Country:        strings.TrimSpace(countrySelect.Selected),
			MinVolume:      strings.TrimSpace(minVolumeEntry.Text),
			MaxCompetition: strings.TrimSpace(maxCompetitionEntry.Text),
			MaxDensity:     strings.TrimSpace(maxDensityEntry.Text),
			MaxRank:        strings.TrimSpace(maxRankEntry.Text),
			IndieOnly:      indieOnlyCheck.Checked,
			Format:         normalizeFormat(formatSelect.Selected),
		}

		saveKeywordPreset(preset)
		refreshPresetList(preset.Name)
		presetNameEntry.SetText(preset.Name)
	})

	deletePresetButton := widget.NewButtonWithIcon("Delete Preset", theme.DeleteIcon(), func() {
		if selectedPresetIndex < 0 {
			dialog.ShowInformation("Research Presets", "Select a preset to delete.", window)
			return
		}
		if !deleteKeywordPreset(selectedPresetIndex) {
			dialog.ShowError(errors.New("unable to delete preset"), window)
			return
		}
		refreshPresetList("")
		presetNameEntry.SetText("")
	})

	var lastKeywordInsights []scraper.KeywordInsight
	keywordTableData := make([]scraper.KeywordInsight, 0)
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

	keywordTable := widget.NewTable(
		func() (int, int) {
			rows := len(keywordTableData)
			if rows == 0 {
				return 1, 5
			}
			return rows + 1, 5
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignLeading
			label.Wrapping = fyne.TextWrapOff
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			label, _ := object.(*widget.Label)
			if label == nil {
				return
			}

			label.Alignment = fyne.TextAlignLeading
			label.Wrapping = fyne.TextWrapOff
			label.TextStyle = fyne.TextStyle{}

			if id.Row == 0 {
				headers := []string{"Keyword", "Search Volume", "Competition", "Relevancy", "Title Density"}
				if id.Col < len(headers) {
					text := headers[id.Col]
					if id.Col > 1 {
						label.Alignment = fyne.TextAlignCenter
					}
					label.TextStyle = fyne.TextStyle{Bold: true}
					label.SetText(text)
				}
				return
			}

			if id.Row-1 >= len(keywordTableData) {
				label.SetText("")
				return
			}

			insight := keywordTableData[id.Row-1]
			switch id.Col {
			case 0:
				label.Alignment = fyne.TextAlignLeading
				label.SetText(insight.Keyword)
			case 1:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(fmt.Sprintf("%d", insight.SearchVolume))
			case 2:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(keywordCompetitionBadge(insight.CompetitionScore))
			case 3:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(fmt.Sprintf("%.0f%%", insight.RelevancyScore*100))
			case 4:
				label.Alignment = fyne.TextAlignCenter
				label.SetText(keywordDensityBadge(insight.TitleDensity))
			}
		},
	)
	keywordTable.SetColumnWidth(0, 260)
	keywordTable.SetColumnWidth(1, 130)
	keywordTable.SetColumnWidth(2, 140)
	keywordTable.SetColumnWidth(3, 130)
	keywordTable.SetColumnWidth(4, 140)

	keywordTableContainer := container.NewPadded(keywordTable)
	keywordTableContainer.Hide()
	keywordStatusContainer := container.NewPadded(keywordLabel)
	keywordResultStack := container.NewStack(keywordTableContainer, keywordStatusContainer)

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

	categoryInfoAction := newInfoButton(window, "Highlights categories where the seed term is trending so you can position listings effectively.")
	bestsellerInfoAction := newInfoButton(window, "Summarises top selling books for the keyword to benchmark pricing, reviews and rank metrics.")

	keywordInfoHeader := newInfoButton(window, "Generates keyword ideas with volume, competition counts and relevancy scores from Amazon auto-complete data.")
	categoryInfoHeader := newInfoButton(window, "Highlights categories where the seed term is trending so you can position listings effectively.")
	bestsellerInfoHeader := newInfoButton(window, "Summarises top selling books for the keyword to benchmark pricing, reviews and rank metrics.")

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

		format := normalizeFormat(formatSelect.Selected)
		searchSeed, alias := resolveBookSearch(seed, format)

		filters, err := parseKeywordFilter(minVolumeEntry.Text, maxCompetitionEntry.Text, maxDensityEntry.Text)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		filters.SearchAlias = alias

		activity.Start()
		quota.Use()

		safeSet(keywordResult, fmt.Sprintf("Collecting ideas for \"%s\"…", searchSeed))
		keywordControls.Hide()
		keywordTableContainer.Hide()
		keywordStatusContainer.Show()
		keywordResultStack.Refresh()
		keywordTableData = keywordTableData[:0]
		keywordTable.Refresh()

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Keyword Research", fmt.Sprintf("Collecting ideas for \"%s\"…", searchSeed), cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			insights, err := service.KeywordSuggestions(ctx, searchSeed, country, filters)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to fetch keyword suggestions: %v", err)
					}
					safeSet(keywordResult, message)
					lastKeywordInsights = nil
					keywordControls.Hide()
					keywordTableData = keywordTableData[:0]
					keywordTableContainer.Hide()
					keywordTable.Refresh()
					keywordStatusContainer.Show()
					keywordResultStack.Refresh()
					return
				}
				keywordTableData = append(keywordTableData[:0], insights...)
				sortKeywordInsights(keywordTableData)
				lastKeywordInsights = append(lastKeywordInsights[:0], keywordTableData...)
				keywordControls.Show()
				keywordControls.Refresh()
				keywordTableContainer.Show()
				keywordTable.Refresh()
				keywordStatusContainer.Hide()
				keywordResultStack.Refresh()
				safeSet(keywordResult, keywordSummaryMessage(len(keywordTableData)))
			})
		}()
	}

	keywordEntry.OnSubmitted = func(string) {
		fetchKeywords()
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

		format := normalizeFormat(formatSelect.Selected)
		searchSeed, alias := resolveBookSearch(seed, format)

		activity.Start()
		quota.Use()

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Category Insights", "Discovering high performing categories…", cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			trends, err := service.FetchCategoryTrends(ctx, searchSeed, country, alias)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to fetch category insights: %v", err)
					}
					safeSet(categoryResult, message)
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

		format := normalizeFormat(formatSelect.Selected)
		searchSeed, alias := resolveBookSearch(seed, format)

		filter, err := parseBestsellerFilter(maxRankEntry.Text, indieOnlyCheck.Checked)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		activity.Start()
		quota.Use()

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Bestseller Snapshot", fmt.Sprintf("Reviewing top results for \"%s\"…", searchSeed), cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			products, err := service.BestsellerAnalysis(ctx, searchSeed, country, alias, filter)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to analyse bestsellers: %v", err)
					}
					safeSet(bestsellerResult, message)
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

	labeledField := func(label string, content fyne.CanvasObject) fyne.CanvasObject {
		title := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		return container.NewVBox(title, content)
	}

	newSearchButton := widget.NewButtonWithIcon("New Search", theme.SearchIcon(), func() {
		fetchKeywords()
	})
	newSearchButton.Importance = widget.HighImportance

	exportButton := widget.NewButtonWithIcon("Export", theme.DocumentIcon(), func() {
		if len(lastKeywordInsights) == 0 {
			dialog.ShowInformation("Export Keywords", "Run a keyword search before exporting results.", window)
			return
		}

		var exportDialog dialog.Dialog
		options := container.NewVBox(
			widget.NewButtonWithIcon("Copy to Clipboard", theme.ContentCopyIcon(), func() {
				copyToClipboard(window, formatKeywordInsights(lastKeywordInsights))
				if exportDialog != nil {
					exportDialog.Hide()
				}
			}),
			widget.NewButtonWithIcon("Download JSON", theme.DocumentIcon(), func() {
				exportJSON(window, "keywords.json", lastKeywordInsights)
				if exportDialog != nil {
					exportDialog.Hide()
				}
			}),
			widget.NewButtonWithIcon("Download CSV", theme.DocumentIcon(), func() {
				exportKeywordCSV(window, lastKeywordInsights)
				if exportDialog != nil {
					exportDialog.Hide()
				}
			}),
		)

		exportDialog = dialog.NewCustomWithoutButtons("Export Keywords", options, window)
		exportDialog.Show()
	})

	searchBar := container.NewAdaptiveGrid(4,
		labeledField("Seed Keyword", keywordEntry),
		labeledField("Marketplace", countrySelect),
		labeledField("Format", formatSelect),
		container.NewVBox(widget.NewLabel(" "), container.NewHBox(newSearchButton, exportButton)),
	)

	keywordFilters := container.NewAdaptiveGrid(3,
		labeledField("Minimum Volume", minVolumeEntry),
		labeledField("Max Competition", maxCompetitionEntry),
		labeledField("Max Title Density", maxDensityEntry),
	)

	bestsellerFilters := container.NewAdaptiveGrid(2,
		labeledField("Max Bestseller Rank", maxRankEntry),
		container.NewVBox(widget.NewLabelWithStyle("Options", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), indieOnlyCheck),
	)

	categoryButton := widget.NewButton("Category Insights", fetchCategories)
	bestsellerButton := widget.NewButton("Bestseller Snapshot", fetchBestsellers)

	actionGrid := container.NewAdaptiveGrid(2,
		container.NewHBox(categoryButton, categoryInfoAction),
		container.NewHBox(bestsellerButton, bestsellerInfoAction),
	)

	keywordHeader := container.NewHBox(
		widget.NewLabelWithStyle("Keyword Suggestions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keywordInfoHeader,
		layout.NewSpacer(),
		keywordControls,
	)

	keywordTabContent := newResultScroll(container.NewVBox(
		keywordHeader,
		widget.NewSeparator(),
		keywordResultStack,
	))

	categoryTabContent := newResultScroll(container.NewVBox(
		container.NewHBox(widget.NewLabelWithStyle("Category Intelligence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), categoryInfoHeader),
		categoryControls,
		categoryLabel,
	))

	bestsellerTabContent := newResultScroll(container.NewVBox(
		container.NewHBox(widget.NewLabelWithStyle("Bestseller Analysis", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), bestsellerInfoHeader),
		bestsellerControls,
		bestsellerLabel,
	))

	keywordOutputs := container.NewAppTabs(
		container.NewTabItem("Keywords", keywordTabContent),
		container.NewTabItem("Categories", categoryTabContent),
		container.NewTabItem("Bestsellers", bestsellerTabContent),
	)
	keywordOutputs.SetTabLocation(container.TabLocationTop)

	var split *container.Split

	presetTitle := widget.NewLabelWithStyle("Research Presets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	minimizePresets := widget.NewButtonWithIcon("Minimize", theme.ContentRemoveIcon(), func() {
		if split == nil {
			return
		}
		split.SetOffset(presetCollapsedOffset)
	})
	minimizePresets.Importance = widget.LowImportance

	maximizePresets := widget.NewButtonWithIcon("Maximize", theme.ViewFullScreenIcon(), func() {
		if split == nil {
			return
		}
		if split.Offset >= presetExpandedOffset-0.01 {
			split.SetOffset(presetDefaultOffset)
			return
		}
		split.SetOffset(presetExpandedOffset)
	})
	maximizePresets.Importance = widget.LowImportance

	presetHeader := container.NewHBox(presetTitle, layout.NewSpacer(), minimizePresets, maximizePresets)

	presetSidebar := container.NewVBox(
		presetHeader,
		presetInfo,
		presetNameEntry,
		container.NewGridWithColumns(1, savePresetButton),
		widget.NewSeparator(),
		container.NewBorder(nil, container.NewHBox(deletePresetButton, layout.NewSpacer()), nil, nil, presetListScroll),
	)

	form := widget.NewForm(
		widget.NewFormItem("Seed Keyword", keywordEntry),
		widget.NewFormItem("Marketplace", countrySelect),
		widget.NewFormItem("Format", formatSelect),
	)
	form.SubmitText = "Fetch Suggestions"
	form.OnSubmit = fetchKeywords

	mainContent := container.NewVBox(
		widget.NewLabelWithStyle("Keyword Research", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewPadded(searchBar),
		widget.NewSeparator(),
		container.NewPadded(keywordFilters),
		container.NewPadded(bestsellerFilters),
		widget.NewSeparator(),
		actionGrid,
		widget.NewSeparator(),
		keywordOutputs,
	)

	presetScroll := container.NewVScroll(container.NewPadded(presetSidebar))
	presetScroll.SetMinSize(fyne.NewSize(260, 0))

	mainScroll := container.NewVScroll(container.NewPadded(mainContent))
	mainScroll.SetMinSize(fyne.NewSize(0, 0))

	split = container.NewHSplit(presetScroll, mainScroll)
	split.SetOffset(presetDefaultOffset)

	return split
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
	maxCompetitionEntry.SetPlaceHolder("5")
	maxDensityEntry := widget.NewEntry()
	maxDensityEntry.SetPlaceHolder("100")

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Book title or product headline")
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Paste your blurb or key selling points here…")

	headlineTemplateOptions := []struct {
		Name    string
		Content string
	}{
		{Name: "Bestseller Boost", Content: "Boost Your Book Sales With Proven Keyword Targeting"},
		{Name: "Reader Hook", Content: "Hook New Readers With High-Intent Amazon Searches"},
		{Name: "Launch Momentum", Content: "Launch Day Momentum For Your Latest Release"},
	}
	descriptionTemplateOptions := []struct {
		Name    string
		Content string
	}{
		{Name: "Data-Backed Pitch", Content: "Target engaged shoppers with keyword clusters proven to convert."},
		{Name: "Benefit Driven", Content: "Show readers the transformation your book delivers in the first line."},
		{Name: "Social Proof", Content: "Highlight reviews and credibility boosters to earn instant trust."},
	}

	headlineChoices := []string{"Custom"}
	headlineLookup := map[string]string{}
	for _, option := range headlineTemplateOptions {
		headlineChoices = append(headlineChoices, option.Name)
		headlineLookup[option.Name] = option.Content
	}
	headlinePreview := widget.NewLabel("Toggle a template to prefill your headline.")
	headlinePreview.Wrapping = fyne.TextWrapWord
	headlineToggle := widget.NewRadioGroup(headlineChoices, func(selected string) {
		if selected == "" || selected == "Custom" {
			headlinePreview.SetText("Toggle a template to prefill your headline.")
			return
		}
		content := headlineLookup[selected]
		titleEntry.SetText(content)
		headlinePreview.SetText(content)
	})
	headlineToggle.Horizontal = true
	headlineToggle.SetSelected("Custom")

	descriptionChoices := []string{"Custom"}
	descriptionLookup := map[string]string{}
	for _, option := range descriptionTemplateOptions {
		descriptionChoices = append(descriptionChoices, option.Name)
		descriptionLookup[option.Name] = option.Content
	}
	descriptionPreview := widget.NewLabel("Activate a description template to spark ideas.")
	descriptionPreview.Wrapping = fyne.TextWrapWord
	descriptionToggle := widget.NewRadioGroup(descriptionChoices, func(selected string) {
		if selected == "" || selected == "Custom" {
			descriptionPreview.SetText("Activate a description template to spark ideas.")
			return
		}
		content := descriptionLookup[selected]
		descriptionEntry.SetText(content)
		descriptionPreview.SetText(content)
	})
	descriptionToggle.Horizontal = true
	descriptionToggle.SetSelected("Custom")

	templateCard := widget.NewCard("Ad Copy Templates", "Toggle a preset to instantly populate your copy fields.", container.NewVBox(
		widget.NewLabelWithStyle("Headline Ideas", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		headlineToggle,
		headlinePreview,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Description Ideas", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		descriptionToggle,
		descriptionPreview,
	))
	competitorEntry := widget.NewMultiLineEntry()
	competitorEntry.SetPlaceHolder("Comma separated competitor keywords or ASIN phrases")
	importCSVButton := widget.NewButtonWithIcon("Import CSV", theme.FolderOpenIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(fmt.Errorf("unable to read CSV: %w", err), window)
				return
			}
			keywords, err := parseCSVKeywords(string(data))
			if err != nil {
				dialog.ShowError(fmt.Errorf("unable to parse CSV: %w", err), window)
				return
			}
			competitorEntry.SetText(strings.Join(keywords, "\n"))
		}, window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fileDialog.Show()
	})

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

	compliancePanel := container.NewVBox()
	compliancePanel.Hide()

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

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Reverse ASIN", fmt.Sprintf("Investigating %s…", strings.ToUpper(asin)), cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			insights, err := service.ReverseASINSearch(ctx, asin, country, filters)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to run reverse ASIN: %v", err)
					}
					safeSet(reverseResult, message)
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

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "Campaign Builder", "Composing Amazon Ads keyword list…", cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			keywords, err := service.GenerateAMSKeywords(ctx, titleEntry.Text, descriptionEntry.Text, competitors, country)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to generate campaign keywords: %v", err)
					}
					safeSet(campaignResult, message)
					lastCampaignKeywords = nil
					campaignControls.Hide()
					updateCompliancePanel(compliancePanel, nil)
					return
				}
				lastCampaignKeywords = keywords
				campaignControls.Show()
				campaignControls.Refresh()
				flagged := scraper.FlagIllegalKeywords(keywords)
				updateCompliancePanel(compliancePanel, flagged)
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

	competitorInput := container.NewVBox(
		competitorEntry,
		container.NewHBox(importCSVButton, layout.NewSpacer()),
	)

	campaignSection := container.NewVBox(
		widget.NewLabelWithStyle("Amazon Ads Planner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		titleEntry,
		descriptionEntry,
		templateCard,
		competitorInput,
		campaignButton,
		widget.NewSeparator(),
		campaignControls,
		compliancePanel,
		newResultScroll(container.NewPadded(campaignLabel)),
	)

	sidebar := widget.NewForm(
		widget.NewFormItem("Marketplace", countrySelect),
	)

	columns := container.NewAdaptiveGrid(2,
		container.NewPadded(reverseSection),
		container.NewPadded(campaignSection),
	)

	return container.NewBorder(sidebar, nil, nil, nil, columns)
}

func buildInternationalTab(window fyne.Window, service *scraper.Service, countries []string, result binding.String, activity *serviceActivity, quota *quotaTracker) fyne.CanvasObject {
	keywordEntry := widget.NewEntry()
	keywordEntry.SetPlaceHolder("mindfulness journal")

	countryGroup := newRegionalCheckGroup(countries)
	defaults := defaultInternationalSelection(countries)
	if len(defaults) > 0 {
		countryGroup.SetSelected(defaults)
	}

	resultLabel := widget.NewLabelWithData(result)
	resultLabel.Wrapping = fyne.TextWrapWord

	summaryLabel := widget.NewLabel("")
	summaryLabel.Hide()

	var (
		lastInternational []scraper.InternationalKeyword
		tableData         []scraper.InternationalKeyword
		sortColumn        = internationalSortByCountry
		sortAscending     = true
	)

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

	statusContainer := container.NewPadded(resultLabel)

	table := widget.NewTable(
		func() (int, int) {
			rows := len(tableData)
			if rows == 0 {
				return 1, 3
			}
			return rows + 1, 3
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignLeading
			label.Wrapping = fyne.TextWrapOff
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			label, _ := object.(*widget.Label)
			if label == nil {
				return
			}
			label.Wrapping = fyne.TextWrapOff
			label.TextStyle = fyne.TextStyle{}
			label.Alignment = fyne.TextAlignLeading

			if id.Row == 0 {
				headers := []string{"Country", "Keyword", "Volume"}
				header := headers[id.Col]
				if sortColumn == internationalSortColumn(id.Col) {
					if sortAscending {
						header = fmt.Sprintf("%s ↑", header)
					} else {
						header = fmt.Sprintf("%s ↓", header)
					}
				}
				label.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col == 2 {
					label.Alignment = fyne.TextAlignTrailing
				}
				label.SetText(header)
				return
			}

			if id.Row-1 >= len(tableData) {
				label.SetText("")
				return
			}

			keyword := tableData[id.Row-1]
			switch id.Col {
			case 0:
				label.SetText(internationalCountryLabel(keyword))
			case 1:
				label.SetText(keyword.Keyword)
			case 2:
				label.Alignment = fyne.TextAlignTrailing
				label.SetText(fmt.Sprintf("%d", keyword.SearchVolume))
			}
		},
	)
	table.SetColumnWidth(0, 190)
	table.SetColumnWidth(1, 260)
	table.SetColumnWidth(2, 110)

	tableContainer := container.NewPadded(table)
	tableContainer.Hide()

	updateResults := func(keywords []scraper.InternationalKeyword) {
		tableData = append(tableData[:0], keywords...)
		sortInternationalKeywords(tableData, sortColumn, sortAscending)
		lastInternational = append(lastInternational[:0], tableData...)

		if len(tableData) == 0 {
			table.Refresh()
			summaryLabel.Hide()
			tableContainer.Hide()
			statusContainer.Show()
			return
		}

		table.Refresh()
		markets := countInternationalMarkets(tableData)
		summaryLabel.SetText(fmt.Sprintf("%d localised keywords across %d markets", len(tableData), markets))
		summaryLabel.Show()
		tableContainer.Show()
		statusContainer.Hide()
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			newColumn := internationalSortColumn(id.Col)
			if sortColumn == newColumn {
				sortAscending = !sortAscending
			} else {
				sortColumn = newColumn
				if newColumn == internationalSortByVolume {
					sortAscending = false
				} else {
					sortAscending = true
				}
			}
			if len(tableData) > 0 {
				sortInternationalKeywords(tableData, sortColumn, sortAscending)
				table.Refresh()
				lastInternational = append(lastInternational[:0], tableData...)
			}
		}
		table.Unselect(id)
	}

	fetch := func() {
		keyword := strings.TrimSpace(keywordEntry.Text)
		if keyword == "" {
			dialog.ShowInformation("International Research", "Enter a seed keyword to continue.", window)
			return
		}

		selected := countryGroup.Selected()
		if len(selected) == 0 {
			dialog.ShowInformation("International Research", "Select at least one marketplace to analyse.", window)
			return
		}

		resultControls.Hide()
		summaryLabel.Hide()
		tableContainer.Hide()
		statusContainer.Show()
		safeSet(result, "Localising your keyword list…")
		tableData = tableData[:0]
		table.Refresh()

		activity.Start()
		quota.Use()

		ctx, cancel := context.WithTimeout(context.Background(), currentRequestTimeout())
		progress := newCancelableProgress(window, "International Research", "Localising your keyword list…", cancel)
		if progress != nil {
			progress.Show()
		}

		go func() {
			defer cancel()

			keywords, err := service.InternationalKeywords(ctx, keyword, selected)

			queueOnMain(window, func() {
				activity.Done()
				if progress != nil {
					progress.Hide()
				}
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						dialog.ShowError(err, window)
					}
					message := "Request cancelled."
					if !errors.Is(err, context.Canceled) {
						message = fmt.Sprintf("Unable to fetch international keywords: %v", err)
					}
					safeSet(result, message)
					lastInternational = nil
					resultControls.Hide()
					updateResults(nil)
					return
				}
				updateResults(keywords)
				if len(lastInternational) > 0 {
					resultControls.Show()
					resultControls.Refresh()
					safeSet(result, "Tap table headers to sort by country, keyword or volume.")
				} else {
					safeSet(result, "No international opportunities found yet. Try selecting more marketplaces.")
				}
			})
		}()
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("International Keyword Expansion", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keywordEntry,
		countryGroup.Content(),
		widget.NewButton("Generate Suggestions", fetch),
		widget.NewSeparator(),
		resultControls,
		summaryLabel,
		newResultScroll(container.NewStack(statusContainer, tableContainer)),
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
	densityRow := ""
	if count := formatCount(details.TitleDensity); count != "" {
		densityRow = fmt.Sprintf("Title Density: %s titles", count)
	}
	pricingContent := container.NewVBox(buildRows(
		fmt.Sprintf("Price: %s", fallback(currencyPrice, "Not available")),
		fmt.Sprintf("Rating: %s", fallback(details.Rating, "Unknown")),
		fmt.Sprintf("Reviews: %s", fallback(details.ReviewCount, "Unknown")),
		densityRow,
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

func keywordCompetitionBadge(score float64) string {
	scaled := int(math.Round(score * 10))
	if scaled < 0 {
		scaled = 0
	}
	if scaled > 100 {
		scaled = 100
	}

	switch {
	case scaled <= 30:
		return fmt.Sprintf("🟢 %d", scaled)
	case scaled <= 60:
		return fmt.Sprintf("🟡 %d", scaled)
	default:
		return fmt.Sprintf("🔴 %d", scaled)
	}
}

func keywordDensityBadge(density float64) string {
	value := int(math.Round(density))
	if value < 0 {
		value = 0
	}
	if value > 10 {
		value = 10
	}

	switch {
	case value <= 3:
		return fmt.Sprintf("🟢 %d", value)
	case value <= 6:
		return fmt.Sprintf("🟡 %d", value)
	default:
		return fmt.Sprintf("🔴 %d", value)
	}
}

func keywordSummaryMessage(total int) string {
	if total <= 0 {
		return "No keyword suggestions available yet."
	}
	if total == 1 {
		return "1 keyword suggestion ready."
	}
	return fmt.Sprintf("%d keyword suggestions ready.", total)
}

func sortKeywordInsights(insights []scraper.KeywordInsight) {
	sort.SliceStable(insights, func(i, j int) bool {
		if insights[i].SearchVolume == insights[j].SearchVolume {
			return strings.ToLower(insights[i].Keyword) < strings.ToLower(insights[j].Keyword)
		}
		return insights[i].SearchVolume > insights[j].SearchVolume
	})
}

func newInfoButton(window fyne.Window, tooltip string) *widget.Button {
	button := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		if win := resolveWindow(window); win != nil {
			dialog.NewInformation("Details", tooltip, win).Show()
		}
	})
	button.Importance = widget.LowImportance
	return button
}

func resolveWindow(window fyne.Window) fyne.Window {
	if window != nil {
		return window
	}
	if app := fyne.CurrentApp(); app != nil {
		if drv := app.Driver(); drv != nil {
			for _, win := range drv.AllWindows() {
				if win != nil {
					return win
				}
			}
		}
	}
	return nil
}

func showThrottleSettingsDialog(window fyne.Window) {
	win := resolveWindow(window)
	if win == nil {
		return
	}

	current := throttleState.Snapshot()

	info := widget.NewLabel("Adjust how aggressively RankBeam sends requests. Higher values speed up research but increase the risk of rate limits.")
	info.Wrapping = fyne.TextWrapWord

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(strconv.Itoa(int(current.Timeout / time.Second)))

	concurrencyEntry := widget.NewEntry()
	concurrencyEntry.SetText(strconv.Itoa(current.RequestsPerMinute))

	errorLabel := widget.NewLabel("")
	errorLabel.Hide()
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.TextStyle = fyne.TextStyle{Italic: true}
	errorLabel.Importance = widget.DangerImportance

	form := widget.NewForm(
		widget.NewFormItem("", info),
		widget.NewFormItem("Request timeout (seconds)", timeoutEntry),
		widget.NewFormItem("Requests per minute", concurrencyEntry),
	)

	validateInputs := func() (int, int, error) {
		timeoutText := strings.TrimSpace(timeoutEntry.Text)
		if timeoutText == "" {
			return 0, 0, errors.New("request timeout is required")
		}
		timeoutSeconds, err := strconv.Atoi(timeoutText)
		if err != nil || timeoutSeconds <= 0 {
			return 0, 0, errors.New("enter a positive timeout in seconds")
		}

		rpmText := strings.TrimSpace(concurrencyEntry.Text)
		if rpmText == "" {
			return 0, 0, errors.New("requests per minute is required")
		}
		requestsPerMinute, err := strconv.Atoi(rpmText)
		if err != nil || requestsPerMinute <= 0 {
			return 0, 0, errors.New("enter a positive number of requests per minute")
		}

		return timeoutSeconds, requestsPerMinute, nil
	}

	content := container.NewVBox(
		form,
		errorLabel,
	)

	var modal dialog.Dialog
	modal = dialog.NewCustomConfirm("Request Settings", "Save", "Cancel", content, func(ok bool) {
		if !ok {
			errorLabel.Hide()
			if modal != nil {
				modal.Hide()
			}
			return
		}

		timeoutSeconds, requestsPerMinute, err := validateInputs()
		if err != nil {
			errorLabel.SetText(err.Error())
			errorLabel.Show()
			if modal != nil {
				modal.Show()
			}
			return
		}

		errorLabel.Hide()
		if modal != nil {
			modal.Hide()
		}
		_, changed := throttleState.Update(time.Duration(timeoutSeconds)*time.Second, requestsPerMinute)
		if changed {
			loadMainApplication(win)
		}
	}, win)
	modal.Resize(fyne.NewSize(420, 0))
	modal.Show()
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
		formatCount(details.TitleDensity),
		boolToString(details.IsIndependent),
		fallback(details.URL, ""),
		strings.Join(ranks, " | "),
	}

	exportCSV(window, "product.csv", []string{
		"ASIN", "Title", "Price", "Currency", "Rating", "ReviewCount", "Availability", "Brand", "Publisher", "PublicationDate", "PrintLength", "Dimensions", "Language", "TitleDensity", "Independent", "URL", "BestsellerRanks",
	}, [][]string{row})
}

func exportSearchCSV(window fyne.Window, results []scraper.SearchResult) {
	if len(results) == 0 {
		return
	}

	rows := make([][]string, 0, len(results))
	for _, result := range results {
		rows = append(rows, []string{
			strconv.Itoa(result.Rank),
			result.ASIN,
			result.Title,
			result.Author,
			result.Price,
			result.Rating,
			result.ReviewCount,
			result.BestSellerRank,
			result.URL,
		})
	}

	exportCSV(window, "catalog-search.csv", []string{"Rank", "ASIN", "Title", "Author", "Price", "Rating", "Reviews", "BestSellerRank", "URL"}, rows)
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
			formatCount(insight.CompetitionScore),
			formatFloat(insight.RelevancyScore),
			formatCount(insight.TitleDensity),
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

func formatCount(value float64) string {
	if value < 0 {
		return ""
	}
	return fmt.Sprintf("%d", int(math.Round(value)))
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

func formatSearchResults(results []scraper.SearchResult) string {
	if len(results) == 0 {
		return "No catalog listings returned. Try another keyword."
	}

	builder := &strings.Builder{}
	for _, result := range results {
		fmt.Fprintf(builder, "#%d %s (%s)\n", result.Rank, fallback(result.Title, "Untitled"), strings.ToUpper(fallback(result.ASIN, "N/A")))
		if result.Author != "" {
			fmt.Fprintf(builder, "   Author: %s\n", result.Author)
		}
		if result.Price != "" {
			fmt.Fprintf(builder, "   Price: %s\n", result.Price)
		}
		if result.Rating != "" || result.ReviewCount != "" {
			fmt.Fprintf(builder, "   Rating: %s (%s reviews)\n", fallback(result.Rating, "N/A"), fallback(result.ReviewCount, "0"))
		}
		if result.BestSellerRank != "" {
			fmt.Fprintf(builder, "   Rank: %s\n", result.BestSellerRank)
		}
		if result.URL != "" {
			fmt.Fprintf(builder, "   URL: %s\n", result.URL)
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
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
		if count := formatCount(details.TitleDensity); count != "" {
			fmt.Fprintf(builder, "Title Density: %s titles\n", count)
		}
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
		fmt.Fprintf(builder, "   Competition: %s titles | Relevancy: %.2f | Title Density: %s titles\n\n", formatCount(insight.CompetitionScore), insight.RelevancyScore, formatCount(insight.TitleDensity))
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
		if count := formatCount(product.TitleDensity); count != "" {
			fmt.Fprintf(builder, "   Title Density: %s titles\n", count)
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

func updateCompliancePanel(panel *fyne.Container, keywords []string) {
	if panel == nil {
		return
	}

	panel.Objects = nil
	if len(keywords) == 0 {
		panel.Hide()
		panel.Refresh()
		return
	}

	badges := make([]fyne.CanvasObject, 0, len(keywords))
	for _, keyword := range keywords {
		if strings.TrimSpace(keyword) == "" {
			continue
		}
		badges = append(badges, newComplianceBadge(keyword))
	}

	if len(badges) == 0 {
		panel.Hide()
		panel.Refresh()
		return
	}

	warning := widget.NewLabel("Amazon Ads policies may reject the highlighted keywords. Remove or refine them before launching campaigns.")
	warning.Wrapping = fyne.TextWrapWord

	grid := container.NewAdaptiveGrid(3, badges...)
	content := container.NewVBox(warning, grid)

	panel.Objects = []fyne.CanvasObject{
		widget.NewCard("Compliance Alerts", "Review policy risks before launching.", content),
	}
	panel.Show()
	panel.Refresh()
}

func newComplianceBadge(keyword string) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(strings.TrimSpace(keyword), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label.Wrapping = fyne.TextWrapWord

	chipContent := container.NewHBox(layout.NewSpacer(), label, layout.NewSpacer())
	padded := container.NewPadded(chipContent)

	background := canvas.NewRectangle(themeColor(theme.ColorNameError, theme.VariantLight))
	background.CornerRadius = 8
	background.StrokeColor = themeColor(theme.ColorNameError, theme.VariantDark)
	background.StrokeWidth = 1

	return container.NewMax(background, padded)
}

func themeColor(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if app := fyne.CurrentApp(); app != nil {
		if th := app.Settings().Theme(); th != nil {
			return th.Color(name, variant)
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

type internationalSortColumn int

const (
	internationalSortByCountry internationalSortColumn = iota
	internationalSortByKeyword
	internationalSortByVolume
)

func internationalCountryLabel(keyword scraper.InternationalKeyword) string {
	name := strings.TrimSpace(keyword.CountryName)
	code := strings.ToUpper(strings.TrimSpace(keyword.CountryCode))
	if name == "" {
		name = code
	}
	if code == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, code)
}

func countInternationalMarkets(keywords []scraper.InternationalKeyword) int {
	seen := make(map[string]struct{})
	for _, keyword := range keywords {
		code := strings.ToUpper(strings.TrimSpace(keyword.CountryCode))
		if code == "" {
			code = strings.TrimSpace(keyword.CountryName)
		}
		if code == "" {
			continue
		}
		seen[code] = struct{}{}
	}
	return len(seen)
}

func sortInternationalKeywords(keywords []scraper.InternationalKeyword, column internationalSortColumn, ascending bool) {
	sort.SliceStable(keywords, func(i, j int) bool {
		left := keywords[i]
		right := keywords[j]

		compareCountry := func() int {
			return strings.Compare(strings.ToLower(internationalCountryLabel(left)), strings.ToLower(internationalCountryLabel(right)))
		}

		compareKeyword := func() int {
			return strings.Compare(strings.ToLower(left.Keyword), strings.ToLower(right.Keyword))
		}

		compareVolume := func() int {
			switch {
			case left.SearchVolume < right.SearchVolume:
				return -1
			case left.SearchVolume > right.SearchVolume:
				return 1
			default:
				return 0
			}
		}

		result := 0
		switch column {
		case internationalSortByCountry:
			result = compareCountry()
		case internationalSortByKeyword:
			result = compareKeyword()
		case internationalSortByVolume:
			result = compareVolume()
		}

		if result == 0 {
			// Secondary sort keeps output stable and groups related entries.
			if column != internationalSortByCountry {
				if country := compareCountry(); country != 0 {
					result = country
				}
			}
			if result == 0 {
				result = -compareVolume()
			}
			if result == 0 {
				result = compareKeyword()
			}
		}

		if ascending {
			return result < 0
		}
		return result > 0
	})
}

func formatInternationalKeywords(keywords []scraper.InternationalKeyword) string {
	if len(keywords) == 0 {
		return "No international opportunities found yet. Try selecting more marketplaces."
	}

	sorted := append([]scraper.InternationalKeyword(nil), keywords...)
	sortInternationalKeywords(sorted, internationalSortByCountry, true)

	headerCountry := "Country"
	headerKeyword := "Keyword"
	headerVolume := "Volume"
	countryWidth := len(headerCountry)
	keywordWidth := len(headerKeyword)
	volumeWidth := len(headerVolume)

	rows := make([][3]string, len(sorted))
	for index, keyword := range sorted {
		country := internationalCountryLabel(keyword)
		rows[index] = [3]string{country, keyword.Keyword, fmt.Sprintf("%d", keyword.SearchVolume)}
		if len(country) > countryWidth {
			countryWidth = len(country)
		}
		if len(keyword.Keyword) > keywordWidth {
			keywordWidth = len(keyword.Keyword)
		}
		if len(rows[index][2]) > volumeWidth {
			volumeWidth = len(rows[index][2])
		}
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "%-*s  %-*s  %*s\n", countryWidth, headerCountry, keywordWidth, headerKeyword, volumeWidth, headerVolume)
	fmt.Fprintf(builder, "%s  %s  %s\n", strings.Repeat("-", countryWidth), strings.Repeat("-", keywordWidth), strings.Repeat("-", volumeWidth))
	for _, row := range rows {
		fmt.Fprintf(builder, "%-*s  %-*s  %*s\n", countryWidth, row[0], keywordWidth, row[1], volumeWidth, row[2])
	}

	return strings.TrimSpace(builder.String())
}

type regionalCheckGroup struct {
	container *fyne.Container
	groups    []*widget.CheckGroup
	onChanged func([]string)
}

func newRegionalCheckGroup(countries []string) *regionalCheckGroup {
	grouped := map[string][]string{}
	for _, code := range countries {
		region := marketplaceRegion(code)
		grouped[region] = append(grouped[region], code)
	}

	orderedRegions := []string{"North America", "Europe", "Asia Pacific", "Middle East", "South America", "Other Markets"}
	objects := make([]fyne.CanvasObject, 0)
	groups := make([]*widget.CheckGroup, 0)

	addRegion := func(region string, codes []string) {
		if len(codes) == 0 {
			return
		}
		sort.SliceStable(codes, func(i, j int) bool { return strings.ToUpper(codes[i]) < strings.ToUpper(codes[j]) })
		header := widget.NewLabelWithStyle(region, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		options := make([]string, len(codes))
		copy(options, codes)
		check := widget.NewCheckGroup(options, nil)
		check.Horizontal = true
		groups = append(groups, check)
		section := container.NewVBox(header, check)
		objects = append(objects, section)
	}

	for _, region := range orderedRegions {
		addRegion(region, grouped[region])
		delete(grouped, region)
	}

	if len(grouped) > 0 {
		leftover := make([]string, 0, len(grouped))
		for region := range grouped {
			leftover = append(leftover, region)
		}
		sort.Strings(leftover)
		for _, region := range leftover {
			addRegion(region, grouped[region])
		}
	}

	if len(objects) == 0 {
		fallback := widget.NewCheckGroup(countries, nil)
		return &regionalCheckGroup{
			container: container.NewVBox(fallback),
			groups:    []*widget.CheckGroup{fallback},
		}
	}

	groupContainer := container.NewVBox(objects...)
	rg := &regionalCheckGroup{
		container: groupContainer,
		groups:    groups,
	}
	for _, group := range rg.groups {
		current := group
		current.OnChanged = func(_ []string) {
			if rg.onChanged != nil {
				rg.onChanged(rg.Selected())
			}
		}
	}
	return rg
}

func (r *regionalCheckGroup) Content() fyne.CanvasObject {
	if r == nil {
		return widget.NewLabel("No marketplaces available.")
	}
	return r.container
}

func (r *regionalCheckGroup) Selected() []string {
	if r == nil {
		return nil
	}
	combined := make([]string, 0)
	for _, group := range r.groups {
		combined = append(combined, group.Selected...)
	}
	sort.SliceStable(combined, func(i, j int) bool { return strings.ToUpper(combined[i]) < strings.ToUpper(combined[j]) })
	return combined
}

func (r *regionalCheckGroup) SetSelected(values []string) {
	if r == nil {
		return
	}
	desired := map[string]struct{}{}
	for _, value := range values {
		desired[strings.ToUpper(value)] = struct{}{}
	}
	for _, group := range r.groups {
		if group == nil {
			continue
		}
		matching := make([]string, 0)
		for _, option := range group.Options {
			if _, ok := desired[strings.ToUpper(option)]; ok {
				matching = append(matching, option)
			}
		}
		group.SetSelected(matching)
	}
}

func (r *regionalCheckGroup) OnChanged(fn func([]string)) {
	if r == nil {
		return
	}
	r.onChanged = fn
}

func marketplaceRegion(code string) string {
	switch strings.ToUpper(code) {
	case "US", "CA", "MX":
		return "North America"
	case "BR":
		return "South America"
	case "UK", "GB", "IE", "DE", "FR", "ES", "IT", "NL", "SE", "PL", "TR", "BE", "CH", "AT":
		return "Europe"
	case "AU", "NZ", "JP", "IN", "SG", "CN", "KR":
		return "Asia Pacific"
	case "AE", "SA", "EG", "QA":
		return "Middle East"
	default:
		return "Other Markets"
	}
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
			return filter, fmt.Errorf("invalid maximum competition count: %w", err)
		}
		if value < 0 {
			return filter, errors.New("maximum competition count cannot be negative")
		}
		filter.MaxCompetitionScore = value
	}

	if trimmed := strings.TrimSpace(maxDensity); trimmed != "" {
		value, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return filter, fmt.Errorf("invalid maximum title density count: %w", err)
		}
		if value < 0 {
			return filter, errors.New("maximum title density count cannot be negative")
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

type cancelableProgress struct {
	dialog         dialog.Dialog
	message        *widget.Label
	cancelBtn      *widget.Button
	cancel         context.CancelFunc
	cancelOnce     sync.Once
	initialMessage string
}

func newCancelableProgress(window fyne.Window, title, message string, cancel context.CancelFunc) *cancelableProgress {
	if window == nil {
		return nil
	}

	label := widget.NewLabel(message)
	bar := widget.NewProgressBarInfinite()
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), nil)

	progress := &cancelableProgress{message: label, cancelBtn: cancelBtn, cancel: cancel, initialMessage: message}
	cancelBtn.OnTapped = func() {
		progress.cancelOnce.Do(func() {
			label.SetText("Cancelling…")
			cancelBtn.Disable()
			if progress.cancel != nil {
				progress.cancel()
			}
		})
	}

	content := container.NewVBox(label, bar, cancelBtn)
	progress.dialog = dialog.NewCustomWithoutButtons(title, content, window)
	return progress
}

func (p *cancelableProgress) Show() {
	if p == nil || p.dialog == nil {
		return
	}
	p.cancelOnce = sync.Once{}
	p.cancelBtn.Enable()
	if p.message != nil {
		p.message.SetText(p.initialMessage)
	}
	p.dialog.Show()
}

func (p *cancelableProgress) Hide() {
	if p == nil || p.dialog == nil {
		return
	}
	p.dialog.Hide()
}

func listKeywordPresetNames() []string {
	keywordPresetLock.Lock()
	defer keywordPresetLock.Unlock()

	names := make([]string, len(keywordPresets))
	for i, preset := range keywordPresets {
		if strings.TrimSpace(preset.Name) == "" {
			names[i] = fmt.Sprintf("Preset %d", i+1)
			continue
		}
		names[i] = preset.Name
	}
	return names
}

func saveKeywordPreset(p keywordPreset) {
	keywordPresetLock.Lock()
	defer keywordPresetLock.Unlock()

	// Replace preset with same name if it exists to keep sidebar tidy.
	for i, existing := range keywordPresets {
		if strings.EqualFold(strings.TrimSpace(existing.Name), strings.TrimSpace(p.Name)) && strings.TrimSpace(p.Name) != "" {
			keywordPresets[i] = p
			return
		}
	}
	keywordPresets = append(keywordPresets, p)
}

func getKeywordPreset(index int) (keywordPreset, bool) {
	keywordPresetLock.Lock()
	defer keywordPresetLock.Unlock()

	if index < 0 || index >= len(keywordPresets) {
		return keywordPreset{}, false
	}
	return keywordPresets[index], true
}

func deleteKeywordPreset(index int) bool {
	keywordPresetLock.Lock()
	defer keywordPresetLock.Unlock()

	if index < 0 || index >= len(keywordPresets) {
		return false
	}
	keywordPresets = append(keywordPresets[:index], keywordPresets[index+1:]...)
	return true
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
