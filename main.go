package main

import (
	"flag"
)

var networkSizeFlag = flag.Int("network-size", 2, "define how many nodes you want in your network")

func main() {
	// Init backend
	flag.Parse()
	network := backend.NewNetwork()
	network.Init(*networkSizeFlag)
}
