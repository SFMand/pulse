package main

import (
	"io"
	"log/slog"

	"github.com/SFMand/pulse/cmd"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	cmd.Execute()
}
