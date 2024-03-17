package fynegui

import (
	"distributed-sys-emulator/bus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the Component interface
var _ Component = (*ConnectionsSelect)(nil)

type ConnectionsSelect struct {
	*fyne.Container
}

func NewConnectionsSelect(eb bus.EventBus) *ConnectionsSelect {
	// keep node count up to date
	var nodeCnt int
	var connections bus.Connections
	connectionsWrap := container.NewHBox()

	addCheckboxes := func() {
		connectionsWrap.RemoveAll()

		// create a checkbox grid to manage nodes/connections
		grid := container.NewGridWithColumns(nodeCnt)

		for row := 0; row < nodeCnt; row++ {
			for col := 0; col < nodeCnt; col++ {
				ccol, crow := col, row // copy for closure
				checkbox := widget.NewCheck("", nil)

				// depending on current connections from core, set checkmarks
				for _, c := range connections {
					if c.From == row && c.To == col {
						checkbox.SetChecked(true)
					}
				}

				checkbox.OnChanged = func(b bool) {
					data := bus.Connection{
						From: crow,
						To:   ccol,
					}

					// connect the two nodes
					if b {
						e := bus.Event{Type: bus.ConnectNodesEvt, Data: data}
						eb.Publish(e)
					} else {
						e := bus.Event{Type: bus.DisconnectNodesEvt, Data: data}
						eb.Publish(e)
					}
				}

				// nodes cannot have connection to self
				if row == col {
					checkbox.Disable()
				}

				grid.Add(checkbox)
			}
		}

		connectionsWrap.Add(grid)
	}

	eb.Bind(bus.NetworkResizeEvt, func(resizeData bus.NetworkResize) {
		connections = resizeData.Connections
		nodeCnt = resizeData.Cnt

		// refresh
		addCheckboxes()
		connectionsWrap.Refresh()
	})

	return &ConnectionsSelect{connectionsWrap}
}

func (c ConnectionsSelect) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}
