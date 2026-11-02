// Package credentials defines credential selection independently of TOML, environment lookup, and HTTP.
package credentials

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const (
	UserEnvironment    = "SENTIARY_USER_API_KEY"
	ProjectEnvironment = "SENTIARY_PROJECT_API_KEY"
)

type Project struct {
	InlineKey           *string
	EnvironmentVariable *string
}

type Sources struct {
	UserEnvironment    string
	LocalUser          string
	ProjectEnvironment string
	References         map[string]string
}

var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Validate checks source declarations even when no credential values are available offline.
func Validate(projects map[string]Project) error {
	for _, alias := range aliases(projects) {
		project := projects[alias]
		if project.EnvironmentVariable == nil {
			continue
		}
		if !environmentName.MatchString(*project.EnvironmentVariable) {
			return fmt.Errorf("project %q: api_key.env must match [A-Za-z_][A-Za-z0-9_]*", alias)
		}
		if project.InlineKey != nil {
			return fmt.Errorf("project %q: api_key cannot combine literal and environment sources", alias)
		}
	}
	return nil
}

// Resolve never retries another source after authentication; it selects exactly one key per project.
func Resolve(projects map[string]Project, sources Sources) (map[string]string, error) {
	if err := Validate(projects); err != nil {
		return nil, err
	}
	globalProject := strings.TrimSpace(sources.ProjectEnvironment)
	globalUser := strings.TrimSpace(sources.UserEnvironment)
	localUser := strings.TrimSpace(sources.LocalUser)
	if len(projects) != 1 {
		globalProject = ""
	}
	resolved := make(map[string]string, len(projects))
	for _, alias := range aliases(projects) {
		project := projects[alias]
		if project.EnvironmentVariable != nil {
			name := *project.EnvironmentVariable
			key := strings.TrimSpace(sources.References[name])
			if key == "" {
				return nil, fmt.Errorf("project %q: api_key.env variable %q is missing or empty", alias, name)
			}
			resolved[alias] = key
			continue
		}
		key := ""
		if project.InlineKey != nil {
			key = strings.TrimSpace(*project.InlineKey)
		}
		if key == "" {
			key = localUser
		}
		if key == "" {
			key = globalProject
		}
		if key == "" {
			key = globalUser
		}
		if key == "" {
			return nil, fmt.Errorf("project %q: provide api_key (a literal string or env reference), or a global credential", alias)
		}
		resolved[alias] = key
	}
	return resolved, nil
}

func aliases(projects map[string]Project) []string {
	result := make([]string, 0, len(projects))
	for alias := range projects {
		result = append(result, alias)
	}
	slices.Sort(result)
	return result
}
