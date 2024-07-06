package fynegui

import (
	"os"

	"github.com/yesoer/p2p-sim/log"

	fynenvim "github.com/yesoer/fyne-nvim"
)

// Declare conformance with the EditorType interface
var _ EditorType = (*Nvim)(nil)

// Extension of the NeoVim widget
type Nvim struct {
	*fynenvim.NeoVim
}

// Create a new instance of the Nvim wrapper
func NewNvim(dirPth string) EditorType {
	log.Debug("Start Nvim at path: ", dirPth)

	if _, err := os.ReadDir(dirPth); err != nil {
		dirPth = "./"
	}

	neovim := fynenvim.New(dirPth)

	// if only one window is open prevent quitting it as it crashes the widget
	smart_quit_vim := `
						" quit single window 
						" -> if only one window : go into explorer
						cnoreabbrev <expr> q winnr('$') <= 1 ? 'Ex' : 'q'
						cnoreabbrev <expr> quit winnr('$') <= 1 ? 'Ex' : 'quit'

						" (if changed) save window and quit
						" -> if only one window : save and go into explorer
						command! WEx execute 'w | Ex'
						cnoreabbrev <expr> wq winnr('$') <= 1 ? 'WEx' : 'wq'
						cnoreabbrev <expr> x winnr('$') <= 1 ? 'WEx' : 'x'
						cnoreabbrev <expr> exit winnr('$') <= 1 ? 'WEx' : 'exit'
						
						" quit all windows 
						" -> close all but one and go into explorer
						command! OnlyEx execute 'only | Ex'
						cnoreabbrev quitall OnlyEx
						cnoreabbrev qall OnlyEx
						cnoreabbrev qa OnlyEx
						
						" (if changed) save all and quit all windows
						" -> save all and close all but one and go into explorer
						command! WAllOnlyEx execute 'wa | OnlyEx'
						cnoreabbrev xall WAllOnlyEx
						cnoreabbrev xa WAllOnlyEx
						cnoreabbrev wqa WAllOnlyEx
						cnoreabbrev wqall WAllOnlyEx

						" normal mode
						nnoremap ZZ :w<CR>:Ex<CR>
						nnoremap ZQ :Ex<CR>
					  `

	_, err := neovim.Engine.Exec(smart_quit_vim, make(map[string]interface{}))
	if err != nil {
		log.Error(err)
	}

	neovimWrap := &Nvim{neovim}
	neovimWrap.ExtendBaseWidget(neovimWrap)
	return neovimWrap
}
