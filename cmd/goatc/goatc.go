package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

var (
	encoder json.Encoder
	config  map[string]map[string]string
)

func initConfig(c *cli.Context) error {
	data, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &config)
}

func initEncoding(c *cli.Context) error {
	writer := c.App.Writer
	compact := c.Bool("compact")
	if writer == nil {
		writer = os.Stdout
	}
	encoder = *json.NewEncoder(writer)
	if !compact {
		encoder.SetIndent("", " ")
	}
	encoder.SetEscapeHTML(false)
	return nil
}

func initLogging(c *cli.Context) error {
	monochrome := c.Bool("monochrome")
	level, err := zerolog.ParseLevel(c.String("verbosity"))
	if err != nil {
		return err
	}
	color.NoColor = monochrome
	zerolog.SetGlobalLevel(level)
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     c.App.ErrWriter,
			NoColor: monochrome,
		},
	)
	return nil
}

func flags() ([]cli.Flag, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "verbosity",
			Aliases: []string{"v"},
			Value:   "info",
			Usage:   "Log level (trace, debug, info, warn, error, fatal, panic)",
		},
		&cli.BoolFlag{
			Name:    "monochrome",
			Aliases: []string{"m"},
			Value:   false,
			Usage:   "Use monochrome logging, color enabled by default",
		},
		&cli.BoolFlag{
			Name:    "compact",
			Aliases: []string{"c"},
			Value:   false,
			Usage:   "Use compact JSON output",
		},
		&cli.BoolFlag{
			Name:  "http-tracing",
			Value: false,
			Usage: "Log all http calls (warning: this will log ids, keys, and other sensitive information)",
		},
		&cli.StringFlag{
			Name:  "config",
			Value: filepath.Join(home, ".goatc.yaml"),
			Usage: "File containing configuration settings",
		},
		&cli.DurationFlag{
			Name:    "timeout",
			Aliases: []string{"t"},
			Value:   time.Millisecond * 10000,
			Usage:   "Timeout duration (eg, 1ms, 2s, 5m, 3h)",
		},
	}, nil
}

func run() error {
	fs, err := flags()
	if err != nil {
		return err
	}
	app := &cli.App{
		Name:      "goatc",
		HelpName:  "goatc",
		Usage:     "Command line access to site stats on https://goatcounter.com",
		UsageText: "goatc - check site stats",
		Flags:     fs,
		Commands: []*cli.Command{
			sitesCommand,
			versionCommand,
			visitsCommand,
		},
		Before: func(c *cli.Context) error {
			fns := []cli.BeforeFunc{initLogging, initEncoding, initConfig}
			for _, fn := range fns {
				if e := fn(c); e != nil {
					return e
				}
			}
			return nil
		},
		Action: visitsCommand.Action,
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				log.Error().Err(err).Msg(c.App.Name)
			}
		},
	}
	ctx := context.Background()
	err = app.RunContext(ctx, os.Args)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
