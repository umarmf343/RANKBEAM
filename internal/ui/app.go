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

	window := application.NewWindow("Amazon Product Intelligence Suite")
	window.Resize(fyne.NewSize(1024, 720))
	window.SetMaster()

	licenseKey, licenseError := enforceLicense()
	if licenseError != "" {
		renderLicenseFailure(window, licenseError)
		window.ShowAndRun()
		return
	}

	title := "Amazon Product Intelligence Suite"
	if licenseKey != "" {
		title = fmt.Sprintf("%s — License %s", title, summarizeKey(licenseKey))
	}
	window.SetTitle(title)

	service := scraper.NewService(25*time.Second, 25)
	defer service.Close()

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
