package main

import (
	"distributed-sys-emulator/backend"
	"distributed-sys-emulator/gui"
	"flag"

	"fyne.io/fyne/v2"
)

var networkSizeFlag = flag.Int("network-size", 2, "define how many nodes you want in your network")

func main() {
	// Init backend
	flag.Parse()
	network := backend.NewNetwork()
	network.Init(*networkSizeFlag)

	// Init frontend
	gui.RunGUI(network, fyne.NewSize(800, 600))
}
