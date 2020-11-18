package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var sites = &cli.Command{
	Name:  "sites",
	Usage: "Return all site names",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "simple",
			Aliases: []string{"s"},
			Value:   false,
			Usage:   "Output the site names one line at a time (default: json)",
		}},
	Action: func(c *cli.Context) error {
		i := 0
		s := make([]string, len(config))
		for k := range config {
			s[i] = k
			i++
		}
		if c.Bool("simple") {
			for i := range s {
				fmt.Println(s[i])
			}
		} else {
			_ = encoder.Encode(s)
		}
		return nil
	},
}
