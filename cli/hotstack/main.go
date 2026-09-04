package main

import (
	"os"

	"github.com/abraaosala/hotstack/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
