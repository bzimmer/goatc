package pkg

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

const (
	columnCount = 12
)

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
	err = s.client.Do(req, &exp)
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
	err = s.client.Do(req, &exp)
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
		// @todo replace with something more sophisticated
		time.Sleep(500 * time.Millisecond)
		exp, err = s.Status(ctx, exp.ID)
		if err != nil {
			return nil, err
		}
	}

	stats, err := s.Download(ctx, exp.ID)
	if err != nil {
		return nil, err
	}

	return &ExportedStats{
		Export: exp,
		Stats:  stats,
	}, nil
}

func parseReader(reader *csv.Reader) ([]*Stats, error) {
	coll := make([]*Stats, 0)

	// consume the header row
	_, err := reader.Read()
	switch err {
	case nil:
	case io.EOF:
		return coll, nil
	default:
		return nil, err
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
		stats := &Stats{}
		for i, w := range row {
			switch i {
			case 0: // 1Path
				stats.Path = w
			case 1: // Title
				stats.Title = w
			case 2: // Event
				if b, err := strconv.ParseBool(w); err == nil {
					stats.Event = b
				} else {
					return nil, err
				}
			case 3: // Bot
				// the docs define bot as a bool but the data returned is an integer
				if i, err := strconv.Atoi(w); err == nil {
					stats.Bot = i
				} else {
					return nil, err
				}
			case 4: // Session
				stats.Session = w
			case 5: // FirstVisit
				if b, err := strconv.ParseBool(w); err == nil {
					stats.FirstVisit = b
				} else {
					return nil, err
				}
			case 6: // Referrer
				stats.Referrer = w
			case 7: // ReferrerScheme
				stats.ReferrerScheme = w
			case 8: // UserAgent
				stats.UserAgent = w
			case 9: // ScreenSize
				stats.ScreenSize = w
			case 10: // Location
				stats.Location = w
			case 11: // Date
				if t, err := time.Parse(time.RFC3339, w); err == nil {
					stats.Date = &t
				} else {
					return nil, err
				}
			default:
				log.Error().Strs("row", row).Int("count", i+1).Msg("too many columns")
				return nil, fmt.Errorf("too many rows, found %d, expected %d", i+1, columnCount)
			}
		}
		coll = append(coll, stats)
	}
}
