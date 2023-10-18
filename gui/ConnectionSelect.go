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

func NewConnectionsSelect(eb *bus.EventBus) ConnectionsSelect {
	// keep node count up to date
	var nodeCnt int
	eb.Bind(bus.NetworkNodeCntChangeEvt, func(e bus.Event) {
		newNodeCnt := e.Data.(int)
		nodeCnt = newNodeCnt
		// TODO : refresh
	})

	// create a checkbox grid to manage nodes/connections
	connections := container.NewGridWithColumns(nodeCnt)

	for row := 0; row < nodeCnt; row++ {
		for col := 0; col < nodeCnt; col++ {
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
