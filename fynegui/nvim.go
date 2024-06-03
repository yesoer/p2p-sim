package fynegui

import (
	"github.com/yesoer/p2p-sim/log"
	"os"

	"fyne.io/fyne/v2"
	nvim "github.com/yesoer/fyne-nvim"
)

// Declare conformance with the EditorType interface
var _ EditorType = (*Nvim)(nil)

// Extension of the NeoVim widget
type Nvim struct {
	*nvim.NeoVim
}

// Create a new instance of the Nvim wrapper
func NewNvim(dirPth string) EditorType {
	log.Debug("Start Nvim at path: ", dirPth)

	if _, err := os.ReadDir(dirPth); err != nil {
		dirPth = "./"
	}

	neovim := nvim.New(dirPth)
	neovimWrap := &Nvim{neovim}
	neovimWrap.ExtendBaseWidget(neovimWrap)
	return neovimWrap
}

// TypedShortcut implements the Shortcutable interface
func (e *Nvim) TypedShortcut(sc fyne.Shortcut) {
	log.Debug("Typed Shortcut Nvim")
}
