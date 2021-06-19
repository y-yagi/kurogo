package runner

import (
	"fmt"
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

func NewRunner(filename string, logger *log.KurogoLogger) (*Runner, error) {
	r := &Runner{eventCh: make(chan string, 1000), logger: logger}

	if err := r.buildWatcher(); err != nil {
		return nil, err
	}

	if err := r.parseConfig(filename); err != nil {
		return nil, err
	}

	return r, nil
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
						r.logger.Printf(log.Red, "'%v' failed! %v\n", action.Command, err)
					}

					if len(string(stdoutStderr)) != 0 {
						r.logger.Printf(nil, "%s\n", stdoutStderr)
					}

					if err == nil {
						r.logger.Printf(log.Green, "'%v' success!\n", action.Command)
					}
				}
			}
		}
	}

	<-done
	return nil
}

func (r *Runner) Terminate() error {
	return r.watcher.Close()
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
				r.logger.Printf(log.Red, "error: %v\n", err)
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

func (r *Runner) parseConfig(filename string) error {
	if _, err := toml.DecodeFile(filename, &r.cfg); err != nil {
		return err
	}

	for k, action := range r.cfg.Actions {
		m := map[string]bool{}
		for _, extension := range action.Extensions {
			m[extension] = true
		}
		r.cfg.Actions[k].extensionsMap = m
	}

	return nil
}

func (r *Runner) buildWatcher() error {
	var err error
	if r.watcher, err = rnotify.NewWatcher(); err != nil {
		return err
	}

	return nil
}
