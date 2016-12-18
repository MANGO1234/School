package main

import (
	"./gui"
	"./network"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <listening port> [public listening port]\n", os.Args[0])
		os.Exit(1)
	}

	localAddr := os.Args[1]
	publicAddr := os.Args[1]
	if len(os.Args) == 3 {
		publicAddr = os.Args[2]
	}

	networkManager, err := network.NewNetworkManager(localAddr, publicAddr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}

	gui.StartMainLoop(networkManager)
}
