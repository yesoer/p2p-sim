package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func NewModal(content fyne.CanvasObject, canvas fyne.Canvas) *widget.PopUp {
	var modal *widget.PopUp
	closeBtn := widget.NewButton("X", func() {
		modal.Hide()
	})
	topBar := container.NewHBox(closeBtn, layout.NewSpacer())
	wrapper := container.NewVBox(topBar, content)
	modal = widget.NewModalPopUp(wrapper, canvas)
	return modal
}
