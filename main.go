package main

import (
	"distributed-sys-emulator/backend"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/gui"
	"flag"
	"time"
)

func main() {
	flag.Parse()

	eb := bus.NewEventbus()

	// Init backend with delay
	// The delay is because network on init publishes events
	go func() {
		time.Sleep(time.Millisecond * 100)
		network := backend.NewNetwork(eb)
		network.Init(eb)
	}()

	// Init frontend
	// Must run in main routine
	gui.RunGUI(eb)
}
