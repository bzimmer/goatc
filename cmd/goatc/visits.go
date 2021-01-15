package main

import (
	"context"
	"sync"
	"time"

	"github.com/bzimmer/goatc/pkg/goatc"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func apikeys(args []string) [][]string {
	var (
		ok   bool
		api  string
		site map[string]string
	)
	if len(args) == 0 {
		for k := range config {
			args = append(args, k)
		}
	}
	res := make([][]string, 0)
	for _, arg := range args {
		if site, ok = config[arg]; !ok {
			log.Warn().Str("site", arg).Msg("did not find site in config")
			continue
		}
		if api, ok = site["api-key"]; !ok {
			log.Warn().Str("site", arg).Msg("did not find api key for site")
			continue
		}
		res = append(res, []string{arg, api})
	}
	return res
}

var visitsCommand = &cli.Command{
	Name:  "visits",
	Usage: "Return all site visits",
	Action: func(c *cli.Context) error {
		type R struct {
			Stats *goatc.ExportedStats
			Error error
		}

		wg := &sync.WaitGroup{}
		stats := make(chan *R)
		deadline := time.Now().Add(c.Duration("timeout"))

		for _, k := range apikeys(c.Args().Slice()) {
			wg.Add(1)
			go func(siteName, apiKey string) {
				defer wg.Done()
				ctx, cancel := context.WithDeadline(c.Context, deadline)
				defer cancel()

				log.Info().Str("site", siteName).Msg(c.App.Name)
				client, err := goatc.NewClient(
					goatc.WithTokenCredentials(apiKey, "", time.Time{}),
					goatc.WithSiteName(siteName),
					goatc.WithHTTPTracing(c.Bool("http-tracing")))
				if err != nil {
					stats <- &R{Stats: nil, Error: err}
					return
				}
				exp, err := client.Export.Stats(ctx)
				stats <- &R{Stats: exp, Error: err}
			}(k[0], k[1])
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
	},
}
