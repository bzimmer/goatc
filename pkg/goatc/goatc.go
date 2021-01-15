package goatc

//go:generate genwith --client --do --auth --package goatc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

const baseURL = "https://%s.goatcounter.com/api/v0"

// Client client
type Client struct {
	config   oauth2.Config
	token    oauth2.Token
	client   *http.Client
	siteName string

	Export *ExportService
}

func withServices() Option {
	return func(c *Client) error {
		c.Export = &ExportService{client: c}
		return nil
	}
}

// WithSiteName .
func WithSiteName(siteName string) Option {
	return func(c *Client) error {
		c.siteName = siteName
		return nil
	}
}

func (c *Client) url(uri string) (*url.URL, error) {
	site := fmt.Sprintf(baseURL, c.siteName)
	return url.Parse(fmt.Sprintf("%s/%s", site, uri))
}

func (c *Client) newAPIRequest(ctx context.Context, method, uri string) (*http.Request, error) {
	if c.token.AccessToken == "" {
		return nil, errors.New("accessToken required")
	}
	u, err := c.url(uri)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token.AccessToken))
	return req, nil
}
