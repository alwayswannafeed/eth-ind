package main

import (
	"os"

	"github.com/alwayswannafeed/eth-ind/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
