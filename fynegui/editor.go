package fynegui

import (
	"errors"
	"github.com/yesoer/p2p-sim/bus"
	"github.com/yesoer/p2p-sim/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Declare conformance with the Component interface
var _ Component = (*editor)(nil)

type editor struct {
	c       *fyne.Container
	current bus.EditorType
}

const defaultEditor = bus.Entry

type EditorType interface {
	fyne.Tappable
	fyne.Focusable
	fyne.Widget
	fyne.Shortcutable
}

// Wraps the different editors and switches them as needed
// TODO : when switching editors/recreating them, they must unbind from the bus
// maybe add a destroy method to the editor/component interface ? Or keep them alive ?
func NewEditor(window fyne.Window, eb bus.EventBus) *editor {
	src := bus.Source{
		Path: ".",
		Type: bus.Directory,
	}

	var ed EditorType
	c := container.NewBorder(nil, nil, nil, nil, ed)
	e := &editor{c, defaultEditor}
	e.tryReset(src, eb, window)

	eb.Bind(bus.OpenEvt, func(newSource bus.Source) {
		if newSource.Type != bus.Directory {
			return
		}

		src = newSource
		e.tryReset(newSource, eb, window)
	})

	eb.Bind(bus.EditorSelectEvt, func(editorType bus.EditorType) {
		switch editorType {
		case bus.Neovim:
			e.current = bus.Neovim
		case bus.Entry:
			e.current = bus.Entry
		}

		e.tryReset(src, eb, window)
	})

	return e
}

// Try to create a new instance of the editor with a new path. If it fails,
// nothing happens
func (e *editor) tryReset(source bus.Source, eb bus.EventBus, window fyne.Window) {
	e.c.RemoveAll()

	var component EditorType
	switch e.current {
	case bus.Neovim:
		component = NewNvim(source.Path)
	case bus.Entry:
		component = NewCodeEntry(source.Path, eb)
	default:
		err := errors.New("Invalid editor type")
		log.Error(err)
		return
	}

	e.c.Add(component)
	e.c.Refresh()
	window.Canvas().Focus(component)
	return
}

func (e *editor) GetCanvasObj() fyne.CanvasObject {
	return e.c
}
