package cmd

import "encoding/json"

var (
	compact     bool
	monochrome  bool
	httptracing bool
	verbosity   string
	config      string
	encoder     *json.Encoder

	accessToken string

	status   int
	download int
)
