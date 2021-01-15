package goatc_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bzimmer/httpwares"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/goatc/pkg/goatc"
)

func Test_ExportError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	c, err := goatc.NewClient(
		goatc.WithTransport(&httpwares.TestDataTransport{
			Status:      http.StatusForbidden,
			Filename:    "single_error.json",
			ContentType: "application/json"}),
		goatc.WithTokenCredentials("fooAccessToken", "", time.Time{}))
	a.NoError(err)
	a.NotNil(c)

	exp, err := c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("oh noes!", err.Error())

	c, err = goatc.NewClient(
		goatc.WithTransport(&httpwares.TestDataTransport{
			Status:      http.StatusForbidden,
			Filename:    "multiple_errors.json",
			ContentType: "application/json"}),
		goatc.WithTokenCredentials("fooAccessToken", "", time.Time{}))
	a.NoError(err)
	a.NotNil(c)

	exp, err = c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("", err.Error())
	a.Equal(2, len(err.(*goatc.Fault).Errs))
}
