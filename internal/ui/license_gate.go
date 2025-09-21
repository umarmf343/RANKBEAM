package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/umar/amazon-product-scraper/internal/license"
)

func enforceLicense() (string, string) {
	client, err := license.NewClientFromEnv()
	if err != nil {
		if errors.Is(err, license.ErrMissingBaseURL) {
			return "", "Set LICENSE_API_URL to your license server endpoint before launching the app."
		}
		return "", fmt.Sprintf("Unable to create license client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	key, err := license.ValidateLocalLicense(ctx, client)
	if err != nil {
		return "", licenseErrorMessage(err)
	}

	return key, ""
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

func renderLicenseFailure(window fyne.Window, message string) {
	title := widget.NewLabelWithStyle("License Activation Required", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapWord

	detail := widget.NewLabel(message)
	detail.Wrapping = fyne.TextWrapWord

	exitButton := widget.NewButton("Exit", func() {
		fyne.CurrentApp().Quit()
	})

	content := container.NewVBox(title, widget.NewSeparator(), detail, layout.NewSpacer(), exitButton)
	window.SetContent(container.NewPadded(content))
}

func summarizeKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "…" + key[len(key)-4:]
}
