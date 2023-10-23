package main

import (
	"distributed-sys-emulator/backend"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/gui"
	"flag"
)

func main() {
	flag.Parse()

	eb := bus.NewEventbus()

	network := backend.NewNetwork(eb)
	network.Init(eb)

	gui.RunGUI(eb)
}
