package fynegui

import (
	"github.com/yesoer/p2p-sim/bus"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

	// right pane canvas
	canvasRaster := NewNetworkDiagram(eb, window.Canvas())

	// right pane top bar
	connectionsSelect := NewConnectionsSelect(eb)
	connectionsCanvasObj := connectionsSelect.GetCanvasObj()
	wcanvas := window.Canvas()
	connectionTab := NewModal(connectionsCanvasObj, wcanvas)
	connect := widget.NewButton("Connect", func() {
		connectionTab.Show()
	})

	canvasTop := NewControlBar(eb)
	canvasTop.Add(connect)

	// left pane editor
	editor := NewEditor(window, eb)

	// left pane top bar
	editorTop := NewEditorTopbar(eb, window)

	// left pane bottom console
	console := NewConsole(eb)

	//-------------------------------------------------------
	// EMBED COMPONENTS IN LAYOUT
	// Layout : resizable middle split with the editor left, the output console
	// below it and everything else on the right
	rightPane := container.NewBorder(canvasTop.GetCanvasObj(), nil, nil, nil, canvasRaster)
	leftPane := container.NewBorder(
		editorTop.GetCanvasObj(),
		console.GetCanvasObj(),
		nil,
		nil,
		editor.GetCanvasObj(),
	)
	split := container.NewHSplit(leftPane, rightPane)

	window.SetContent(split)
	window.ShowAndRun()
}
