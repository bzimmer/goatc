package goatc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const columnCount = 14

// ExportService .
type ExportService service

// Export initiates the export process
func (s *ExportService) Export(ctx context.Context) (*Export, error) {
	uri := "export"
	req, err := s.client.newAPIRequest(ctx, http.MethodPost, uri)
	if err != nil {
		return nil, err
	}
	exp := &Export{}
	err = s.client.do(req, &exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// Status checks the status of an export
func (s *ExportService) Status(ctx context.Context, id int) (*Export, error) {
	uri := fmt.Sprintf("export/%d", id)
	req, err := s.client.newAPIRequest(ctx, http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	exp := &Export{}
	err = s.client.do(req, &exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// Download returns the contents of the download csv file
func (s *ExportService) Download(ctx context.Context, id int) ([]*Stats, error) {
	uri := fmt.Sprintf("export/%d/download", id)
	req, err := s.client.newAPIRequest(ctx, http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	res, err := s.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	switch res.Header.Get("Content-Type") {
	case "application/gzip":
		// decompress
		reader, err = gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
	default:
		reader = bytes.NewReader(body)
	}
	return parseReader(csv.NewReader(reader))
}

// Stats initiates an export and polls for the download file to be available
func (s *ExportService) Stats(ctx context.Context) (*ExportedStats, error) {
	exp, err := s.Export(ctx)
	if err != nil {
		return nil, err
	}

	exp, err = s.Status(ctx, exp.ID)
	if err != nil {
		return nil, err
	}

	for exp.FinishedAt == nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(250 * time.Millisecond):
			exp, err = s.Status(ctx, exp.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	stats, err := s.Download(ctx, exp.ID)
	if err != nil {
		return nil, err
	}

	return &ExportedStats{Export: exp, Stats: stats}, nil
}

func parseReader(reader *csv.Reader) ([]*Stats, error) { // nolint:funlen,gocyclo
	coll := make([]*Stats, 0)

	// https://github.com/zgoat/goatcounter/blob/80b2b59ab175f192b6e50878596598aec95de605/export.go#L99

	// consume the header row
	header, err := reader.Read()
	switch err {
	case nil:
	case io.EOF:
		return coll, nil
	default:
		return nil, err
	}
	if len(header) != columnCount {
		log.Error().Strs("header", header).Int("count", len(header)).Msg("incorrect column count")
		return nil, fmt.Errorf("incorrect column count, found %d, expected %d", len(header), columnCount)
	}

	n := 0
	for {
		row, err := reader.Read()
		switch err {
		case nil:
			n++
		case io.EOF:
			log.Debug().Int("rows", n).Msg("parse")
			return coll, nil
		default:
			return nil, err
		}
		if len(row) != columnCount {
			log.Error().Strs("row", row).Int("count", len(row)).Msg("incorrect column count")
			return nil, fmt.Errorf("incorrect column count, found %d, expected %d", len(row), columnCount)
		}
		stats := &Stats{}
		for i, w := range row {
			switch i {
			case 0: // {Version}Path
				stats.Path = w
			case 1: // Title
				stats.Title = w
			case 2: // Event
				if b, err := strconv.ParseBool(w); err == nil {
					stats.Event = b
				} else {
					return nil, err
				}
			case 3: // UserAgent
				stats.UserAgent = w
			case 4: // Browser
			case 5: // System
			case 6: // Session
				stats.Session = w
			case 7: // Bot
				// the docs define bot as a bool but the data returned is an integer
				if b, err := strconv.Atoi(w); err == nil {
					stats.Bot = b
				} else {
					return nil, err
				}
			case 8: // Referrer
				stats.Referrer = w
			case 9: // ReferrerScheme
				stats.ReferrerScheme = w
			case 10: // ScreenSize
				stats.ScreenSize = w
			case 11: // Location
				stats.Location = w
			case 12: // FirstVisit
				if b, err := strconv.ParseBool(w); err == nil {
					stats.FirstVisit = b
				} else {
					return nil, err
				}
			case 13: // Date
				if t, err := time.Parse(time.RFC3339, w); err == nil {
					stats.Date = &t
				} else {
					return nil, err
				}
			}
		}
		coll = append(coll, stats)
	}
}
