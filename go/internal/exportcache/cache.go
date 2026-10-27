// Package exportcache stores immutable export objects and atomically published download sets.
package exportcache

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gofrs/flock"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

type Cache struct{ directory string }

// Open scopes cache data to a normalized server. It does not retain credentials.
func Open(directory, server string) (*Cache, error) {
	if directory == "" {
		directory = os.Getenv("SENTIARY_CACHE_DIR")
	}
	if directory == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		directory = filepath.Join(base, "sentiary", "exports-v1")
	}
	identity, err := NormalizeServer(server)
	if err != nil {
		return nil, err
	}
	return &Cache{filepath.Join(directory, download.Digest([]byte(identity)))}, nil
}
func NormalizeServer(server string) (string, error) {
	u, err := url.Parse(server)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("invalid cache server URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		u.Host = strings.TrimSuffix(u.Host, ":"+u.Port())
	}
	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}
func (c *Cache) locked(operation func() error) (err error) {
	if err = os.MkdirAll(c.directory, 0o700); err != nil {
		return err
	}
	lock := flock.New(filepath.Join(c.directory, "cache.lock"))
	if err = lock.Lock(); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	return operation()
}
func validHash(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && value == strings.ToLower(value)
}
func (c *Cache) metadata() ([]download.Snapshot, error) {
	paths, err := filepath.Glob(filepath.Join(c.directory, "sets", "*.json"))
	if err != nil {
		return nil, err
	}
	var result []download.Snapshot
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var snapshot download.Snapshot
		if err = json.Unmarshal(content, &snapshot); err != nil {
			return nil, fmt.Errorf("corrupt cache metadata: %w", err)
		}
		if snapshot.Version != 1 || !validHash(snapshot.Selection) || snapshot.CompletedAt.IsZero() {
			return nil, fmt.Errorf("invalid cache metadata")
		}
		for _, state := range snapshot.Projects {
			if err := state.Validate(); err != nil {
				return nil, fmt.Errorf("invalid cache project state: %w", err)
			}
		}
		for key, export := range snapshot.Exports {
			if key != export.Request.ID() || !validHash(export.Checksum) {
				return nil, fmt.Errorf("invalid cache export reference")
			}
		}
		result = append(result, snapshot)
	}
	slices.SortFunc(result, func(a, b download.Snapshot) int { return b.CompletedAt.Compare(a.CompletedAt) })
	return result, nil
}
func (c *Cache) Read() (snapshots []download.Snapshot, err error) {
	err = c.locked(func() error {
		var err error
		snapshots, err = c.metadata()
		if err != nil {
			return err
		}
		for i := range snapshots {
			for key, export := range snapshots[i].Exports {
				content, err := os.ReadFile(filepath.Join(c.directory, "objects", export.Checksum))
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				if err == nil && download.Digest(content) == export.Checksum {
					export.Content = content
					export.Valid = true
				}
				snapshots[i].Exports[key] = export
			}
		}
		return nil
	})
	return
}

func (c *Cache) Publish(snapshot download.Snapshot) error {
	return c.locked(func() error {
		previous, err := c.metadata()
		if err != nil {
			return err
		}
		if snapshot.Version != 1 || !validHash(snapshot.Selection) || snapshot.CompletedAt.IsZero() {
			return fmt.Errorf("invalid download set")
		}
		for key, export := range snapshot.Exports {
			if key != export.Request.ID() || !export.Valid || export.Checksum != download.Digest(export.Content) {
				return fmt.Errorf("invalid export content")
			}
			// Atomic replacement also repairs an object corrupted since a prior publication.
			if err := atomicWrite(filepath.Join(c.directory, "objects", export.Checksum), export.Content); err != nil {
				return err
			}
		}
		content, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(c.directory, "sets", snapshot.Selection+".json"), content); err != nil {
			return err
		}
		referenced := map[string]bool{}
		for _, set := range append(previous, snapshot) {
			if set.Selection == snapshot.Selection && set.CompletedAt != snapshot.CompletedAt {
				continue
			}
			for _, export := range set.Exports {
				referenced[export.Checksum] = true
			}
		}
		// Publication has succeeded. Cleanup is best effort; leftover immutable objects are safe.
		paths, _ := filepath.Glob(filepath.Join(c.directory, "objects", "*"))
		for _, path := range paths {
			if !referenced[filepath.Base(path)] {
				_ = os.Remove(path)
			}
		}
		return nil
	})
}
func atomicWrite(path string, content []byte) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".pending-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(content); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
