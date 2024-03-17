package fynegui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Modal struct {
	*widget.PopUp
}

func NewModal(content fyne.CanvasObject, canvas fyne.Canvas) Modal {
	var popup *widget.PopUp
	closeBtn := widget.NewButton("X", func() {
		popup.Hide()
	})
	topBar := container.NewHBox(closeBtn, layout.NewSpacer())
	wrapper := container.NewBorder(topBar, nil, nil, nil, content)
	popup = widget.NewModalPopUp(wrapper, canvas)

	return Modal{popup}
}
