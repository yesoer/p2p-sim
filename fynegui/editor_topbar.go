package fynegui

import (
	"github.com/yesoer/p2p-sim/bus"
	"github.com/yesoer/p2p-sim/embed"
	"github.com/yesoer/p2p-sim/log"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

// Declare conformance with the Component interface
var _ Component = (*EditorTopbar)(nil)

type EditorTopbar struct {
	fyne.CanvasObject
	stateMu sync.Mutex

	// state data
	projectSrc bus.Source
}

func NewEditorTopbar(eb bus.EventBus, window fyne.Window) *EditorTopbar {
	editorTop := &EditorTopbar{}
	editorTop.stateMu = sync.Mutex{}
	editorTop.projectSrc = bus.Source{
		Path: "./",
		Type: bus.Directory,
	}

	wcanvas := window.Canvas()

	settingsBtn := editorTop.settings(eb, wcanvas)
	openProjectBtn := editorTop.openProject(eb, window)
	examplesExplorerBtn := editorTop.examplesExplorer(eb, wcanvas)
	fileExplorerBtn := editorTop.fileExplorer(eb, wcanvas)

	editorTop.CanvasObject = container.NewHBox(
		settingsBtn,
		examplesExplorerBtn,
		openProjectBtn,
		fileExplorerBtn,
	)

	// Show/hide the file explorer depending on the editor selected
	eb.Bind(bus.EditorSelectEvt, func(editorType bus.EditorType) {
		switch editorType {
		case bus.Neovim:
			fileExplorerBtn.Hide()
		case bus.Entry:
			fileExplorerBtn.Show()
		}
	})

	return editorTop
}

func (e *EditorTopbar) settings(eb bus.EventBus, wcanvas fyne.Canvas) *widget.Button {
	// radio group to select an editor
	opts := []string{
		string(bus.Neovim),
		string(bus.Entry),
	}

	editorSelect := widget.NewRadioGroup(opts, func(value string) {
		e := bus.Event{Type: bus.EditorSelectEvt, Data: bus.EditorType(value)}
		eb.Publish(e)
	})
	editorSelect.Horizontal = false
	editorSelect.Selected = string(bus.Entry)

	// a modal to wrap the various settings
	settings := container.NewVBox(editorSelect)
	settingsModal := NewModal(settings, wcanvas)

	settingsIcon := theme.SettingsIcon()
	settingsBtn := widget.NewButtonWithIcon("", settingsIcon, func() {
		settingsModal.Resize(fyne.NewSize(300, 300))
		settingsModal.Show()
	})

	return settingsBtn
}

// A button to open a file explorer to open a project from the local filesystem
func (e *EditorTopbar) openProject(eb bus.EventBus, window fyne.Window) *widget.Button {
	showOpenProjectDialog := func(w fyne.Window) {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Error(err)
				return
			}
			if uri == nil {
				return
			}

			pth := uri.Path()
			data := bus.Source{
				Path: pth,
				Type: bus.Directory,
			}
			evt := bus.Event{Type: bus.OpenEvt, Data: data}
			eb.Publish(evt)

			e.projectSrc = data
		}, w)
	}

	folderOpenIcon := theme.FolderOpenIcon()
	openProjectBtn := widget.NewButtonWithIcon("", folderOpenIcon, func() {
		showOpenProjectDialog(window)
	})

	return openProjectBtn
}

// Open a file from the current project
func (e *EditorTopbar) fileExplorer(eb bus.EventBus, wcanvas fyne.Canvas) *widget.Button {
	var explorerModal Modal

	// system file explorer
	tree := xwidget.NewFileTree(storage.NewFileURI(e.projectSrc.Path))
	tree.OnSelected = func(rawUrl string) {
		pth := strings.Replace(rawUrl, "file://", "", 1)
		relativePath := bus.Source{
			Path: pth,
			Type: bus.File,
		}
		e := bus.Event{Type: bus.OpenEvt, Data: relativePath}
		eb.Publish(e)

		explorerModal.Hide()
	}

	explorerModal = NewModal(tree, wcanvas)
	fileIcon := theme.FileIcon()
	btn := widget.NewButtonWithIcon("", fileIcon, func() {
		explorerModal.Resize(fyne.NewSize(300, 300))
		explorerModal.Show()
	})

	// update according to the current project
	eb.Bind(bus.OpenEvt, func(source bus.Source) {
		if source.Type != bus.Directory {
			return
		}

		tree.Show()

		// parse to uri
		pthUri := storage.NewFileURI(source.Path)
		tree.Root = pthUri.String()
		tree.Refresh()
	})

	return btn
}

// Explore the example projects embedded in the binary
func (e *EditorTopbar) examplesExplorer(eb bus.EventBus, wcanvas fyne.Canvas) *widget.Button {
	var examplesModal Modal
	examples := embed.GetExampleList()

	// create a button for each example
	var examplesBtns []fyne.CanvasObject
	for _, example := range examples {
		btn := widget.NewButton(example.Name, func() {
			source := bus.Source{
				Path: example.Path,
				Type: bus.Directory,
			}
			evt := bus.Event{Type: bus.OpenEvt, Data: source}
			eb.Publish(evt)

			e.projectSrc = source

			examplesModal.Hide()
		})
		examplesBtns = append(examplesBtns, btn)
	}

	// create a modal to show the buttons for example selection
	examplesContainer := container.NewVScroll(container.NewVBox(examplesBtns...))
	examplesModal = NewModal(examplesContainer, wcanvas)
	searchIcon := theme.SearchIcon()
	examplesBtn := widget.NewButtonWithIcon("", searchIcon, func() {
		examplesModal.Resize(fyne.NewSize(300, 300))
		examplesModal.Show()
	})

	return examplesBtn
}

// GetCanvasObj implements the Component Interface
func (e *EditorTopbar) GetCanvasObj() fyne.CanvasObject {
	return e.CanvasObject
}
