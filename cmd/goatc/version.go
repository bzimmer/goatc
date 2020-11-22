package goatc

import (
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/goatc/pkg"
)

var versionCommand = &cli.Command{
	Name:     "version",
	Category: "api",
	Usage:    "Version",
	Action: func(c *cli.Context) error {
		return encoder.Encode(map[string]string{
			"version":   pkg.BuildVersion,
			"timestamp": pkg.BuildTime,
		})
	},
}
