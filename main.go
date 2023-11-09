package main

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/core"
	fynegui "distributed-sys-emulator/fyne-gui"
	"flag"
)

func main() {
	flag.Parse()

	eb := bus.NewEventbus()

	network := core.NewNetwork(eb)
	network.Init(eb)

	fynegui.RunGUI(eb)
}
