package gui

import (
	"distributed-sys-emulator/backend"
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

func RunGUI(eb *bus.EventBus) {

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

	// TODO : could move this to editor.go
	onSubmitted := func(e *Editor) {
		text := e.Content()
		code := backend.Code(text)
		evt := bus.Event{Type: bus.CodeChangedEvt, Data: code}
		eb.Publish(evt)
	}
	editor := NewTextEditor(path, window, onSubmitted, eb)

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

	// Layout : resizable middle split with the editor left and everything else
	// on the right
	view := container.NewBorder(execution.GetCanvasObj(), nil, nil, nil, canvasRaster.GetCanvasObj())
	split := container.NewHSplit(editor.GetCanvasObj(), view)

	window.SetContent(split)
	window.ShowAndRun()
}
