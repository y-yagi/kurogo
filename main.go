package main

import (
	"fmt"
	"io"
	"os"

	"github.com/y-yagi/kurogo/internal/log"
	"github.com/y-yagi/kurogo/internal/runner"
)

const cmd = "kurogo"

var (
	logger *log.KurogoLogger
)

func main() {
	logger = log.NewKurogoLogger(os.Stdout, false)
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func msg(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "%s: %+v\n", cmd, err)
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) (exitCode int) {
	r, err := runner.NewRunner("kurogo.toml", logger)
	if err != nil {
		return msg(err, stderr)
	}

	if err = r.Run(); err != nil {
		return msg(err, stderr)
	}

	return 0
}
