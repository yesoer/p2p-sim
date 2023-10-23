package gui

import (
	"distributed-sys-emulator/bus"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Console struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*Console)(nil)

func NewConsole(eb *bus.EventBus) *Console {
	c := container.NewBorder(nil, nil, nil, nil)
	cmu := sync.Mutex{}

	var nodeCnt int
	var outputs []string

	// depending on the outputs, create console boxes
	refresh := func() {
		cmu.Lock()
		c.RemoveAll()
		grid := container.NewGridWithColumns(nodeCnt)
		for i := 0; i < nodeCnt; i++ {
			entry := widget.NewLabel(outputs[i])
			grid.Add(entry)
		}
		c.Add(grid)
		c.Refresh()
		cmu.Unlock()
	}

	// update node count and output slice size as required
	eb.Bind(bus.NetworkNodeCntChangeEvt, func(e bus.Event) {
		nodeCnt = e.Data.(int)
		if nodeCnt > len(outputs) {
			for i := 0; i <= nodeCnt-len(outputs); i++ {
				outputs = append(outputs, "")
			}
		}
		refresh()
	})

	// update outputs
	eb.Bind(bus.OutputChanged, func(e bus.Event) {
		out := e.Data.(bus.Output)
		outputs[out.NodeId] = out.Str
		refresh()
	})

	console := &Console{c}
	return console
}

func (e *Console) GetCanvasObj() fyne.CanvasObject {
	return e.Container
}
