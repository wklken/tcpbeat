package main

import (
	"os"

	"github.com/wklken/tcpbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
