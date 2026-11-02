package configfile

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExportListsResolveFromConfigOnlineAndOffline(t *testing.T) {
	for _, offline := range []bool{false, true} {
		t.Run(map[bool]string{false: "online", true: "offline"}[offline], func(t *testing.T) {
			path := writeConfig(t, validConfig+"keys_file = 'selected.txt'\nrequired_keys_file = 'required.txt'\ncontexts = ['Exact.Context']\n")
			for name, content := range map[string]string{"selected.txt": "  welcome \n\n#literal\n", "required.txt": ""} {
				if err := os.WriteFile(filepath.Join(filepath.Dir(path), name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			config, err := load(path, t.TempDir(), userEnvironment(""), offline)
			if err != nil {
				t.Fatal(err)
			}
			options := config.Outputs["ios"].Options
			if options.Keys == nil || !reflect.DeepEqual(*options.Keys, []string{"welcome", "#literal"}) || len(options.RequiredKeys) != 0 || !reflect.DeepEqual(options.Contexts, []string{"Exact.Context"}) {
				t.Fatalf("incorrect resolved options: %+v", options)
			}
		})
	}
}

func TestEmptyKeysFileRemainsExplicit(t *testing.T) {
	path := writeConfig(t, validConfig+"keys_file = 'empty.txt'\n")
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "empty.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	config, err := LoadOffline(path, "")
	if err != nil {
		t.Fatal(err)
	}
	options := config.Outputs["ios"].Options
	if options.Keys == nil || len(*options.Keys) != 0 || options.IsDefault() {
		t.Fatal("empty file lost explicit selection")
	}
}

func TestInvalidExportListsFailOnlineAndOffline(t *testing.T) {
	for _, test := range []struct {
		name       string
		content    []byte
		diagnostic string
	}{
		{"duplicate", []byte("one\n\n one \n"), "selected.txt:3: duplicate"},
		{"invalid utf8", []byte{0xff}, "selected.txt:1: invalid UTF-8"},
		{"nul", []byte("one\x00\n"), "selected.txt:1:"},
		{"too long", []byte(strings.Repeat("a", 1025)), "selected.txt:1:"},
		{"missing", nil, "selected.txt"},
	} {
		for _, offline := range []bool{false, true} {
			t.Run(test.name+map[bool]string{false: " online", true: " offline"}[offline], func(t *testing.T) {
				path := writeConfig(t, validConfig+"keys_file = 'selected.txt'\n")
				if test.content != nil {
					if err := os.WriteFile(filepath.Join(filepath.Dir(path), "selected.txt"), test.content, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := load(path, "", userEnvironment(""), offline); err == nil || !strings.Contains(err.Error(), test.diagnostic) {
					t.Fatalf("wrong error: %v", err)
				}
			})
		}
	}
}

func TestEmptyContextRejected(t *testing.T) {
	if _, err := LoadOffline(writeConfig(t, validConfig+"contexts = ['']\n"), ""); err == nil {
		t.Fatal("accepted empty context")
	}
}
