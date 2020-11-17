package main

import (
	"os"

	"github.com/bzimmer/goatc/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}
