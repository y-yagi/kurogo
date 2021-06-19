package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/y-yagi/kurogo/internal/log"
	"github.com/y-yagi/kurogo/internal/runner"
)

const cmd = "kurogo"

var (
	// Command line flags.
	flags       *flag.FlagSet
	showVersion bool
	configFile  string
	debug       bool

	logger  *log.KurogoLogger
	version = "devel"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] \n\n", cmd)
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flags.PrintDefaults()
}

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.BoolVar(&showVersion, "v", false, "print version number")
	flags.BoolVar(&debug, "d", false, "enable debug log")
	flags.StringVar(&configFile, "f", "kurogo.toml", "use file as a configuration file")
	flags.Usage = usage
}

func main() {
	logger = log.NewKurogoLogger(os.Stdout, false)
	setFlags()
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func msg(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "%s: %+v\n", cmd, err)
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) int {
	err := flags.Parse(args[1:])
	if err != nil {
		return msg(err, stderr)
	}

	if showVersion {
		fmt.Fprintf(stdout, "%s %s (runtime: %s)\n", cmd, version, runtime.Version())
		return 0
	}

	if debug {
		logger.EnableDebugLog()
	}

	r, err := runner.NewRunner(configFile, logger)
	if err != nil {
		return msg(err, stderr)
	}

	if err = r.Run(); err != nil {
		return msg(err, stderr)
	}

	return 0
}
