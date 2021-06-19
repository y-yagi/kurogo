package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/y-yagi/kurogo/internal/log"
	"github.com/y-yagi/rnotify"
)

type Runner struct {
	watcher *rnotify.Watcher
	eventCh chan string
	cfg     Config
	logger  *log.KurogoLogger
}

type Config struct {
	Actions []Action
}

type Action struct {
	Command       string
	Extensions    []string
	extensionsMap map[string]bool
}

func (r *Runner) Run() error {
	done := make(chan bool)

	if err := r.watch(); err != nil {
		return err
	}

	for {
		var filename string

		select {
		case filename = <-r.eventCh:
			time.Sleep(500 * time.Millisecond)
			r.discardEvents()

			for _, action := range r.cfg.Actions {
				if _, ok := action.extensionsMap[filepath.Ext(filename)]; ok {
					fmt.Printf("Run '%v'\n", action.Command)
					cmd := strings.Split(action.Command, " ")
					stdoutStderr, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
					if err != nil {
						logger.Printf(log.Red, "'%v' failed! %v\n", action.Command, err)
					}

					if len(string(stdoutStderr)) != 0 {
						logger.Printf(nil, "%s\n", stdoutStderr)
					}

					if err == nil {
						logger.Printf(log.Green, "'%v' success!\n", action.Command)
					}
				}
			}
		}
	}

	<-done
	return nil
}

func (r *Runner) watch() error {
	if err := r.watcher.Add("."); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-r.watcher.Events:
				// fmt.Printf("%v\n", event)
				if !ok {
					return
				}

				for _, action := range r.cfg.Actions {
					if _, ok := action.extensionsMap[filepath.Ext(event.Name)]; ok {
						r.eventCh <- event.Name
					}
				}
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}
				logger.Printf(log.Red, "error: %v\n", err)
			}
		}
	}()

	return nil
}

func (r *Runner) discardEvents() {
	for {
		select {
		case <-r.eventCh:
		default:
			return
		}
	}
}

func msg(err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintf(stderr, "%s: %+v\n", cmd, err)
		return 1
	}
	return 0
}

const cmd = "kurogo"

var (
	logger *log.KurogoLogger
)

func main() {
	logger = log.NewKurogoLogger(os.Stdout, false)
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) (exitCode int) {
	// cmd.Env = append(os.Environ(), command.envs...)

	cfg, err := parseConfig()
	if err != nil {
		return msg(err, stderr)
	}

	watcher, err := rnotify.NewWatcher()
	if err != nil {
		return msg(err, stderr)
	}

	r := Runner{
		eventCh: make(chan string, 1000),
		watcher: watcher,
		cfg:     *cfg,
	}

	defer r.watcher.Close()

	if err = r.Run(); err != nil {
		return msg(err, stderr)
	}

	return 0
}

func parseConfig() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile("kurogo.toml", &cfg); err != nil {
		return nil, err
	}

	for k, action := range cfg.Actions {
		m := map[string]bool{}
		for _, extension := range action.Extensions {
			m[extension] = true
		}
		cfg.Actions[k].extensionsMap = m
	}

	return &cfg, nil
}
