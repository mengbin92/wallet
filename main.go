package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	a := app.New()
	w := a.NewWindow("Crypto Tool UI")

	tabs := container.NewAppTabs(
		NewWalletTab(w),
		NewCollectTab(w),
	)

	w.SetContent(tabs)
	w.Resize(DefaultWindowSize)
	w.ShowAndRun()
}
