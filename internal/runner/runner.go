package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	actionWithPattern   map[*regexp.Regexp]*Action
	path                string
}

type Config struct {
	Ignore  []string
	Actions []Action
}

type Action struct {
	Commands   []string
	Extensions []string
	Files      []string
	Patterns   []string
}

type Event struct {
	filename string
}

func NewRunner(filename string, logger *log.KurogoLogger, path string) (*Runner, error) {
	r := &Runner{
		eventCh:             make(chan Event, 1000),
		logger:              logger,
		actionWithExtension: map[string]*Action{},
		actionWithFile:      map[string]*Action{},
		actionWithPattern:   map[*regexp.Regexp]*Action{},
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
		var actions []*Action
		event := <-r.eventCh
		if action, ok := r.actionWithExtension[filepath.Ext(event.filename)]; ok {
			actions = append(actions, action)
		}
		if action, ok := r.actionWithFile[filepath.Base(event.filename)]; ok {
			actions = append(actions, action)
		}
		for reg, action := range r.actionWithPattern {
			if matched := reg.Match([]byte(event.filename)); matched {
				actions = append(actions, action)
			}
		}

		if len(actions) != 0 {
			time.Sleep(500 * time.Millisecond)
			r.discardEvents()

			for _, action := range actions {
				r.executeCmd(action, event.filename)
			}
		}
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

				r.eventCh <- Event{filename: event.Name}
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

		for _, pattern := range action.Patterns {
			reg, err := regexp.Compile(pattern)
			if err != nil {
				return err
			}
			r.actionWithPattern[reg] = &r.cfg.Actions[m]
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
	for _, command := range action.Commands {
		parsedCmd := strings.ReplaceAll(command, "{{.Filename}}", filename)
		r.logger.Printf(nil, "Run '%v'\n", parsedCmd)
		cmd := strings.Split(parsedCmd, " ")
		stdoutStderr, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			r.logger.Printf(log.Red, "'%v' failed! %v\n", command, err)
		}

		if len(string(stdoutStderr)) != 0 {
			r.logger.Printf(nil, "%s\n", stdoutStderr)
		}

		if err == nil {
			r.logger.Printf(log.Green, "'%v' success!\n", parsedCmd)
		}
	}
}
