package main_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path"
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
	time.Sleep(1 * time.Second)
	os.Chtimes("testdata/sample.toml", current, current)
	time.Sleep(1 * time.Second)
	os.Chtimes("testdata/dummy_test.rb", current, current)

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	got := buf.String()

	want := "run command"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	want = "testdata/sample.toml"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}

	want = "testdata/dummy_test.rb"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestIgnore(t *testing.T) {
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
	os.WriteFile("tmp/main.go", []byte("Hello"), 0644)
	defer os.Remove("tmp/main.go")

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	got := buf.String()

	want := "run command"
	if strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nnot included, but got \n\n%s\n", want, got)
	}
}

func TestPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kurogotest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./kurogo", "-f", "testdata/no_ignore.toml", "-p", tempDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout getting failed: %v\n", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("command start failed: %v\n", err)
	}

	time.Sleep(2 * time.Second)
	os.WriteFile(path.Join(tempDir, "main.go"), []byte("Hello"), 0644)

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	got := buf.String()

	want := "run command"
	if !strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nbut got \n\n%s\n", want, got)
	}
}

func TestMultipleCommand_FailFast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./kurogo", "-f", "testdata/no_ignore.toml")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout getting failed: %v\n", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("command start failed: %v\n", err)
	}

	time.Sleep(2 * time.Second)
	os.WriteFile("tmp/dummy.toml", []byte("Hello"), 0644)
	defer os.Remove("tmp/dummy.toml")

	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	got := buf.String()

	want := "'echo 2' success!"
	if strings.Contains(got, want) {
		t.Fatalf("expected \n%s\n\nnot included, but got \n\n%s\n", want, got)
	}
}
