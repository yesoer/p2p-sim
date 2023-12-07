package fynegui

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
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
	canvasRaster := NewNetworkDiagram(eb, window.Canvas())

	// connections
	connections := NewConnectionsSelect(eb)

	// create a pane to control execution
	execution := NewControlBar(eb)

	// create an editor for the nodes behaviour
	workingDir := "."
	pth := workingDir + "/code.go"

	editor := NewTextEditor(pth, window, eb)

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

	// explorer
	saveIcon := theme.DocumentSaveIcon()
	basePath := "./"
	tree := xwidget.NewFileTree(storage.NewFileURI(basePath))
	tree.OnSelected = func(rawUrl string) {
		pth := strings.Replace(rawUrl, "file://", "", 1)
		relativePath := bus.FilePath(path.Join(basePath, pth))
		e := bus.Event{Type: bus.FileOpenEvt, Data: relativePath}
		eb.Publish(e)
	}

	content := NewModal(tree, wcanvas)
	btn := widget.NewButtonWithIcon("", saveIcon, func() {
		content.Resize(fyne.NewSize(300, 300))
		content.Show()
	})
	editorTop := container.NewHBox(btn)

	// save edited file
	// TODO : trigger on editor, not canvas/else
	// TODO : can we get 'command' to work ?
	saveSC := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(saveSC, func(shortcut fyne.Shortcut) {
		log.Debug("Stored to disk")
		editor.Save()
	})

	// Layout : resizable middle split with the editor left, the output console
	// below it and everything else on the right
	view := container.NewBorder(execution.GetCanvasObj(), nil, nil, nil, canvasRaster)
	devenv := container.NewBorder(editorTop, console.GetCanvasObj(), nil, nil, editor.GetCanvasObj())
	split := container.NewHSplit(devenv, view)

	window.SetContent(split)
	window.ShowAndRun()
}
