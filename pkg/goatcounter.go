package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	goatcounterURI = "https://%s.goatcounter.com/api/v0"
)

// Client client
type Client struct {
	siteName    string
	accessToken string

	client *http.Client

	Export *ExportService
}

type service struct {
	client *Client
}

// Option .
type Option func(*Client) error

// WithHTTPTracing .
func WithHTTPTracing(debug bool) Option {
	return func(c *Client) error {
		if !debug {
			return nil
		}
		c.client.Transport = &VerboseTransport{
			Transport: c.client.Transport,
		}
		return nil
	}
}

// WithTransport transport
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) error {
		if transport != nil {
			c.client.Transport = transport
		}
		return nil
	}
}

// WithHTTPClient .
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client != nil {
			c.client = client
		}
		return nil
	}
}

// WithAPICredentials provides the API credentials
func WithAPICredentials(accessToken string) Option {
	return func(client *Client) error {
		client.accessToken = accessToken
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

// NewClient creates new clients
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{client: &http.Client{}}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// Services used for talking to Strava
	c.Export = &ExportService{client: c}

	return c, nil
}

func (c *Client) url(uri string) (*url.URL, error) {
	site := fmt.Sprintf(goatcounterURI, c.siteName)
	return url.Parse(fmt.Sprintf("%s/%s", site, uri))
}

func (c *Client) newAPIRequest(method, uri string) (*http.Request, error) {
	if c.accessToken == "" {
		return nil, errors.New("accessToken required")
	}
	u, err := c.url(uri)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.accessToken))
	return req, nil
}

// Do executes the request
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	if ctx == nil {
		return errors.New("context must be non-nil")
	}

	res, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return err
	}
	defer res.Body.Close()

	httpError := res.StatusCode >= http.StatusBadRequest

	var obj interface{}
	if httpError {
		obj = &Fault{}
	} else {
		obj = v
	}

	if obj != nil {
		err := json.NewDecoder(res.Body).Decode(obj)
		if err == io.EOF {
			err = nil // ignore EOF errors caused by empty response body
		}
		if httpError {
			return obj.(error)
		}
		return err
	}

	return nil
}
