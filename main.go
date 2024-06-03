package main

import (
	"flag"
	"github.com/yesoer/p2p-sim/bus"
	"github.com/yesoer/p2p-sim/core"
	"github.com/yesoer/p2p-sim/embed"
	fynegui "github.com/yesoer/p2p-sim/fynegui"
	"github.com/yesoer/p2p-sim/log"
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
