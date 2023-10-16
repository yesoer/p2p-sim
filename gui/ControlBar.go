package gui

import (
	"distributed-sys-emulator/bus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ControlBar struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*ControlBar)(nil)

func NewControlBar(eb *bus.EventBus) ControlBar {
	execution := container.NewHBox(
		widget.NewButton("Start", func() {
			e := bus.Event{Type: bus.StartEvt, Data: nil}
			eb.Publish(e)
		}),
		widget.NewButton("Stop", func() {
			e := bus.Event{Type: bus.StopEvt, Data: nil}
			eb.Publish(e)
		}),
	)

	return ControlBar{execution}
}

func (c ControlBar) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}
