package gui

import (
	"distributed-sys-emulator/backend"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Component interface {
	GetCanvasObj() fyne.CanvasObject
}

func RunGUI(network backend.Network, size fyne.Size) {
	// basics
	a := app.New()
	window := a.NewWindow("Distributed System Emulator")
	window.SetMaster()
	window.Resize(size)
	window.CenterOnScreen()

	//-------------------------------------------------------
	// CREATE COMPONENTS

	// canvas
	canvasRaster := NewCanvas(network, window.Canvas())

	// connections
	connections := NewConnectionsSelect(network, canvasRaster)

	// create a pane to control execution
	execution := NewControlBar(network)

	// create an editor for the nodes behaviour
	workingDir := "~/Documents/GitHub/P2PSim"
	path := workingDir + "/code.go"

	// TODO : move this to editor.go ?
	onSubmitted := func(e *Editor) {
		code, err := e.GetContent()
		if err == nil {
			network.SetCode(code)
		}
		fmt.Println("error : ", err)
	}

	editor := NewEditor(path, window, onSubmitted)

	code, err := editor.GetContent()
	if err == nil {
		network.SetCode(code)
	}

	//-------------------------------------------------------
	// EMBED COMPONENTS IN LAYOUT

	// Popup
	connectionsCanvasObj := connections.GetCanvasObj()
	wcanvas := window.Canvas()
	connectionTab := NewModal(connectionsCanvasObj, wcanvas)
	connect := widget.NewButton("Connect", func() {
		connectionTab.Show()
	})
	execution.Add(connect)

	// Layout : resizable middle split with the editor left and everything else
	// on the right
	view := container.NewBorder(execution.GetCanvasObj(), nil, nil, nil, canvasRaster.GetCanvasObj())
	split := container.NewHSplit(editor.GetCanvasObj(), view)

	window.SetContent(split)
	window.ShowAndRun()
}
