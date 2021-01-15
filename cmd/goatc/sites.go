package main

import (
	"github.com/urfave/cli/v2"
)

var sitesCommand = &cli.Command{
	Name:  "sites",
	Usage: "Return all site names",
	Action: func(c *cli.Context) error {
		i := 0
		s := make([]string, len(config))
		for k := range config {
			s[i] = k
			i++
		}
		return encoder.Encode(s)
	},
}
