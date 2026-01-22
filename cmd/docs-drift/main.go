package main

import (
	"os"

	"github.com/georg-nikola/docs-drift/internal/cli"
)

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	exitCode := cli.Run(os.Args[1:], Version)
	os.Exit(exitCode)
}
