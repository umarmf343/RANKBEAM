package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app.png
var appScreenshot []byte

// AppScreenshot returns an embedded preview of the desktop interface.
func AppScreenshot() fyne.Resource {
	if len(appScreenshot) == 0 {
		return nil
	}
	return fyne.NewStaticResource("app.png", appScreenshot)
}
