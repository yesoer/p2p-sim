package gui

import (
	"distributed-sys-emulator/backend"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/fyne-io/terminal"
)

type Editor struct {
	term   fyne.CanvasObject
	path   string
	edited bool // TODO : unused atm
}

// Declare conformance with the Component interface
var _ Component = (*Editor)(nil)

// Declare conformance with the Component interface
func NewEditor(path string, _ fyne.Window, changeCB func(e *Editor)) *Editor {
	path, err := expandTilde(path)
	if err != nil {
		log.Print("Error :", err)
	}

	t := terminal.New()
	// TODO : change start dir
	t.SetStartDir(path)

	// shell start
	go func() {
		err := t.RunLocalShell()
		if err != nil {
			log.Print("Error :", err)
		}
	}()

	// delayed vim start to wait for shell
	// TODO : is the waiting improvable ?
	// TODO : set workdir
	// TODO : how to update code in network ? extra routine to continuously check for new save ?
	go func() {
		time.Sleep(time.Millisecond * 500)
		s := fmt.Sprintf("vim %s\n", path)
		_, err := t.Write([]byte(s))
		if err != nil {
			log.Print("Error :", err)
		}
	}()

	editor := &Editor{t, path, false}

	// TODO : check path for changes to support changeCB
	editor.ProcessOnChangeCB(path, changeCB)

	return editor
}

// starts a routine to detect write events on the given path
func (e *Editor) ProcessOnChangeCB(path string, cb func(e *Editor)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// listen for changes
	go func() {
		defer watcher.Close()
		for {
			err = watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			}
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("editor fail on receiving file events")
					continue
				}

				if event.Has(fsnotify.Write) {
					fmt.Println("cb")
					cb(e)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				fmt.Println("error ", err)
			}
		}
	}()
}

func (e *Editor) GetCanvasObj() fyne.CanvasObject {
	return e.term
}

func expandTilde(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}

func (e *Editor) GetContent() (backend.Code, error) {
	content, err := os.ReadFile(e.path)
	return backend.Code(content), err
}
