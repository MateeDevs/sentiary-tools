// Package fileoutput manages downloaded files without taking ownership of unrelated files.
package fileoutput

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/flock"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

type manifest map[string]map[string]string

type Store struct {
	root     string
	name     string
	path     string
	lock     *flock.Flock
	force    bool
	selected []string
	previous manifest
}

// Open locks the selected config's ownership manifest until Close is called.
func Open(config download.Config, selected []string, force bool) (*Store, error) {
	return OpenContext(context.Background(), config, selected, force)
}

func OpenContext(ctx context.Context, config download.Config, selected []string, force bool) (*Store, error) {
	directory := filepath.Join(config.Directory, ".sentiary")
	if info, err := os.Lstat(directory); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf(".sentiary must not be a symbolic link")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	store := &Store{root: config.Directory, name: config.Name, path: filepath.Join(directory, config.Name+".json"),
		lock: flock.New(filepath.Join(directory, config.Name+".lock")), force: force, selected: selected, previous: manifest{}}
	locked, err := store.lock.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("wait for download lock: %w", err)
	}
	if !locked {
		return nil, ctx.Err()
	}
	data, err := os.ReadFile(store.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		store.Close()
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &store.previous); err != nil {
			store.Close()
			return nil, fmt.Errorf("invalid output manifest")
		}
		if store.previous == nil {
			store.Close()
			return nil, fmt.Errorf("invalid output manifest")
		}
	}
	return store, nil
}

func (store *Store) Close() error { return store.lock.Close() }

func (store *Store) selectedOutput(name string) bool {
	return len(store.selected) == 0 || slices.Contains(store.selected, name)
}

func (store *Store) managed() (map[string]string, error) {
	result := map[string]string{}
	for _, paths := range store.previous {
		for path, digest := range paths {
			if !filepath.IsAbs(path) || len(digest) != sha256.Size*2 {
				return nil, fmt.Errorf("invalid output manifest entry")
			}
			if _, err := hex.DecodeString(digest); err != nil {
				return nil, fmt.Errorf("invalid output manifest checksum")
			}
			if _, exists := result[path]; exists {
				return nil, fmt.Errorf("duplicate output manifest path")
			}
			result[path] = digest
		}
	}
	return result, nil
}

// Validate checks ownership before network requests and again before writes.
func (store *Store) Validate(files []download.File) error {
	managed, err := store.managed()
	if err != nil {
		return err
	}
	paths := map[string]bool{}
	destinations := map[string]string{}
	for _, file := range files {
		paths[file.Path] = true
		destinations[strings.ToLower(file.Path)] = file.Path
	}
	for name, previous := range store.previous {
		for path := range previous {
			if destination, exists := destinations[strings.ToLower(path)]; exists && destination != path {
				return fmt.Errorf("case-only changes to owned output paths are unsupported; use a distinct destination")
			}
			if store.selectedOutput(name) {
				paths[path] = true
			} else if paths[path] {
				return fmt.Errorf("destination belongs to an unselected output %q", name)
			}
		}
	}
	for path := range paths {
		if err := store.validatePath(path); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read destination: %w", err)
		}
		previous, owned := managed[path]
		if !store.force && (!owned || checksum(data) != previous) {
			return fmt.Errorf("destination %q is untracked or manually modified; use --force to take ownership or discard edits", path)
		}
	}
	return nil
}

func (store *Store) validatePath(path string) error {
	for _, name := range []string{store.name, "sentiary.toml", "sentiary.local.toml", ".sentiary"} {
		if strings.EqualFold(path, filepath.Join(store.root, name)) {
			return fmt.Errorf("output must not overwrite configuration or download state")
		}
	}
	if strings.HasPrefix(strings.ToLower(path), strings.ToLower(filepath.Join(store.root, ".sentiary"))+string(filepath.Separator)) {
		return fmt.Errorf("output must not overwrite configuration or download state")
	}
	// Do not follow links when inspecting or replacing managed output files.
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output path contains a symbolic link: %s", current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}

func (store *Store) Commit(files []download.DownloadedFile) error {
	next := manifest{}
	for name, paths := range store.previous {
		if !store.selectedOutput(name) {
			next[name] = paths
		}
	}
	staged := map[string]string{}
	defer func() {
		for _, temp := range staged {
			os.Remove(temp)
		}
	}()
	for _, file := range files {
		if next[file.Output] == nil {
			next[file.Output] = map[string]string{}
		}
		next[file.Output][file.Path] = checksum(file.Content)
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return err
		}
		temp, err := stage(file.Path, file.Content, 0o644)
		if err != nil {
			return err
		}
		staged[file.Path] = temp
	}
	for _, file := range files {
		if err := os.Rename(staged[file.Path], file.Path); err != nil {
			return fmt.Errorf("replace destination (some files may have changed; retry with --force): %w", err)
		}
	}
	active := map[string]bool{}
	for _, paths := range next {
		for path := range paths {
			active[path] = true
		}
	}
	for name, paths := range store.previous {
		if !store.selectedOutput(name) {
			continue
		}
		for path := range paths {
			if !active[path] {
				if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("remove obsolete output: %w", err)
				}
			}
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	temp, err := stage(store.path, data, 0o600)
	if err != nil {
		return err
	}
	defer os.Remove(temp)
	if err := os.Rename(temp, store.path); err != nil {
		return fmt.Errorf("save ownership manifest: %w", err)
	}
	store.previous = next
	return nil
}

func stage(path string, data []byte, mode os.FileMode) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".sentiary-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Chmod(mode); err != nil {
		file.Close()
		os.Remove(name)
		return "", err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(name)
		return "", errors.Join(writeErr, closeErr)
	}
	return name, nil
}

func checksum(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
