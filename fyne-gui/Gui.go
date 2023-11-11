package fynegui

import (
	"distributed-sys-emulator/bus"
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

var InitialWindowSize = fyne.NewSize(1000, 800)

func RunGUI(eb bus.EventBus) {

	// basics
	a := app.New()
	window := a.NewWindow("Distributed System Emulator")
	window.SetMaster()
	window.Resize(InitialWindowSize)
	window.CenterOnScreen()

	//-------------------------------------------------------
	// CREATE COMPONENTS

	// canvas
	canvasRaster := NewCanvas(eb, window.Canvas())

	// connections
	connections := NewConnectionsSelect(eb)

	// create a pane to control execution
	execution := NewControlBar(eb)

	// create an editor for the nodes behaviour
	workingDir := "."
	path := workingDir + "/code.go"

	editor := NewTextEditor(path, window, eb)

	console := NewConsole(eb)

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
	// TODO : can we get 'command' to work ?
	saveSC := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(saveSC, func(shortcut fyne.Shortcut) {
		fmt.Println("Save triggered")
		editor.Save()
	})

	// Layout : resizable middle split with the editor left, the output console
	// below it and everything else on the right
	view := container.NewBorder(execution.GetCanvasObj(), nil, nil, nil, canvasRaster.GetCanvasObj())
	devenv := container.NewBorder(nil, console.GetCanvasObj(), nil, nil, editor.GetCanvasObj())
	split := container.NewHSplit(devenv, view)

	window.SetContent(split)
	window.ShowAndRun()
}
