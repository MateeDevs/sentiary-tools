package updates

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testManager(t *testing.T) (*Manager, *bytes.Buffer) {
	t.Helper()
	var output bytes.Buffer
	manager := &Manager{
		Version: "v1.0.0", Executable: "fixture", Directory: t.TempDir(),
		Interactive: true, Input: strings.NewReader(""), Output: &output,
		Now:     func() time.Time { return time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC) },
		Owned:   func(string) bool { return true },
		Latest:  func(context.Context) (string, error) { t.Fatal("unexpected network check"); return "", nil },
		Install: func(context.Context, string, string) error { t.Fatal("unexpected installation"); return nil },
		Spawn:   func(string) error { return nil },
	}
	return manager, &output
}

func TestAutomaticCheckIsDeferredAndDaily(t *testing.T) {
	manager, _ := testManager(t)
	spawned := 0
	manager.Spawn = func(string) error { spawned++; return nil }
	manager.AfterSuccess(context.Background(), []string{"pull"})
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if spawned != 1 || manager.read().LastAttempt != manager.Now() {
		t.Fatalf("expected one reserved detached check, got %d", spawned)
	}
}

func TestAutomaticEOFRemindsForOneDay(t *testing.T) {
	manager, output := testManager(t)
	if err := manager.save(State{Latest: "v2.0.0", LastAttempt: manager.Now()}); err != nil {
		t.Fatal(err)
	}
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if manager.read().RemindAfter != manager.Now().Add(24*time.Hour) {
		t.Fatal("EOF must defer for 24 hours")
	}
	output.Reset()
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if output.Len() != 0 {
		t.Fatal("reminder prompted too soon")
	}
}

func TestAutomaticSkipAppliesOnlyToExactRelease(t *testing.T) {
	manager, output := testManager(t)
	manager.Input = strings.NewReader("3\n")
	if err := manager.save(State{Latest: "v2.0.0", LastAttempt: manager.Now()}); err != nil {
		t.Fatal(err)
	}
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if manager.read().Skipped != "v2.0.0" {
		t.Fatal("skip was not saved")
	}
	output.Reset()
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if output.Len() != 0 {
		t.Fatal("skipped release prompted again")
	}
	state := manager.read()
	state.Latest = "v2.0.1"
	if err := manager.save(state); err != nil {
		t.Fatal(err)
	}
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if !strings.Contains(output.String(), "v2.0.1") {
		t.Fatal("new release should prompt")
	}
}

func TestAutomaticUpdateFailureIsOnlyFeedback(t *testing.T) {
	manager, output := testManager(t)
	manager.Input = strings.NewReader("1\n")
	manager.Install = func(context.Context, string, string) error { return errors.New("fixture failure") }
	if err := manager.save(State{Latest: "v2.0.0", LastAttempt: manager.Now()}); err != nil {
		t.Fatal(err)
	}
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if !strings.Contains(output.String(), "Update failed: fixture failure") {
		t.Fatal("missing failure feedback")
	}
}

func TestAutomaticExclusions(t *testing.T) {
	for _, args := range [][]string{
		{"help"}, {"version"}, {"update"}, {"pull", "--offline"}, {"pull", "--json"},
		{"pull", "--dry-run"}, {"pull", "--integration-request", "request.json"}, {"list-projects"}, {"init", "--help"},
		{"init", "-help"}, {"init", "--help=true"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			manager, output := testManager(t)
			manager.Spawn = func(string) error { t.Fatal("excluded command spawned a check"); return nil }
			manager.AfterSuccess(context.Background(), args)
			if output.Len() != 0 {
				t.Fatal("excluded command produced feedback")
			}
		})
	}
}

func TestAutomaticEnvironmentGates(t *testing.T) {
	for _, gate := range []string{"disabled", "noninteractive", "development", "staging"} {
		t.Run(gate, func(t *testing.T) {
			manager, output := testManager(t)
			switch gate {
			case "disabled":
				manager.Disabled = true
			case "noninteractive":
				manager.Interactive = false
			case "development":
				manager.Version = "dev"
			case "staging":
				manager.Version = "staging-abcdef"
			}
			manager.Spawn = func(string) error { t.Fatal("disabled check spawned"); return nil }
			manager.AfterSuccess(context.Background(), []string{"pull"})
			if output.Len() != 0 {
				t.Fatal("disabled check produced feedback")
			}
		})
	}
}

func TestBackgroundFailurePreservesCache(t *testing.T) {
	manager, output := testManager(t)
	state := State{Latest: "v2.0.0", Skipped: "v2.0.0", LastAttempt: manager.Now()}
	if err := manager.save(state); err != nil {
		t.Fatal(err)
	}
	manager.Latest = func(context.Context) (string, error) { return "", errors.New("network failure") }
	manager.Background(context.Background())
	if manager.read() != state || output.Len() != 0 {
		t.Fatal("failed refresh changed cache or produced output")
	}
}

func TestAutomaticFailureRetriesInNextDailyWindow(t *testing.T) {
	manager, _ := testManager(t)
	now := manager.Now()
	manager.Now = func() time.Time { return now }
	spawned := 0
	manager.Spawn = func(string) error { spawned++; return errors.New("cannot start child") }
	manager.AfterSuccess(context.Background(), []string{"pull"})
	now = now.Add(23 * time.Hour)
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if spawned != 1 {
		t.Fatal("failed checks must not retry before daily window")
	}
	now = now.Add(time.Hour)
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if spawned != 2 {
		t.Fatal("failed checks must retry in next daily window")
	}
}

func TestManualCheckIgnoresAutomaticOptOutAndPreferences(t *testing.T) {
	manager, output := testManager(t)
	manager.Disabled = true
	manager.Interactive = false
	manager.Latest = func(context.Context) (string, error) { return "v2.0.0", nil }
	if err := manager.save(State{Latest: "v2.0.0", Skipped: "v2.0.0", RemindAfter: manager.Now().Add(24 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Manual(context.Background(), []string{"--check"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "v2.0.0 is available") {
		t.Fatal("manual check omitted release")
	}
}

func TestManualNoninteractiveRequiresConsent(t *testing.T) {
	manager, _ := testManager(t)
	manager.Interactive = false
	manager.Latest = func(context.Context) (string, error) { return "v2.0.0", nil }
	err := manager.Manual(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected consent instructions, got %v", err)
	}
}

func TestManualDeclineSucceeds(t *testing.T) {
	manager, _ := testManager(t)
	manager.Input = strings.NewReader("n\n")
	manager.Latest = func(context.Context) (string, error) { return "v2.0.0", nil }
	if err := manager.Manual(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestManualApprovedInstallFailureIsReturned(t *testing.T) {
	manager, _ := testManager(t)
	manager.Disabled = true
	manager.Interactive = false
	manager.Latest = func(context.Context) (string, error) { return "v2.0.0", nil }
	want := errors.New("install failed")
	manager.Install = func(context.Context, string, string) error { return want }
	if err := manager.Manual(context.Background(), []string{"--yes"}); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestConcurrentInvocationDoesNotCheckOrPrompt(t *testing.T) {
	manager, output := testManager(t)
	lock, err := manager.lock()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Unlock()
	manager.Spawn = func(string) error { t.Fatal("concurrent check spawned"); return nil }
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if output.Len() != 0 {
		t.Fatal("concurrent invocation prompted")
	}
}

func TestLegacyInstallGetsReinstallInstructions(t *testing.T) {
	manager, output := testManager(t)
	manager.Owned = func(string) bool { return false }
	manager.Spawn = func(string) error { t.Fatal("legacy installation spawned a check"); return nil }
	manager.AfterSuccess(context.Background(), []string{"pull"})
	if !strings.Contains(output.String(), "Reinstall with:") {
		t.Fatal("missing reinstall instructions")
	}
	if _, err := os.Stat(filepath.Join(manager.Directory, "state.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("legacy invocation wrote update state")
	}
}
