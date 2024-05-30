package main

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/core"
	"distributed-sys-emulator/embed"
	fynegui "distributed-sys-emulator/fynegui"
	"distributed-sys-emulator/log"
	"flag"
	"os"
)

func main() {
	flag.Parse()

	log.Info("Setup")
	tmpExamplesDir := embed.InitEmbeddedExamples()
	eb := bus.NewEventbus()

	log.Info("Init Core")
	network := core.NewNetwork(eb)
	network.Init(eb)

	log.Info("Run GUI")
	fynegui.RunGUI(eb)

	log.Info("Cleanup")
	err := os.RemoveAll(tmpExamplesDir)
	if err != nil {
		log.Error(err)
	}
}
