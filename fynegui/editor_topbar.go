package fynegui

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

func NewEditorTopbar(eb bus.EventBus, window fyne.Window) fyne.CanvasObject {

	wcanvas := window.Canvas()

	openProjectBtn := openProject(eb, window)
	fileExplorerBtn := fileExplorer(eb, wcanvas)
	examplesExplorerBtn := examplesExplorer(eb, wcanvas)

	// top bar of the editor
	editorTop := container.NewHBox(
		openProjectBtn,
		fileExplorerBtn,
		examplesExplorerBtn,
	)

	return editorTop
}

// A button to open a file explorer to open a project
func openProject(eb bus.EventBus, window fyne.Window) *widget.Button {
	showOpenProjectDialog := func(w fyne.Window) {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if err != nil {
				log.Error(err)
				return
			}
			if u == nil {
				return
			}

			pth := u.Path()
			e := bus.Event{Type: bus.ProjectOpenEvt, Data: pth}
			eb.Publish(e)
		}, w)
	}

	openProjectBtn := widget.NewButton("Open Project", func() {
		showOpenProjectDialog(window)
	})

	return openProjectBtn
}

func fileExplorer(eb bus.EventBus, wcanvas fyne.Canvas) *widget.Button {
	// system file explorer
	saveIcon := theme.DocumentSaveIcon()
	basePath := "./"
	tree := xwidget.NewFileTree(storage.NewFileURI(basePath))
	tree.OnSelected = func(rawUrl string) {
		pth := strings.Replace(rawUrl, "file://", "", 1)
		relativePath := bus.File{Path: path.Join(basePath, pth), Source: bus.LocalFile}
		e := bus.Event{Type: bus.FileOpenEvt, Data: relativePath}
		eb.Publish(e)
	}

	explorerModal := NewModal(tree, wcanvas)
	btn := widget.NewButtonWithIcon("", saveIcon, func() {
		explorerModal.Resize(fyne.NewSize(300, 300))
		explorerModal.Show()
	})

	return btn
}

func examplesExplorer(eb bus.EventBus, wcanvas fyne.Canvas) *widget.Button {
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
				filepath := bus.File{Path: f.Name(), Source: bus.EmbeddedFile}
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

	return examplesBtn
}
