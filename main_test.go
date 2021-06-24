package main_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./kurogo", "-f", "testdata/sample.toml")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout getting failed: %v\n", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("command start failed: %v\n", err)
	}

	time.Sleep(2 * time.Second)
	current := time.Now().Local()
	os.Chtimes("main_test.go", current, current)

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	got := buf.String()

	want := "run command"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}
