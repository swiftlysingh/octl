package main

import (
	"os"

	"github.com/pp/octl/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
