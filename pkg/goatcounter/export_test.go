package goatcounter_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bzimmer/goatc/pkg/goatcounter"
	"github.com/bzimmer/httpwares"
	"github.com/stretchr/testify/assert"
)

func Test_ExportError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	c, err := goatcounter.NewClient(
		goatcounter.WithTransport(&httpwares.TestDataTransport{
			Status:      http.StatusForbidden,
			Filename:    "single_error.json",
			ContentType: "application/json"}),
		goatcounter.WithAPICredentials("fooAccessToken"))
	a.NoError(err)
	a.NotNil(c)

	exp, err := c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("oh noes!", err.Error())

	c, err = goatcounter.NewClient(
		goatcounter.WithTransport(&httpwares.TestDataTransport{
			Status:      http.StatusForbidden,
			Filename:    "multiple_errors.json",
			ContentType: "application/json"}),
		goatcounter.WithAPICredentials("fooAccessToken"))
	a.NoError(err)
	a.NotNil(c)

	exp, err = c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("", err.Error())
	a.Equal(2, len(err.(*goatcounter.Fault).Errs))
}
