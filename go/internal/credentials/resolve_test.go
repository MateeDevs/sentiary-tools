package credentials

import (
	"reflect"
	"strings"
	"testing"
)

func pointer(value string) *string { return &value }

func TestResolveSources(t *testing.T) {
	for _, test := range []struct {
		name      string
		projects  map[string]Project
		sources   Sources
		want      map[string]string
		errorText string
	}{
		{"global project", map[string]Project{"app": {}}, Sources{ProjectEnvironment: " project-secret "}, map[string]string{"app": "project-secret"}, ""},
		{"global replaces inline", map[string]Project{"app": {InlineKey: pointer("inline-secret")}}, Sources{ProjectEnvironment: "project-secret"}, map[string]string{"app": "project-secret"}, ""},
		{"different references", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}, "web": {EnvironmentVariable: pointer("WEB_KEY")}}, Sources{References: map[string]string{"MOBILE_KEY": "mobile-secret", "WEB_KEY": "web-secret"}}, map[string]string{"app": "mobile-secret", "web": "web-secret"}, ""},
		{"reference ignores users", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}}, Sources{UserEnvironment: "user-secret", LocalUser: "local-secret", References: map[string]string{"MOBILE_KEY": "project-secret"}}, map[string]string{"app": "project-secret"}, ""},
		{"mixed sources", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}, "web": {InlineKey: pointer("inline-secret")}}, Sources{UserEnvironment: "user-secret", References: map[string]string{"MOBILE_KEY": "mobile-secret"}}, map[string]string{"app": "mobile-secret", "web": "user-secret"}, ""},
		{"missing reference no fallback", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}}, Sources{UserEnvironment: "user-secret"}, nil, "MOBILE_KEY"},
		{"empty reference no fallback", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}}, Sources{LocalUser: "local-secret", References: map[string]string{"MOBILE_KEY": ""}}, nil, "missing or empty"},
		{"blank reference no fallback", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}}, Sources{UserEnvironment: "user-secret", References: map[string]string{"MOBILE_KEY": " \n\t"}}, nil, "missing or empty"},
		{"conflict user environment", map[string]Project{"app": {}}, Sources{ProjectEnvironment: "project-secret", UserEnvironment: "user-secret"}, nil, "conflicts"},
		{"conflict local user", map[string]Project{"app": {}}, Sources{ProjectEnvironment: "project-secret", LocalUser: "local-secret"}, nil, "conflicts"},
		{"conflict explicit reference", map[string]Project{"app": {EnvironmentVariable: pointer("MOBILE_KEY")}}, Sources{ProjectEnvironment: "project-secret"}, nil, "conflicts with api_key.env"},
		{"multiple projects", map[string]Project{"app": {}, "web": {}}, Sources{ProjectEnvironment: "project-secret"}, nil, "exactly one declared project"},
		{"legacy environment precedence", map[string]Project{"app": {InlineKey: pointer("inline-secret")}}, Sources{UserEnvironment: "user-secret", LocalUser: "local-secret"}, map[string]string{"app": "user-secret"}, ""},
		{"legacy local precedence", map[string]Project{"app": {InlineKey: pointer("inline-secret")}}, Sources{UserEnvironment: " \t", LocalUser: "local-secret"}, map[string]string{"app": "local-secret"}, ""},
		{"legacy inline", map[string]Project{"app": {InlineKey: pointer(" inline-secret ")}}, Sources{}, map[string]string{"app": "inline-secret"}, ""},
		{"blank global project absent", map[string]Project{"app": {}}, Sources{ProjectEnvironment: " \n", UserEnvironment: "user-secret"}, map[string]string{"app": "user-secret"}, ""},
		{"missing credential", map[string]Project{"app": {}}, Sources{}, nil, "provide"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := Resolve(test.projects, test.sources)
			if test.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), test.errorText) {
					t.Fatalf("wrong error: %v", err)
				}
				if strings.Contains(err.Error(), "secret") {
					t.Fatal("credential leaked in error")
				}
				if got != nil {
					t.Fatal("partial credential result on failure")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatal("wrong credential selection")
			}
		})
	}
}

func TestEnvironmentNameValidation(t *testing.T) {
	for _, test := range []struct {
		name  string
		valid bool
	}{
		{"MOBILE_KEY", true}, {"_key123", true}, {"key", true}, {"", false}, {" ", false}, {" KEY", false}, {"KEY ", false}, {"1KEY", false}, {"KEY-NAME", false}, {"KÉY", false}, {"${KEY}", false}, {"KEY=VALUE", false}, {"KEY\x00", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(map[string]Project{"app": {EnvironmentVariable: pointer(test.name)}})
			if (err == nil) != test.valid {
				t.Fatalf("wrong name validation: %v", err)
			}
		})
	}
}

func TestReferenceAndInlineDeclarationConflictEvenWhenInlineEmpty(t *testing.T) {
	for _, inline := range []string{"", "inline-secret"} {
		t.Run(map[bool]string{true: "empty", false: "nonempty"}[inline == ""], func(t *testing.T) {
			if err := Validate(map[string]Project{"app": {InlineKey: pointer(inline), EnvironmentVariable: pointer("MOBILE_KEY")}}); err == nil {
				t.Fatal("accepted conflicting declarations")
			}
		})
	}
}
