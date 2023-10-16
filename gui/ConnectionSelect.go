package gui

import (
	"distributed-sys-emulator/bus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ConnectionsSelect struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*ConnectionsSelect)(nil)

func NewConnectionsSelect(eb *bus.EventBus, nodesCnt int) ConnectionsSelect {
	// create a checkbox grid to manage nodes/connections
	connections := container.NewGridWithColumns(nodesCnt)

	for row := 0; row < nodesCnt; row++ {
		for col := 0; col < nodesCnt; col++ {
			ccol, crow := col, row // copy for closure
			checkboxHandler := func(b bool) {
				data := bus.CheckboxPos{
					Ccol: ccol,
					Crow: crow,
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
			checkbox := widget.NewCheck("", checkboxHandler)

			// nodes cannot have connection to self
			if row == col {
				checkbox.Disable()
			}

			connections.Add(checkbox)
		}
	}

	return ConnectionsSelect{connections}
}

func (c ConnectionsSelect) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}
