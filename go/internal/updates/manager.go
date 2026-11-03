package updates

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const BackgroundArgument = "__sentiary-update-check"
const dailyWindow = 24 * time.Hour

type State struct {
	LastAttempt time.Time `json:"lastAttempt"`
	Latest      string    `json:"latest"`
	RemindAfter time.Time `json:"remindAfter"`
	Skipped     string    `json:"skipped"`
}

// Manager injects network, process, clock, and terminal boundaries for offline tests.
type Manager struct {
	Version     string
	Executable  string
	Directory   string
	Interactive bool
	Disabled    bool
	Input       io.Reader
	Output      io.Writer
	Now         func() time.Time
	Latest      func(context.Context) (string, error)
	Install     func(context.Context, string, string) error
	Owned       func(string) bool
	Spawn       func(string) error
}

func (manager *Manager) lock() (*flock.Flock, error) {
	if !filepath.IsAbs(manager.Directory) {
		return nil, fmt.Errorf("cannot determine an absolute per-user update cache directory")
	}
	if err := os.MkdirAll(manager.Directory, 0700); err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(manager.Directory, "update.lock"))
	ok, err := lock.TryLock()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("another update invocation is active; try again later")
	}
	return lock, nil
}

func (manager *Manager) read() State {
	data, err := os.ReadFile(filepath.Join(manager.Directory, "state.json"))
	if err != nil {
		return State{}
	}
	var state State
	if json.Unmarshal(data, &state) != nil {
		return State{}
	}
	return state
}

func (manager *Manager) save(state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(manager.Directory, ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), filepath.Join(manager.Directory, "state.json"))
}

func (manager *Manager) automaticAllowed() bool {
	return manager.Interactive && !manager.Disabled && stableVersion.MatchString(manager.Version)
}

// AfterSuccess never changes the completed command's result and never waits on a check.
// The detached child is started only after any cached prompt has been handled.
func (manager *Manager) AfterSuccess(ctx context.Context, args []string) {
	if !manager.automaticAllowed() || !Eligible(args) {
		return
	}
	if !manager.Owned(manager.Executable) {
		fmt.Fprintln(manager.Output, ReinstallInstructions())
		return
	}
	lock, err := manager.lock()
	if err != nil {
		return
	}
	defer lock.Unlock()
	state := manager.read()
	if Newer(state.Latest, manager.Version) && state.Skipped != state.Latest && !manager.Now().Before(state.RemindAfter) {
		fmt.Fprintf(manager.Output, "Sentiary %s is available. [1] Update now / [2] Remind me later / [3] Skip this version: ", state.Latest)
		answer, readErr := bufio.NewReader(manager.Input).ReadString('\n')
		choice := strings.TrimSpace(answer)
		if readErr != nil {
			choice = "2"
		}
		switch choice {
		case "1", "Update now", "update now":
			if err := manager.Install(ctx, manager.Executable, state.Latest); err != nil {
				fmt.Fprintf(manager.Output, "Update failed: %v\n", err)
			} else {
				fmt.Fprintf(manager.Output, "Updated to %s.\n", state.Latest)
			}
			state.RemindAfter = manager.Now().Add(dailyWindow)
		case "3", "Skip this version", "skip this version":
			state.Skipped = state.Latest
		default:
			state.RemindAfter = manager.Now().Add(dailyWindow)
		}
		_ = manager.save(state)
	}
	if manager.Now().Sub(state.LastAttempt) < dailyWindow {
		return
	}
	// Reserve the daily window before starting the child. A failed check keeps the
	// previous cached release and retries only in the next daily window.
	state.LastAttempt = manager.Now()
	if manager.save(state) == nil {
		_ = lock.Unlock()
		_ = manager.Spawn(manager.Executable)
	}
}

// Background refreshes the reserved daily check under the same cross-process lock.
func (manager *Manager) Background(ctx context.Context) {
	if manager.Disabled || !stableVersion.MatchString(manager.Version) || !manager.Owned(manager.Executable) {
		return
	}
	lock, err := manager.lock()
	if err != nil {
		return
	}
	defer lock.Unlock()
	state := manager.read()
	latest, err := manager.Latest(ctx)
	if err != nil || !stableVersion.MatchString(latest) {
		return
	}
	state.Latest = latest
	_ = manager.save(state)
}

func Eligible(args []string) bool {
	// Other commands emit JSON or raw export streams; never append an update UI.
	if len(args) == 0 || (args[0] != "pull" && args[0] != "init") {
		return false
	}
	for _, argument := range args[1:] {
		name, value, assigned := strings.Cut(argument, "=")
		switch name {
		case "--help", "-help", "--h", "-h":
			return false
		case "--json", "-json", "--offline", "-offline", "--dry-run", "-dry-run":
			if !assigned || value != "false" {
				return false
			}
		case "--integration-request", "-integration-request":
			return false
		}
	}
	return true
}

// Manual bypasses automatic opt-outs and preferences, but never bypasses consent
// or installer ownership. Check-only requests do not need replacement ownership.
func (manager *Manager) Manual(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(manager.Output)
	check := flags.Bool("check", false, "Check immediately without prompting or installing.")
	yes := flags.Bool("yes", false, "Approve installation without an interactive prompt.")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("update does not accept positional arguments")
	}
	if !stableVersion.MatchString(manager.Version) {
		return fmt.Errorf("self-update requires a stable official release. %s", ReinstallInstructions())
	}
	lock, err := manager.lock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
	latest, err := manager.Latest(ctx)
	if err != nil {
		return err
	}
	if !stableVersion.MatchString(latest) {
		return fmt.Errorf("latest release is not stable")
	}
	if !Newer(latest, manager.Version) {
		fmt.Fprintln(manager.Output, "Sentiary is up to date.")
		return nil
	}
	fmt.Fprintf(manager.Output, "Sentiary %s is available (installed: %s).\n", latest, manager.Version)
	if *check {
		return nil
	}
	if !manager.Owned(manager.Executable) {
		return fmt.Errorf("installation ownership could not be verified. %s", ReinstallInstructions())
	}
	if !*yes {
		if !manager.Interactive {
			return fmt.Errorf("installation requires approval; rerun update --yes or use update --check")
		}
		fmt.Fprint(manager.Output, "Install this update? [y/N]: ")
		answer, err := bufio.NewReader(manager.Input).ReadString('\n')
		if err != nil || !strings.EqualFold(strings.TrimSpace(answer), "y") {
			return nil
		}
	}
	if err := manager.Install(ctx, manager.Executable, latest); err != nil {
		return err
	}
	fmt.Fprintf(manager.Output, "Updated to %s.\n", latest)
	return nil
}
