package fynegui

import (
	"distributed-sys-emulator/bus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the Component interface
var _ Component = (*Console)(nil)

type Console struct {
	*fyne.Container
}

func NewConsole(eb bus.EventBus) *Console {
	c := container.NewBorder(nil, nil, nil, nil)

	var nodeCnt int
	var outputs []string

	// depending on the outputs, create console boxes
	refresh := func() {
		c.RemoveAll()
		grid := container.NewGridWithColumns(nodeCnt)
		for i := 0; i < len(outputs); i++ {
			entry := widget.NewLabel(outputs[i])
			grid.Add(entry)
		}
		c.Add(grid)
		c.Refresh()
	}

	// update node count and output slice size as required
	eb.Bind(bus.NetworkResizeEvt, func(resizeData bus.NetworkResize) {
		nodeCnt = resizeData.Cnt
		if nodeCnt > len(outputs) {
			for i := 0; i <= nodeCnt-len(outputs); i++ {
				outputs = append(outputs, "")
			}
		}
		refresh()
	})

	// update outputs
	eb.Bind(bus.NodeOutputLogEvt, func(out bus.NodeLog) {
		outputs[out.NodeId] = out.Str
		refresh()
	})

	console := &Console{c}
	return console
}

func (e *Console) GetCanvasObj() fyne.CanvasObject {
	return e.Container
}
