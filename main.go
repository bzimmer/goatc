package main

import (
	"os"

	"github.com/bzimmer/goatc/cmd/goatc"
)

func main() {
	err := goatc.Run()
	if err != nil {
		os.Exit(1)
	}
}
