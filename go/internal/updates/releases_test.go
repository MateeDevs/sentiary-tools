package updates

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStableVersionComparison(t *testing.T) {
	for _, test := range []struct {
		candidate, current string
		want               bool
	}{
		{"v2.0.0", "v1.9.9", true}, {"v1.10.0", "v1.9.9", true},
		{"1.0.1", "v1.0.0", true}, {"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v2.0.0", false}, {"v2.0.0-rc1", "v1.0.0", false},
		{"v2.0.0", "dev", false}, {"v2.0.0", "staging-abc", false},
	} {
		t.Run(test.candidate+"_"+test.current, func(t *testing.T) {
			if got := Newer(test.candidate, test.current); got != test.want {
				t.Fatalf("got %t", got)
			}
		})
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (transport transportFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func fixtureReleases(body func(string) []byte) Releases {
	return Releases{Client: &http.Client{Transport: transportFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body(request.URL.String()))), Header: make(http.Header)}, nil
	})}}
}

func TestLatestRejectsPrerelease(t *testing.T) {
	releases := fixtureReleases(func(string) []byte { return []byte(`{"tag_name":"v2.0.0","prerelease":true}`) })
	if _, err := releases.Latest(context.Background()); err == nil {
		t.Fatal("accepted prerelease")
	}
}

func TestChecksumRejectsAmbiguousEntries(t *testing.T) {
	line := strings.Repeat("0", 64) + "  asset\n"
	if _, err := checksum([]byte(line+line), "asset"); err == nil {
		t.Fatal("accepted ambiguous manifest")
	}
}

func archiveFixture(t *testing.T, name string, contents []byte) []byte {
	t.Helper()
	var data bytes.Buffer
	zipper := gzip.NewWriter(&data)
	archive := tar.NewWriter(zipper)
	if err := archive.WriteHeader(&tar.Header{Name: name, Size: int64(len(contents)), Mode: 0755}); err != nil {
		t.Fatal(err)
	}
	if _, err := archive.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipper.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func TestArchiveRejectsTraversalWithoutWriting(t *testing.T) {
	data := archiveFixture(t, "../sentiary-cli", []byte("unexpected"))
	if _, err := extractBinary(data, "sentiary-cli"); err == nil {
		t.Fatal("accepted traversal entry")
	}
}

func installFixture(t *testing.T) (string, []byte, string, []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sentiary-cli")
	old := []byte("old fixture executable")
	if err := os.WriteFile(path, old, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+ownershipSuffix, []byte(fmt.Sprintf("%x\n", sha256.Sum256(old))), 0600); err != nil {
		t.Fatal(err)
	}
	base := "sentiary-cli-" + runtime.GOOS + "-" + runtime.GOARCH
	newBinary := []byte("new fixture executable")
	asset := base + ".tar.gz"
	data := archiveFixture(t, base, newBinary)
	if runtime.GOOS == "windows" {
		asset, data = base+".exe", newBinary
	}
	return path, old, asset, data
}

func TestInstallChecksumFailurePreservesOldBinary(t *testing.T) {
	path, old, asset, data := installFixture(t)
	releases := fixtureReleases(func(url string) []byte {
		if strings.HasSuffix(url, "SHA256SUMS") {
			return []byte(strings.Repeat("0", 64) + "  " + asset + "\n")
		}
		return data
	})
	if err := releases.Install(context.Background(), path, "v2.0.0"); err == nil {
		t.Fatal("accepted bad checksum")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, old) || !Owned(path) {
		t.Fatal("failed preparation changed old installation")
	}
}

func TestInstallVerifiedReleaseReplacesOnlyFixture(t *testing.T) {
	path, _, asset, data := installFixture(t)
	releases := fixtureReleases(func(url string) []byte {
		if strings.HasSuffix(url, "SHA256SUMS") {
			return []byte(fmt.Sprintf("%x  %s\n", sha256.Sum256(data), asset))
		}
		return data
	})
	if err := releases.Install(context.Background(), path, "v2.0.0"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "new fixture executable" || !Owned(path) {
		t.Fatal("incorrect replacement or ownership")
	}
}

func TestInstallRejectsChangedOwnershipWithoutNetwork(t *testing.T) {
	path, _, _, _ := installFixture(t)
	if err := os.WriteFile(path, []byte("externally replaced"), 0755); err != nil {
		t.Fatal(err)
	}
	releases := fixtureReleases(func(string) []byte { t.Fatal("unowned install accessed network"); return nil })
	if err := releases.Install(context.Background(), path, "v2.0.0"); err == nil {
		t.Fatal("accepted changed ownership")
	}
}
