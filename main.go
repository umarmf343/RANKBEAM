package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/umarmf343/Umar-kdp-product-api/internal/ui"
)

func main() {
	// Create a new Fyne app
	myApp := app.New()

	// Load the app icon (adjust the path if needed)
	iconFile := "assets/app.png" // Adjust if your icon is located elsewhere
	icon, err := fyne.LoadResourceFromPath(iconFile)
	if err != nil {
		panic("Icon not found: " + err.Error())
	}

	// Create the main window
	myWindow := myApp.NewWindow("RankBeam")
	myWindow.SetIcon(icon) // Set the app icon

	// Run your original UI logic
	ui.Run()

	// Show the window and run the app
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.SetContent(widget.NewLabel("Hello, RankBeam!"))
	myWindow.ShowAndRun()
}
