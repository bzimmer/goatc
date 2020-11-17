package cmd

import (
	"encoding/json"
	"io"
	"os"
)

func newEncoder(writer io.Writer, compact bool) *json.Encoder {
	if writer == nil {
		writer = os.Stdout
	}
	encoder := json.NewEncoder(writer)
	if !compact {
		encoder.SetIndent("", " ")
	}
	encoder.SetEscapeHTML(false)
	return encoder
}
