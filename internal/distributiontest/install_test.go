package distributiontest

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShellInstallerChecksPinnedReleaseChecksum(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX installer")
	}
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprint("valid=", valid), func(t *testing.T) {
			root := t.TempDir()
			mock := filepath.Join(root, "bin")
			fixture := filepath.Join(root, "fixture")
			install := filepath.Join(root, "install")
			for _, path := range []string{mock, fixture} {
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatal(err)
				}
			}
			arch := runtime.GOARCH
			asset := "sentiary-cli-" + runtime.GOOS + "-" + arch
			archive := asset + ".tar.gz"
			var content strings.Builder
			zipped := gzip.NewWriter(&content)
			tarred := tar.NewWriter(zipped)
			binary := []byte("#!/bin/sh\necho fixture-version\n")
			if err := tarred.WriteHeader(&tar.Header{Name: asset, Size: int64(len(binary)), Mode: 0755}); err != nil {
				t.Fatal(err)
			}
			if _, err := tarred.Write(binary); err != nil {
				t.Fatal(err)
			}
			if err := tarred.Close(); err != nil {
				t.Fatal(err)
			}
			if err := zipped.Close(); err != nil {
				t.Fatal(err)
			}
			data := []byte(content.String())
			digest := fmt.Sprintf("%x", sha256.Sum256(data))
			if !valid {
				digest = strings.Repeat("0", 64)
			}
			if err := os.WriteFile(filepath.Join(fixture, archive), data, 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(fixture, "SHA256SUMS"), []byte(digest+"  "+archive+"\n"), 0644); err != nil {
				t.Fatal(err)
			}
			script := `#!/bin/sh
case "$2" in */releases/download/v2.0.0/*) ;; *) echo wrong-release >&2; exit 9 ;; esac
name="${2##*/}"
cp "$SENTIARY_INSTALL_FIXTURE/$name" "$4"
`
			if err := os.WriteFile(filepath.Join(mock, "curl"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("sh", "../../install.sh", "v2.0.0")
			command.Env = append(os.Environ(), "SENTIARY_VERSION=", "INSTALL_DIR="+install, "PATH="+mock+string(os.PathListSeparator)+os.Getenv("PATH"), "SENTIARY_INSTALL_FIXTURE="+fixture)
			output, err := command.CombinedOutput()
			if valid {
				if err != nil {
					t.Fatalf("install: %v: %s", err, output)
				}
				if _, err := os.Stat(filepath.Join(install, "sentiary-cli")); err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil || !strings.Contains(string(output), "checksum mismatch") {
					t.Fatalf("bad checksum accepted: %v %s", err, output)
				}
				if _, err := os.Stat(filepath.Join(install, "sentiary-cli")); !os.IsNotExist(err) {
					t.Fatal("installed corrupt archive")
				}
			}
		})
	}
}
