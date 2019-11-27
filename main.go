package main

import (
	"github.com/nokamoto/print-github-stats/cmd"
)

var (
	buildTimeTimestamp = ":timestamp:"
	buildTimeHash      = ":hash:"
)

func main() {
	cmd.BuildTimeVersion.Timestamp = buildTimeTimestamp
	cmd.BuildTimeVersion.Hash = buildTimeHash

	cmd.Execute()
}
