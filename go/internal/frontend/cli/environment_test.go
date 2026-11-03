package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExplicitEnvironmentParsing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	contents := "export SENTIARY_TEST_DOTENV='first'\nSENTIARY_TEST_DOTENV=\"literal $OTHER ${OTHER} $(command)\" # comment\n"
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	args, environment, err := ReadEnvironment([]string{"pull", "--env-file", path})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []string{"pull"}) {
		t.Fatalf("unexpected arguments: %v", args)
	}
	if got := environment.Get("SENTIARY_TEST_DOTENV"); got != "literal $OTHER ${OTHER} $(command)" {
		t.Fatal("dotenv must use literal, last-assignment-wins values")
	}
	if _, exists := os.LookupEnv("SENTIARY_TEST_DOTENV"); exists {
		t.Fatal("dotenv must not mutate process environment")
	}
	t.Setenv("SENTIARY_TEST_DOTENV", "")
	if environment.Get("SENTIARY_TEST_DOTENV") != "" {
		t.Fatal("explicitly empty process value must take precedence")
	}
}

func TestEnvironmentFailuresAreSecretSafe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials")
	if err := os.WriteFile(path, []byte("super-secret-value-without-assignment"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, err := ReadEnvironment([]string{"init", "--env-file=" + path})
	if err == nil || strings.Contains(err.Error(), "super-secret") || strings.Contains(err.Error(), path) {
		t.Fatalf("expected sanitized parse error, got %v", err)
	}
}

func TestEnvironmentExcludedCommands(t *testing.T) {
	for _, args := range [][]string{
		{"help"}, {"version"}, {"update"}, {"pull", "--offline"}, {"init", "--help"},
		{"init", "-help"}, {"init", "--help=true"}, {"pull", "--offline=1"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			_, _, err := ReadEnvironment(append(args, "--env-file=/does/not/exist"))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEnvironmentRejectsRepeatedSelection(t *testing.T) {
	_, _, err := ReadEnvironment([]string{"init", "--env-file=first", "--env-file=second"})
	if !IsUsageError(err) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestEnvironmentMissingFileFailsDespiteProcessCredentials(t *testing.T) {
	t.Setenv("SENTIARY_USER_API_KEY", "process-key")
	_, _, err := ReadEnvironment([]string{"list-projects", "--env-file=/does/not/exist"})
	if err == nil {
		t.Fatal("explicit missing file must fail")
	}
}

func TestEnvironmentOfflineUsesLastFlagValue(t *testing.T) {
	_, _, err := ReadEnvironment([]string{"pull", "--offline", "--offline=false", "--env-file=/does/not/exist"})
	if err == nil {
		t.Fatal("online pull must read selected dotenv after offline is overridden")
	}
}
