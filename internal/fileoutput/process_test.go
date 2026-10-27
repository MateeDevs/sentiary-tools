package fileoutput

import (
	"bufio"
	"context"
	"fmt"
	"github.com/MateeDevs/sentiary-tools/internal/download"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestProcessLockHelper(t *testing.T) {
	if os.Getenv("SENTIARY_LOCK_HELPER") != "1" {
		return
	}
	store, err := Open(download.Config{Name: "sentiary.toml", Directory: os.Getenv("SENTIARY_LOCK_ROOT")}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	fmt.Println("locked")
	var line string
	_, _ = fmt.Scanln(&line)
}

func TestTerminatedProcessReleasesLock(t *testing.T) {
	cfg := testConfig(t)
	child := exec.Command(os.Args[0], "-test.run=^TestProcessLockHelper$")
	child.Env = append(os.Environ(), "SENTIARY_LOCK_HELPER=1", "SENTIARY_LOCK_ROOT="+cfg.Directory)
	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	stdout, err := child.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	if line, _ := bufio.NewReader(stdout).ReadString('\n'); line != "locked\n" {
		t.Fatalf("child did not acquire lock: %q", line)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if store, err := OpenContext(ctx, cfg, nil, false); err == nil {
		store.Close()
		t.Fatal("parallel process acquired held lock")
	}
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	store, err := OpenContext(ctx2, cfg, nil, false)
	if err != nil {
		t.Fatalf("dead process left blocking lock: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}
