package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/umar/amazon-product-scraper/internal/licenseclient"
)

func newLicenseGate(win fyne.Window, client *licenseclient.Client, storage *licenseclient.Storage, onSuccess func()) fyne.CanvasObject {
	fingerprint, err := licenseclient.Fingerprint()
	statusBinding := binding.NewString()
	if err != nil {
		statusBinding.Set(fmt.Sprintf("Unable to fingerprint this machine: %v", err))
		return container.NewVBox(widget.NewLabelWithData(statusBinding))
	}

	instructions := widget.NewLabel("This copy of Amazon Product Intelligence Suite requires activation. Enter the license key displayed during installation or emailed after purchase.")
	instructions.Wrapping = fyne.TextWrapWord

	statusBinding.Set("Awaiting license key...")

	licenseEntry := widget.NewEntry()
	licenseEntry.SetPlaceHolder("ABCDE-FGHIJ-KLMNO-PQRST-UVWXY")

	statusLabel := widget.NewLabelWithData(statusBinding)
	statusLabel.Wrapping = fyne.TextWrapWord

	progress := widget.NewProgressBarInfinite()
	progress.Stop()
	progress.Hide()

	activate := widget.NewButton("Activate", nil)
	activate.Disable()

	quit := widget.NewButton("Quit", func() {
		win.Close()
	})

	var mu sync.Mutex
	active := false

	runValidation := func(key string, showError bool) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}

		mu.Lock()
		if active {
			mu.Unlock()
			return
		}
		active = true
		mu.Unlock()

		fyne.CurrentApp().Driver().RunOnMain(func() {
			progress.Show()
			progress.Start()
			activate.Disable()
		})
		statusBinding.Set("Contacting license server...")

		go func() {
			defer func() {
				mu.Lock()
				active = false
				mu.Unlock()
				fyne.CurrentApp().Driver().RunOnMain(func() {
					progress.Stop()
					progress.Hide()
					if strings.TrimSpace(licenseEntry.Text) != "" {
						activate.Enable()
					}
				})
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			envelope, err := client.Validate(ctx, key, fingerprint)
			if err != nil {
				statusBinding.Set(fmt.Sprintf("Activation failed: %v", err))
				if showError {
					fyne.CurrentApp().Driver().RunOnMain(func() {
						dialog.ShowError(err, win)
					})
				}
				return
			}

			if envelope.LicenseKey == "" {
				envelope.LicenseKey = key
			}

			if err := storage.Save(envelope); err != nil {
				statusBinding.Set(fmt.Sprintf("Saved license but persistence failed: %v", err))
				fyne.CurrentApp().Driver().RunOnMain(func() {
					dialog.ShowError(err, win)
				})
				return
			}

			statusBinding.Set("License validated successfully. Loading application...")
			fyne.CurrentApp().Driver().RunOnMain(onSuccess)
		}()
	}

	activate.OnTapped = func() {
		runValidation(licenseEntry.Text, true)
	}

	licenseEntry.OnChanged = func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			activate.Disable()
		} else if !active {
			activate.Enable()
		}
	}

	// Attempt automatic validation using a stored license key.
	if stored, err := storage.Load(); err == nil && strings.TrimSpace(stored.LicenseKey) != "" {
		licenseEntry.SetText(stored.LicenseKey)
		statusBinding.Set("Validating stored license...")
		runValidation(stored.LicenseKey, false)
	}

	content := container.NewVBox(
		instructions,
		widget.NewForm(widget.NewFormItem("License key", licenseEntry)),
		container.NewHBox(activate, quit),
		progress,
		statusLabel,
		widget.NewLabel(fmt.Sprintf("Machine fingerprint: %s", fingerprint)),
	)

	return container.NewCenter(container.NewVBox(content))
}
