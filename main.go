package main

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/core"
	fynegui "distributed-sys-emulator/fynegui"
	"distributed-sys-emulator/log"
	"flag"
)

func main() {
	flag.Parse()

	eb := bus.NewEventbus()

	log.Info("Init Core")
	network := core.NewNetwork(eb)
	network.Init(eb)

	log.Info("Run GUI")
	fynegui.RunGUI(eb)
}
