package fynegui

// NOTE : extended (text-)editor from github.com/fyne-io/defyne/

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/core"
	"distributed-sys-emulator/log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the Component interface
var _ Component = (*Editor)(nil)

type Editor struct {
	*widget.Entry
	path string
}

func NewTextEditor(path string, _ fyne.Window, eb bus.EventBus) *Editor {
	input := widget.NewMultiLineEntry()
	input.TextStyle.Monospace = true
	input.Wrapping = fyne.TextTruncate
	input.PlaceHolder = "Type"

	// read code from disk
	b, err := os.ReadFile(path)
	if err == nil {
		input.SetText(string(b))
	}

	// publish code to core
	code := core.Code(b)
	e := bus.Event{Type: bus.CodeChangeEvt, Data: code}
	eb.Publish(e)

	// make sure to send editor changes to core
	changeCB := func(e *Editor) {
		text := e.Content()
		code := core.Code(text)
		evt := bus.Event{Type: bus.CodeChangeEvt, Data: code}
		eb.Publish(evt)
	}

	editor := Editor{input, path}
	editor.OnChanged = func(_ string) {
		changeCB(&editor)
	}

	// process changes from the file explorer
	eb.Bind(bus.FileOpenEvt, func(path bus.FilePath) {
		b, err := os.ReadFile(string(path))
		if err == nil {
			input.SetText(string(b))
			editor.path = string(path)
			return
		}
		log.Error(err)
	})

	return &editor
}

func (e *Editor) GetCanvasObj() fyne.CanvasObject {
	return e.Entry
}

func (e *Editor) Content() string {
	return e.Text
}

func (e *Editor) Save() {
	err := os.WriteFile(e.path, []byte(e.Text), 0644)
	if err != nil {
		log.Error(err)
	}
}
