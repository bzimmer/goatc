package pkg_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/goatc/pkg"
)

func Test_ExportError(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	c, err := pkg.NewClient(
		pkg.WithTransport(&pkg.TestDataTransport{
			Status:      http.StatusForbidden,
			Filename:    "single_error.json",
			ContentType: "application/json"}),
		pkg.WithAPICredentials("fooAccessToken"))
	a.NoError(err)
	a.NotNil(c)

	exp, err := c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("oh noes!", err.Error())

	pkg.WithTransport(&pkg.TestDataTransport{
		Status:      http.StatusForbidden,
		Filename:    "multiple_errors.json",
		ContentType: "application/json"})(c)

	exp, err = c.Export.Status(context.Background(), 600)
	a.Error(err)
	a.Nil(exp)
	a.Equal("", err.Error())
	a.Equal(2, len(err.(*pkg.Fault).Errs))
}
