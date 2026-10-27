package fileoutput

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

func testConfig(t *testing.T) download.Config {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return download.Config{Name: "sentiary.toml", Directory: directory}
}

func openStore(t *testing.T, config download.Config, selected []string, force bool) *Store {
	t.Helper()
	store, err := Open(config, selected, force)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	return store
}

func TestCleanupRemovesOnlyOwnedObsoleteFiles(t *testing.T) {
	config := testConfig(t)
	store := openStore(t, config, nil, false)
	old := filepath.Join(config.Directory, "old.strings")
	other := filepath.Join(config.Directory, "handwritten.txt")
	if err := os.WriteFile(other, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := store.Commit([]download.DownloadedFile{{File: download.File{Output: "ios", Path: old}, Content: []byte("old")}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Validate(nil); err != nil {
		t.Fatal(err)
	}
	if err := store.Commit(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("obsolete output remains")
	}
	if data, err := os.ReadFile(other); err != nil || string(data) != "keep" {
		t.Fatal("unrelated file changed")
	}
}

func TestManualChangesBlockOverwriteAndRemoval(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "overwrite", true: "removal"}[remove], func(t *testing.T) {
			config := testConfig(t)
			store := openStore(t, config, nil, false)
			file := download.File{Output: "ios", Path: filepath.Join(config.Directory, "catalog")}
			if err := store.Commit([]download.DownloadedFile{{File: file, Content: []byte("generated")}}); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file.Path, []byte("edited"), 0o644); err != nil {
				t.Fatal(err)
			}
			files := []download.File{file}
			if remove {
				files = nil
			}
			if err := store.Validate(files); err == nil {
				t.Fatal("manual edit was not protected")
			}
		})
	}
}

func TestUntrackedDestinationsRequireForce(t *testing.T) {
	config := testConfig(t)
	file := download.File{Output: "ios", Path: filepath.Join(config.Directory, "catalog")}
	if err := os.WriteFile(file.Path, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := openStore(t, config, nil, false)
	if err := store.Validate([]download.File{file}); err == nil {
		t.Fatal("untracked file was accepted")
	}
	store.force = true
	if err := store.Validate([]download.File{file}); err != nil {
		t.Fatal(err)
	}
}

func TestSelectedPullPreservesOtherOutputOwnership(t *testing.T) {
	config := testConfig(t)
	store := openStore(t, config, nil, false)
	ios := download.File{Output: "ios", Path: filepath.Join(config.Directory, "ios")}
	web := download.File{Output: "web", Path: filepath.Join(config.Directory, "web")}
	if err := store.Commit([]download.DownloadedFile{{File: ios, Content: []byte("ios")}, {File: web, Content: []byte("web")}}); err != nil {
		t.Fatal(err)
	}
	store.selected = []string{"ios"}
	if err := store.Validate(nil); err != nil {
		t.Fatal(err)
	}
	if err := store.Commit(nil); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(web.Path); err != nil || string(data) != "web" {
		t.Fatal("unselected output changed")
	}
	if len(store.previous["web"]) != 1 {
		t.Fatal("unselected ownership lost")
	}
}

func TestConcurrentPullWaitCanBeCancelled(t *testing.T) {
	config := testConfig(t)
	openStore(t, config, nil, false)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if other, err := OpenContext(ctx, config, nil, false); err == nil {
		other.Close()
		t.Fatal("concurrent pull acquired lock")
	}
}

func TestOutputCannotOverwriteConfiguration(t *testing.T) {
	config := testConfig(t)
	store := openStore(t, config, nil, true)
	if err := store.Validate([]download.File{{Path: filepath.Join(config.Directory, "sentiary.toml")}}); err == nil {
		t.Fatal("config overwrite accepted")
	}
}

func TestCaseOnlyPathChangeCannotDeleteTheReplacement(t *testing.T) {
	config := testConfig(t)
	store := openStore(t, config, nil, true)
	file := download.File{Output: "ios", Path: filepath.Join(config.Directory, "Catalog")}
	if err := store.Commit([]download.DownloadedFile{{File: file, Content: []byte("generated")}}); err != nil {
		t.Fatal(err)
	}
	file.Path = filepath.Join(config.Directory, "catalog")
	if err := store.Validate([]download.File{file}); err == nil {
		t.Fatal("case-only relocation was accepted")
	}
}

func TestSymbolicLinkDestinationsAreRejectedEvenWithForce(t *testing.T) {
	config := testConfig(t)
	file := filepath.Join(config.Directory, "catalog")
	if err := os.Symlink(filepath.Join(config.Directory, "outside"), file); err != nil {
		t.Skip(err)
	}
	store := openStore(t, config, nil, true)
	if err := store.Validate([]download.File{{Path: file}}); err == nil {
		t.Fatal("symbolic link accepted")
	}
}
