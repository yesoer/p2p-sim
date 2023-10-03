package gui

import (
	"distributed-sys-emulator/backend"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ControlBar struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*ControlBar)(nil)

func NewControlBar(network backend.Network) ControlBar {
	execution := container.NewHBox(
		widget.NewButton("Start", func() {
			network.Emit(backend.START)
		}),
		widget.NewButton("Stop", func() {
			network.Emit(backend.STOP)
		}),
	)

	return ControlBar{execution}
}

func (c ControlBar) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}
