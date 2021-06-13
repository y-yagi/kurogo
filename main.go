package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/y-yagi/rnotify"
)

type Runner struct {
	watcher    *rnotify.Watcher
	cmd        *exec.Cmd
	eventCh    chan string
	extensions map[string]bool
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
			if _, ok := r.extensions[filepath.Ext(filename)]; ok {
				fmt.Printf("Run command\n")
				r.cmd.Run()
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
				fmt.Printf("%v\n", event)
				if !ok {
					return
				}

				r.eventCh <- event.Name
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
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

func main() {
	// cmd.Env = append(os.Environ(), command.envs...)

	watcher, err := rnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	r := Runner{
		eventCh:    make(chan string, 1000),
		watcher:    watcher,
		cmd:        exec.Command("go", "build", "."),
		extensions: map[string]bool{".go": true},
	}

	defer r.watcher.Close()

	r.Run()
}
