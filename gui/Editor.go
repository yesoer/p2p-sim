package gui

// NOTE : extended (text-)editor from github.com/fyne-io/defyne/

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Editor struct {
	*widget.Entry
	path   string
	edited bool
}

// Declare conformance with the Component interface
var _ Component = (*Editor)(nil)

func NewTextEditor(path string, _ fyne.Window, changeCB func(e *Editor)) *Editor {
	input := widget.NewMultiLineEntry()
	input.TextStyle.Monospace = true
	input.Wrapping = fyne.TextTruncate
	input.PlaceHolder = "Type"

	b, err := os.ReadFile(path)
	if err == nil {
		input.SetText(string(b))
	}

	editor := &Editor{input, path, false}
	editor.OnChanged = func(_ string) {
		changeCB(editor)
		editor.edited = true
	}
	return editor
}

func (e *Editor) GetCanvasObj() fyne.CanvasObject {
	return e.Entry
}

func (e *Editor) Changed() bool {
	return e.edited
}

func (e *Editor) Content() string {
	return e.Text
}

func (e *Editor) Save() {
	err := os.WriteFile(e.path, []byte(e.Text), 0644)
	if err != nil {
		fmt.Println("Save Error :", err)
	}

	e.edited = false
}
