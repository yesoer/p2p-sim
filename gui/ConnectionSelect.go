package gui

import (
	"distributed-sys-emulator/backend"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ConnectionsSelect struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*ConnectionsSelect)(nil)

func NewConnectionsSelect(network backend.Network, canvasRaster Canvas) ConnectionsSelect {
	// create a checkbox grid to manage nodes/connections
	nodesCnt := network.GetNodeCnt()
	connections := container.NewGridWithColumns(nodesCnt)

	for row := 0; row < nodesCnt; row++ {
		for col := 0; col < nodesCnt; col++ {
			ccol, crow := col, row // copy for closure
			checkboxHandler := func(b bool) {
				// connect the two nodes
				if b {
					network.ConnectNodes(crow, ccol)
				} else {
					network.DisconnectNodes(crow, ccol)
				}

				canvasRaster.Refresh()
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
