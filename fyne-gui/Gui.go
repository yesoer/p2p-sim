package fynegui

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"embed"
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

// embed code examples
//
//
//go:embed resources/*.go
var content embed.FS

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

	// system file explorer
	saveIcon := theme.DocumentSaveIcon()
	basePath := "./"
	tree := xwidget.NewFileTree(storage.NewFileURI(basePath))
	tree.OnSelected = func(rawUrl string) {
		pth := strings.Replace(rawUrl, "file://", "", 1)
		relativePath := bus.File{Path: path.Join(basePath, pth), Source: bus.Local}
		e := bus.Event{Type: bus.FileOpenEvt, Data: relativePath}
		eb.Publish(e)
	}

	explorerModal := NewModal(tree, wcanvas)
	btn := widget.NewButtonWithIcon("", saveIcon, func() {
		explorerModal.Resize(fyne.NewSize(300, 300))
		explorerModal.Show()
	})

	// embedded examples explorer
	// create a list of files from the embedded filesystem
	embeddedFiles, err := content.ReadDir("resources")
	if err != nil {
		log.Error(err)
}

	// create a list of buttons from the embedded files which open them
	var examples []fyne.CanvasObject
	for _, f := range embeddedFiles {
		if !f.IsDir() {
			btn := widget.NewButton(f.Name(), func() {
				filepath := bus.File{Path: f.Name(), Source: bus.Embed}
				e := bus.Event{Type: bus.FileOpenEvt, Data: filepath}
				eb.Publish(e)
			})
			examples = append(examples, btn)
		}
	}

	// create a modal for the buttons
	examplesContainer := container.NewVScroll(container.NewVBox(examples...))
	examplesModal := NewModal(examplesContainer, wcanvas)
	examplesBtn := widget.NewButton("Examples", func() {
		examplesModal.Resize(fyne.NewSize(300, 300))
		examplesModal.Show()
	})

	// top bar of the editor
	editorTop := container.NewHBox(btn, examplesBtn)

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
