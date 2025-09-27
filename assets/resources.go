package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app.png
var appIcon []byte

// AppIcon returns the embedded application icon resource.
func AppIcon() fyne.Resource {
	if len(appIcon) == 0 {
		return nil
	}
	return fyne.NewStaticResource("app.png", appIcon)
}

// AppScreenshot returns an embedded preview of the desktop interface.
// It currently reuses the application icon as a placeholder for marketing
// screens that expect a screenshot resource.
func AppScreenshot() fyne.Resource {
	return AppIcon()
}
