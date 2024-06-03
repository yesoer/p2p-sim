package fynegui

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the EditorType interface
var _ EditorType = (*CodeEntry)(nil)

// Extension of the entry widget to handle specific shortcuts
type CodeEntry struct {
	widget.Entry

	path string
}

// Create a new instance of the CodeEntry wrapper
func NewCodeEntry(dirPth string, eb bus.EventBus) *CodeEntry {
	entries, err := os.ReadDir(dirPth)
	if err != nil {
		log.Error(err)
	}

	defaultFile := ""
	for _, entry := range entries {
		if !entry.IsDir() {
			defaultFile = dirPth + "/" + entry.Name()
		}
	}

	editor := &CodeEntry{}
	editor.MultiLine = true
	editor.path = defaultFile
	editor.TextStyle.Monospace = true
	editor.Wrapping = fyne.TextTruncate
	editor.ExtendBaseWidget(editor)

	if defaultFile != "" {
		editor.setText(defaultFile)
	}

	// when a new file is opened, display its contents
	eb.Bind(bus.OpenEvt, func(source bus.Source) {
		if source.Type != bus.File {
			return
		}

		editor.setText(source.Path)
		editor.path = source.Path
	})

	return editor
}

// Set the current text based on a file path
func (e *CodeEntry) setText(pth string) {
	b, err := os.ReadFile(pth)
	if err != nil {
		log.Error(err)
		return
	}

	e.Entry.SetText(string(b))
}

// Overwrites the shortcutable interface of the underlying multi line entry
func (e *CodeEntry) TypedShortcut(sc fyne.Shortcut) {
	// Save to disk on Ctrl+S or Cmd+S
	if sh, ok := sc.(*desktop.CustomShortcut); ok {
		cmd := runtime.GOOS == "darwin" && sh.Modifier == fyne.KeyModifierSuper
		ctrl := runtime.GOOS != "darwin" && sh.Modifier == fyne.KeyModifierControl
		if sh.KeyName == "S" && (cmd || ctrl) {
			e.save()
			return
		}
	}

	// Else hand down to the entries shortcut handler
	e.Entry.TypedShortcut(sc)
}

// Save the current code buffer to the file
func (e *CodeEntry) save() {
	log.Debug("Store code ", e.Text, " to ", e.path)
	err := os.WriteFile(e.path, []byte(e.Text), 0644)
	if err != nil {
		log.Error(err)
	}
}
