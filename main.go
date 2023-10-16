package main

import (
	"distributed-sys-emulator/backend"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/gui"
	"flag"

	"fyne.io/fyne/v2"
)

var networkSizeFlag = flag.Int("network-size", 2, "define how many nodes you want in your network")

// TODO : use the eventbus designed in DEV to facilitate back and forth
//
//	       communication between network and backend.
//			  but with centralized definition of available topics or something like that
func main() {
	eb := bus.NewEventbus()

	// Init backend
	flag.Parse()
	network := backend.NewNetwork(eb)
	network.Init(*networkSizeFlag, eb)

	// Init frontend
	gui.RunGUI(*networkSizeFlag, fyne.NewSize(1000, 800), eb)
}
