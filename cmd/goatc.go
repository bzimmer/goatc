package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/bzimmer/goatc/pkg"
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

func apikeys(args []string) [][]string {
	var (
		i    int
		ok   bool
		api  string
		site map[string]string
	)
	if len(args) == 0 {
		for k := range config {
			args = append(args, k)
		}
	}
	res := make([][]string, len(args))
	for _, arg := range args {
		if site, ok = config[arg]; !ok {
			continue
		}
		if api, ok = site["api-key"]; !ok {
			continue
		}
		res[i] = []string{arg, api}
		i++
	}
	return res
}

func goatc(c *cli.Context) error {
	type R struct {
		Stats *pkg.ExportedStats
		Error error
	}

	wg := &sync.WaitGroup{}
	stats := make(chan *R)
	deadline := time.Now().Add(c.Duration("timeout"))

	for _, s := range apikeys(c.Args().Slice()) {
		wg.Add(1)
		go func(siteName, apiKey string) {
			defer wg.Done()
			ctx, cancel := context.WithDeadline(c.Context, deadline)
			defer cancel()

			log.Info().Str("site", siteName).Msg(c.App.Name)
			client, err := pkg.NewClient(
				pkg.WithSiteName(siteName),
				pkg.WithHTTPTracing(c.Bool("http-tracing")),
				pkg.WithAPICredentials(apiKey))
			if err != nil {
				stats <- &R{Stats: nil, Error: err}
			}
			exp, err := client.Export.Stats(ctx)
			stats <- &R{Stats: exp, Error: err}
		}(s[0], s[1])
	}

	go func() {
		wg.Wait()
		close(stats)
	}()

	errors := make([]error, 0)
	for ret := range stats {
		if ret.Error != nil {
			errors = append(errors, ret.Error)
			continue
		}
		err := encoder.Encode(ret.Stats)
		if err != nil {
			errors = append(errors, ret.Error)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// Run .
func Run() error {
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
			sites,
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
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				log.Error().Err(err).Msg(c.App.Name)
			}
		},
		Action: goatc,
	}
	ctx := context.Background()
	err = app.RunContext(ctx, os.Args)
	if err != nil {
		return err
	}
	return nil
}
