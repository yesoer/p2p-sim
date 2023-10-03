package gui

import (
	"distributed-sys-emulator/backend"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
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
	workingDir := "."
	path := workingDir + "/code.go"
	// TODO : could probably move this to editor.go
	onSubmitted := func(e *Editor) {
		text := e.Content()
		code := backend.Code(text)
		network.SetCode(code)
	}
	editor := NewTextEditor(path, window, onSubmitted)

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

	// save edited file
	// TODO : trigger on editor, not canvas/else
	saveSC := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(saveSC, func(shortcut fyne.Shortcut) {
		fmt.Println("Save triggered")
		editor.Save()
	})

	// Layout : resizable middle split with the editor left and everything else
	// on the right
	view := container.NewBorder(execution.GetCanvasObj(), nil, nil, nil, canvasRaster.GetCanvasObj())
	split := container.NewHSplit(editor.GetCanvasObj(), view)

	window.SetContent(split)
	window.ShowAndRun()
}
