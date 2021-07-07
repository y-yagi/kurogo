package runner

import (
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
	watcher             *rnotify.Watcher
	eventCh             chan Event
	cfg                 Config
	logger              *log.KurogoLogger
	actionWithExtension map[string]*Action
	actionWithFile      map[string]*Action
	path                string
}

type Config struct {
	Ignore  []string
	Actions []Action
}

type Action struct {
	Command    string
	Extensions []string
	Files      []string
}

type Event struct {
	action   *Action
	filename string
}

func NewRunner(filename string, logger *log.KurogoLogger, path string) (*Runner, error) {
	r := &Runner{
		eventCh:             make(chan Event, 1000),
		logger:              logger,
		actionWithExtension: map[string]*Action{},
		actionWithFile:      map[string]*Action{},
		path:                path,
	}

	if err := r.parseConfig(filename); err != nil {
		return nil, err
	}

	if err := r.buildWatcher(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Runner) Run() error {
	if err := r.watch(); err != nil {
		return err
	}

	for {
		event := <-r.eventCh
		time.Sleep(500 * time.Millisecond)
		r.discardEvents()

		r.executeCmd(event.action, event.filename)
	}
}

func (r *Runner) Terminate() error {
	return r.watcher.Close()
}

func (r *Runner) watch() error {
	if err := r.watcher.Add(r.path); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-r.watcher.Events:
				r.logger.DebugPrintf(nil, "file event: %+v\n", event)
				if !ok {
					return
				}

				if action, ok := r.actionWithExtension[filepath.Ext(event.Name)]; ok {
					r.eventCh <- Event{action: action, filename: event.Name}
				} else if action, ok := r.actionWithFile[filepath.Base(event.Name)]; ok {
					r.eventCh <- Event{action: action, filename: event.Name}
				}
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}

				if os.IsNotExist(err) {
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

	for m, action := range r.cfg.Actions {
		for _, extension := range action.Extensions {
			r.actionWithExtension[extension] = &r.cfg.Actions[m]
		}

		for _, file := range action.Files {
			r.actionWithFile[file] = &r.cfg.Actions[m]
		}
	}

	return nil
}

func (r *Runner) buildWatcher() error {
	var err error
	if r.watcher, err = rnotify.NewWatcher(); err != nil {
		return err
	}

	if len(r.cfg.Ignore) != 0 {
		r.watcher.Ignore(r.cfg.Ignore)
	}

	return nil
}

func (r *Runner) executeCmd(action *Action, filename string) {
	parsedCmd := strings.ReplaceAll(action.Command, "{{.Filename}}", filename)
	r.logger.Printf(nil, "Run '%v'\n", parsedCmd)
	cmd := strings.Split(parsedCmd, " ")
	stdoutStderr, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		r.logger.Printf(log.Red, "'%v' failed! %v\n", action.Command, err)
	}

	if len(string(stdoutStderr)) != 0 {
		r.logger.Printf(nil, "%s\n", stdoutStderr)
	}

	if err == nil {
		r.logger.Printf(log.Green, "'%v' success!\n", parsedCmd)
	}
}
