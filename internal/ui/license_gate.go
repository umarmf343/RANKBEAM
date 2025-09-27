package ui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"net/url"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/umarmf343/Umar-kdp-product-api/internal/license"
)

func enforceLicense() (*license.Client, string, string) {
	client, err := license.NewClientFromEnv()
	if err != nil {
		if errors.Is(err, license.ErrMissingBaseURL) {
			return nil, "", "Set LICENSE_API_URL to your license server endpoint before launching the app."
		}
		return nil, "", fmt.Sprintf("Unable to create license client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	key, err := license.ValidateLocalLicense(ctx, client)
	if err != nil {
		return client, "", licenseErrorMessage(err)
	}

	return client, key, ""
}

func licenseErrorMessage(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "License key not found. Please run the installer to activate this machine."
	case errors.Is(err, license.ErrEmptyLicenseKey):
		return "The stored license key is empty. Re-run the installer or paste a valid key."
	case errors.Is(err, license.ErrInvalidLicense):
		return "The license key on this machine is invalid or expired. Contact support to refresh it."
	case errors.Is(err, license.ErrUnauthorizedToken):
		return "The installer token configured for this app is not authorized. Check LICENSE_API_TOKEN."
	default:
		return fmt.Sprintf("Unable to validate license: %v", err)
	}
}

func renderLicenseFailure(window fyne.Window, client *license.Client, message string) {
	window.SetTitle("RankBeam — Activate License")

	heroTitle := canvas.NewText("Activate Your Competitive Edge", theme.PrimaryColor())
	heroTitle.TextStyle = fyne.TextStyle{Bold: true}
	heroTitle.TextSize = 26
	heroTitle.Alignment = fyne.TextAlignLeading

	emotionalHook := widget.NewRichTextFromMarkdown("**Your rivals are watching every stale listing and every unguarded keyword.** Activate now to weaponize precision data before they do.")
	emotionalHook.Wrapping = fyne.TextWrapWord

	featureHighlights := widget.NewRichTextFromMarkdown("- Laser-guided keyword vaults for disruptive launches\n- Predator-level competitor heatmaps revealing weak spots\n- Campaign blueprints engineered to hijack buyer intent")
	featureHighlights.Wrapping = fyne.TextWrapWord

	warning := widget.NewRichTextFromMarkdown(fmt.Sprintf("⚠️ %s", message))
	warning.Wrapping = fyne.TextWrapWord
	if len(warning.Segments) > 0 {
		if segment, ok := warning.Segments[0].(*widget.TextSegment); ok {
			segment.Style = widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}, ColorName: theme.ColorNameError}
		}
	}

	ctaURL, _ := url.Parse("https://rankbeam.hannyshive.com.ng/")
	ctaButton := widget.NewButtonWithIcon("Get Your License", theme.MailComposeIcon(), func() {
		if fyne.CurrentApp() != nil {
			if err := fyne.CurrentApp().OpenURL(ctaURL); err != nil {
				dialog.ShowError(err, window)
			}
		}
	})
	ctaButton.Importance = widget.HighImportance

	licenseEntry := widget.NewMultiLineEntry()
	licenseEntry.SetPlaceHolder("Paste your license key here…")
	licenseEntry.Wrapping = fyne.TextWrapWord
	licenseEntry.SetMinRowsVisible(6)

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	submitButton := widget.NewButtonWithIcon("Activate & Launch", theme.ConfirmIcon(), nil)
	submitButton.Importance = widget.HighImportance

	var activationInProgress bool

	updateSubmitButtonState := func() {
		if activationInProgress || strings.TrimSpace(licenseEntry.Text) == "" {
			submitButton.Disable()
			return
		}
		submitButton.Enable()
	}

	licenseEntry.OnChanged = func(string) {
		updateSubmitButtonState()
	}

	updateSubmitButtonState()

	submitButton.OnTapped = func() {
		if client == nil {
			dialog.ShowInformation("Activation Unavailable", "Configure LICENSE_API_URL and LICENSE_API_TOKEN before attempting activation.", window)
			return
		}

		key := strings.TrimSpace(licenseEntry.Text)
		if key == "" {
			dialog.ShowInformation("License Activation", "Paste a valid license key before continuing.", window)
			return
		}

		activationInProgress = true
		updateSubmitButtonState()
		statusLabel.SetText("Validating your license with the command server…")

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			fingerprint, err := license.Fingerprint()
			if err == nil {
				err = client.ValidateLicense(ctx, key, fingerprint)
			}
			if err == nil {
				_, err = license.SaveLicenseKey(key)
			}

			queueOnMain(window, func() {
				if err != nil {
					activationInProgress = false
					updateSubmitButtonState()
					statusLabel.SetText(activationErrorMessage(err))
					if errors.Is(err, license.ErrInvalidLicense) {
						if fyne.CurrentApp() != nil {
							if openErr := fyne.CurrentApp().OpenURL(ctaURL); openErr != nil {
								dialog.ShowError(openErr, window)
							}
						}
					}
					return
				}
				statusLabel.SetText("License activated. Summoning the intelligence suite…")
				loadMainApplication(window, key)
			})
		}()
	}

	formHeader := widget.NewLabelWithStyle("Activate License", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	formHeader.Wrapping = fyne.TextWrapWord

	form := container.NewVBox(
		formHeader,
		widget.NewSeparator(),
		licenseEntry,
		container.NewHBox(layout.NewSpacer(), submitButton),
		statusLabel,
	)

	marketingPanel := container.NewVBox(
		heroTitle,
		emotionalHook,
		widget.NewSeparator(),
		featureHighlights,
		widget.NewSeparator(),
		warning,
		container.NewHBox(ctaButton, layout.NewSpacer()),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("© 2024 RankBeam Intelligence Labs. Unauthorized replication triggers encrypted watermark tracing, dark-ops analytics, and immediate legal pursuit."),
	)

	heroBackground := canvas.NewRectangle(color.NRGBA{R: 241, G: 245, B: 255, A: 255})

	marketingCard := container.NewAdaptiveGrid(1,
		container.NewMax(heroBackground, container.NewPadded(marketingPanel)),
	)

	window.SetContent(container.NewPadded(marketingCard))
}

func activationErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "Activation timed out. Check your connection and try again."
	case errors.Is(err, license.ErrInvalidLicense),
		errors.Is(err, license.ErrEmptyLicenseKey),
		errors.Is(err, license.ErrUnauthorizedToken):
		return licenseErrorMessage(err)
	case errors.Is(err, os.ErrPermission):
		return "Activation failed: insufficient permissions to store the license key on this device."
	default:
		return fmt.Sprintf("Activation failed: %v", err)
	}
}

func summarizeKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "…" + key[len(key)-4:]
}
