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

	"github.com/rs/zerolog/log"
)

// ExportService .
type ExportService service

// Export .
func (s *ExportService) Export(ctx context.Context) (*Export, error) {
	req, err := s.client.newAPIRequest(http.MethodPost, "export")
	if err != nil {
		return nil, err
	}
	exp := &Export{}
	err = s.client.Do(ctx, req, &exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// Status .
func (s *ExportService) Status(ctx context.Context, id int) (*Export, error) {
	uri := fmt.Sprintf("export/%d", id)
	req, err := s.client.newAPIRequest(http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	exp := &Export{}
	err = s.client.Do(ctx, req, &exp)
	if err != nil {
		return nil, err
	}
	return exp, nil
}

// Download .
func (s *ExportService) Download(ctx context.Context, id int) ([]*Stats, error) {
	uri := fmt.Sprintf("export/%d/download", id)
	req, err := s.client.newAPIRequest(http.MethodGet, uri)
	if err != nil {
		return nil, err
	}
	res, err := s.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// unfortunately the entire body needs to be read into memory first
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

func parseReader(reader *csv.Reader) ([]*Stats, error) {
	coll := make([]*Stats, 0)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		stats, err := unmarshal(row)
		if err != nil {
			return nil, err
		}
		coll = append(coll, stats)
	}
	return coll, nil
}

func unmarshal(row []string) (*Stats, error) {
	log.Info().Interface("row", row).Msg("unmarshall")
	return &Stats{}, nil
}

// 	f := gj.NewFeature(
// 		gj.NewPointGeometry(make([]float64, 3)))

// 	parts := strings.Split(line, "|")
// 	if len(parts) != gnisLength {
// 		return nil, fmt.Errorf("found %d parts, expected %d", len(parts), gnisLength)
// 	}

// 	for i, s := range parts {
// 		switch i {
// 		case 0: // FEATURE_ID
// 			x, err := strconv.Atoi(s)
// 			if err != nil {
// 				return nil, err
// 			}
// 			f.ID = x
// 		case 1: // FEATURE_NAME
// 			f.Properties["name"] = s
// 		case 2: // FEATURE_CLASS
// 			f.Properties["class"] = s
// 		case 3: // STATE_ALPHA
// 			f.Properties["state"] = s
// 		case 4: // STATE_NUMERIC
// 		case 5: // COUNTY_NAME
// 		case 6: // COUNTY_NUMERIC
// 		case 7: // PRIMARY_LAT_DMS
// 		case 8: // PRIM_LONG_DMS
// 		case 9: // PRIM_LAT_DEC
// 			x, err := strconv.ParseFloat(s, 64)
// 			if err != nil {
// 				return nil, err
// 			}
// 			f.Geometry.Point[1] = x
// 		case 10: // PRIM_LONG_DEC
// 			x, err := strconv.ParseFloat(s, 64)
// 			if err != nil {
// 				return nil, err
// 			}
// 			f.Geometry.Point[0] = x
// 		case 11: // SOURCE_LAT_DMS
// 		case 12: // SOURCE_LONG_DMS
// 		case 13: // SOURCE_LAT_DEC
// 		case 14: // SOURCE_LONG_DEC
// 		case 15: // ELEV_IN_M
// 			if s == "" {
// 				// not important enough to care about _though_ 0 m elevation is a legit value -- hmmm
// 				continue
// 			}
// 			x, err := strconv.ParseFloat(s, 64)
// 			if err != nil {
// 				return nil, err
// 			}
// 			f.Geometry.Point[2] = x
// 		case 16: // ELEV_IN_FT
// 		case 17: // MAP_NAME
// 		case 18: // DATE_CREATED
// 		case 19: // DATE_EDITED
// 		default:
// 		}
// 	}
// 	return f, nil
// }
