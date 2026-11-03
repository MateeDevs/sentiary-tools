package cli

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MateeDevs/sentiary-tools/go/internal/buildinfo"
	"github.com/hashicorp/go-envparse"
)

// CredentialEnvironment is invocation-local; it never changes the process environment.
type CredentialEnvironment struct {
	values map[string]string
}

func (environment CredentialEnvironment) Get(name string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}
	return environment.values[name]
}

// ReadEnvironment removes the one explicit dotenv option before command parsing.
// Operational settings must continue to read the process environment directly.
func ReadEnvironment(args []string) ([]string, CredentialEnvironment, error) {
	var environment CredentialEnvironment
	var cleaned []string
	path := ""
	seen := false
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument != "--env-file" && !strings.HasPrefix(argument, "--env-file=") {
			cleaned = append(cleaned, argument)
			continue
		}
		if seen {
			return nil, environment, &UsageError{Message: "--env-file may only be specified once"}
		}
		seen = true
		if argument == "--env-file" {
			index++
			if index == len(args) {
				return nil, environment, &UsageError{Message: "--env-file requires a path"}
			}
			path = args[index]
		} else {
			path = strings.TrimPrefix(argument, "--env-file=")
		}
		if path == "" {
			return nil, environment, &UsageError{Message: "--env-file requires a path"}
		}
	}
	if !seen || skipsEnvironment(cleaned) {
		return cleaned, environment, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, environment, fmt.Errorf("cannot read selected --env-file")
	}
	environment.values, err = envparse.Parse(bytes.NewReader(data))
	if err != nil {
		// Parser errors may contain a complete source line, including secrets.
		return nil, CredentialEnvironment{}, fmt.Errorf("invalid selected --env-file")
	}
	return cleaned, environment, nil
}

func skipsEnvironment(args []string) bool {
	if len(args) == 0 {
		return true
	}
	if buildinfo.IsVersionCommand(args) {
		return true
	}
	for _, argument := range args {
		name, _, _ := strings.Cut(argument, "=")
		if name == "--help" || name == "-help" || name == "--h" || name == "-h" {
			return true
		}
	}
	switch args[0] {
	case "help", "update":
		return true
	case "pull":
		offline := false
		for _, argument := range args[1:] {
			name, value, assigned := strings.Cut(argument, "=")
			if name != "--offline" && name != "-offline" {
				continue
			}
			enabled, _ := strconv.ParseBool(value)
			offline = !assigned || enabled
		}
		return offline
	}
	return false
}
