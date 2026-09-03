package main

import (
	"os"

	"github.com/abraa/hotstack/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
