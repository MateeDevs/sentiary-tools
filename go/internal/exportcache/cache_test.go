package exportcache

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

func testSet(selection, content string) download.Snapshot {
	request := download.Request{ProjectID: "p", Format: "xlsx", Layout: download.PerLanguage, Languages: []string{"en"}}
	data := []byte(content)
	return download.Snapshot{Version: 1, Selection: download.Digest([]byte(selection)), CompletedAt: time.Now().UTC(),
		Projects: map[string]download.State{"p": {ProtocolVersion: 1, ProjectGeneration: "g", DataRevision: "1", ExporterVersion: "v", Languages: []string{"en"}}},
		Exports:  map[string]download.CachedExport{request.ID(): {Request: request, Checksum: download.Digest(data), Content: data, ETag: `W/"tag"`, Valid: true}}}
}
func TestCacheNormalizesServerAndKeepsBinaryContent(t *testing.T) {
	directory := t.TempDir()
	first, _ := Open(directory, "https://EXAMPLE.test:443/")
	second, _ := Open(directory, "https://example.test")
	set := testSet("selection", string([]byte{0, 255, 254, 10}))
	if err := first.Publish(set); err != nil {
		t.Fatal(err)
	}
	sets, err := second.Read()
	if err != nil || len(sets) != 1 {
		t.Fatalf("read cache: %v", err)
	}
	for key, export := range sets[0].Exports {
		if !export.Valid || !bytes.Equal(export.Content, set.Exports[key].Content) {
			t.Fatal("binary object changed")
		}
	}
}
func TestPruningKeepsObjectsReferencedByOtherSelections(t *testing.T) {
	cache, _ := Open(t.TempDir(), "https://example.test")
	a := testSet("a", "old")
	b := testSet("b", "old")
	replacement := testSet("a", "new")
	for _, set := range []download.Snapshot{a, b, replacement} {
		if err := cache.Publish(set); err != nil {
			t.Fatal(err)
		}
	}
	sets, err := cache.Read()
	if err != nil || len(sets) != 2 {
		t.Fatalf("sets: %v", err)
	}
	for _, set := range sets {
		for _, export := range set.Exports {
			if !export.Valid {
				t.Fatal("pruned referenced content")
			}
		}
	}
	if err := cache.Publish(testSet("b", "new")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cache.directory, "objects", download.Digest([]byte("old")))); !os.IsNotExist(err) {
		t.Fatal("unreferenced old content was not pruned")
	}
}
func TestFailedPublicationPreservesPreviousSet(t *testing.T) {
	cache, _ := Open(t.TempDir(), "https://example.test")
	old := testSet("selection", "old")
	if err := cache.Publish(old); err != nil {
		t.Fatal(err)
	}
	// An invalid object must fail before replacing the completed metadata.
	invalid := testSet("selection", "new")
	for key, export := range invalid.Exports {
		export.Checksum = "bad"
		invalid.Exports[key] = export
	}
	if err := cache.Publish(invalid); err == nil {
		t.Fatal("published invalid set")
	}
	sets, err := cache.Read()
	if err != nil || len(sets) != 1 || !sets[0].CompletedAt.Equal(old.CompletedAt) {
		t.Fatal("failed publication replaced set")
	}
}
func TestMissingEmptyObjectIsNotValid(t *testing.T) {
	cache, _ := Open(t.TempDir(), "https://example.test")
	if err := cache.Publish(testSet("selection", "")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(cache.directory, "objects", download.Digest(nil))); err != nil {
		t.Fatal(err)
	}
	sets, err := cache.Read()
	if err != nil {
		t.Fatal(err)
	}
	for _, export := range sets[0].Exports {
		if export.Valid {
			t.Fatal("missing empty object accepted")
		}
	}
}
func TestServersAndProjectsHaveDifferentIdentities(t *testing.T) {
	directory := t.TempDir()
	first, _ := Open(directory, "https://one.test")
	second, _ := Open(directory, "https://two.test")
	if err := first.Publish(testSet("selection", "secret")); err != nil {
		t.Fatal(err)
	}
	sets, err := second.Read()
	if err != nil || len(sets) != 0 {
		t.Fatal("server identities collided")
	}
	request := download.Request{ProjectID: "one"}
	other := request
	other.ProjectID = "two"
	if request.ID() == other.ID() {
		t.Fatal("project identities collided")
	}
}
func TestConcurrentProcessesPublishCompleteSelections(t *testing.T) {
	if directory := os.Getenv("SENTIARY_CACHE_TEST_CHILD"); directory != "" {
		cache, _ := Open(directory, "https://example.test")
		if err := cache.Publish(testSet(os.Getenv("SENTIARY_CACHE_TEST_SET"), "shared")); err != nil {
			t.Fatal(err)
		}
		return
	}
	directory := t.TempDir()
	var commands []*exec.Cmd
	for i := 0; i < 4; i++ {
		command := exec.Command(os.Args[0], "-test.run=^TestConcurrentProcessesPublishCompleteSelections$")
		command.Env = append(os.Environ(), "SENTIARY_CACHE_TEST_CHILD="+directory, "SENTIARY_CACHE_TEST_SET="+strconv.Itoa(i))
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		commands = append(commands, command)
	}
	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatal(err)
		}
	}
	cache, _ := Open(directory, "https://example.test")
	sets, err := cache.Read()
	if err != nil || len(sets) != 4 {
		t.Fatal(fmt.Sprintf("concurrent publication lost sets: %d, %v", len(sets), err))
	}
	for _, set := range sets {
		for _, export := range set.Exports {
			if !export.Valid {
				t.Fatal("concurrent prune removed content")
			}
		}
	}
}

func TestExplicitCacheDirectoryOverridesEnvironment(t *testing.T) {
	environment := t.TempDir()
	explicit := t.TempDir()
	t.Setenv("SENTIARY_CACHE_DIR", environment)
	cache, err := Open(explicit, "https://example.test")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(cache.directory) != explicit {
		t.Fatal("explicit cache directory did not win")
	}
	implicit, err := Open("", "https://example.test")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(implicit.directory) != environment {
		t.Fatal("environment cache directory ignored")
	}
}

func TestObjectWriteFailureCannotReplaceCompletedMetadata(t *testing.T) {
	cache, _ := Open(t.TempDir(), "https://example.test")
	old := testSet("selection", "old")
	if err := cache.Publish(old); err != nil {
		t.Fatal(err)
	}
	next := testSet("selection", "new")
	for _, export := range next.Exports {
		if err := os.Mkdir(filepath.Join(cache.directory, "objects", export.Checksum), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := cache.Publish(next); err == nil {
		t.Fatal("cache write should fail")
	}
	sets, err := cache.Read()
	if err != nil || len(sets) != 1 || !sets[0].CompletedAt.Equal(old.CompletedAt) {
		t.Fatal("object failure replaced completed metadata")
	}
}
