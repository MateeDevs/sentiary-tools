// Package updates owns official release checking and installer-owned replacement.
package updates

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/minio/selfupdate"
)

const repository = "https://github.com/MateeDevs/sentiary-tools"
const latestRelease = "https://api.github.com/repos/MateeDevs/sentiary-tools/releases/latest"
const ownershipSuffix = ".sentiary-owned"
const maxAssetBytes = 100 << 20

var stableVersion = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// Newer accepts only stable release versions, never development or prerelease builds.
func Newer(candidate, current string) bool {
	a, b := stableVersion.FindStringSubmatch(candidate), stableVersion.FindStringSubmatch(current)
	if a == nil || b == nil {
		return false
	}
	for index := 1; index <= 3; index++ {
		x, errX := strconv.ParseUint(a[index], 10, 64)
		y, errY := strconv.ParseUint(b[index], 10, 64)
		if errX != nil || errY != nil {
			return false
		}
		if x != y {
			return x > y
		}
	}
	return false
}

type Releases struct {
	Client *http.Client
}

func (releases Releases) fetch(ctx context.Context, url string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "sentiary-cli-updater")
	response, err := releases.Client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("release download failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release download returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, fmt.Errorf("cannot read release download within size limit")
	}
	return data, nil
}

func (releases Releases) Latest(ctx context.Context) (string, error) {
	data, err := releases.fetch(ctx, latestRelease, 1<<20)
	if err != nil {
		return "", err
	}
	var release struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if err := json.Unmarshal(data, &release); err != nil || release.Draft || release.Prerelease || !stableVersion.MatchString(release.Tag) {
		return "", fmt.Errorf("latest release is not a stable official release")
	}
	return release.Tag, nil
}

func Owned(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	markerInfo, err := os.Lstat(path + ownershipSuffix)
	if err != nil || !markerInfo.Mode().IsRegular() || markerInfo.Size() > 128 {
		return false
	}
	marker, err := os.ReadFile(path + ownershipSuffix)
	if err != nil {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}
	return strings.TrimSpace(string(marker)) == hex.EncodeToString(hash.Sum(nil))
}

func ReinstallInstructions() string {
	if runtime.GOOS == "windows" {
		return "Reinstall with: & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.ps1).Content))"
	}
	return "Reinstall with: curl -fsSL https://raw.githubusercontent.com/MateeDevs/sentiary-tools/main/go/install.sh | sh"
}

func (releases Releases) Install(ctx context.Context, path, version string) error {
	if !Owned(path) {
		return fmt.Errorf("installation ownership could not be verified. %s", ReinstallInstructions())
	}
	if !stableVersion.MatchString(version) {
		return fmt.Errorf("only stable official releases may be installed")
	}
	if (runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows") || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		return fmt.Errorf("unsupported release platform")
	}
	base := "sentiary-cli-" + runtime.GOOS + "-" + runtime.GOARCH
	asset := base + ".tar.gz"
	if runtime.GOOS == "windows" {
		asset = base + ".exe"
	}
	url := repository + "/releases/download/" + version + "/"
	manifest, err := releases.fetch(ctx, url+"SHA256SUMS", 1<<20)
	if err != nil {
		return err
	}
	expected, err := checksum(manifest, asset)
	if err != nil {
		return err
	}
	data, err := releases.fetch(ctx, url+asset, maxAssetBytes)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(data)
	if !bytes.Equal(hash[:], expected) {
		return fmt.Errorf("release checksum mismatch")
	}
	binary := data
	if runtime.GOOS != "windows" {
		binary, err = extractBinary(data, base)
		if err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Check ownership again after network preparation, before touching the installation.
	if !Owned(path) {
		return fmt.Errorf("installation ownership changed during update")
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	binaryHash := sha256.Sum256(binary)
	if err := selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{
		TargetPath: path, TargetMode: info.Mode().Perm(), Checksum: binaryHash[:],
	}); err != nil {
		if rollback := selfupdate.RollbackError(err); rollback != nil {
			backup := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".old")
			return fmt.Errorf("replacement and rollback failed; recover backup %s or reinstall: %w", backup, rollback)
		}
		return fmt.Errorf("cannot replace installed binary: %w", err)
	}
	if err := os.WriteFile(path+ownershipSuffix, []byte(hex.EncodeToString(binaryHash[:])+"\n"), 0600); err != nil {
		return fmt.Errorf("updated binary, but cannot refresh ownership record; reinstall before another update")
	}
	return nil
}

func checksum(manifest []byte, asset string) ([]byte, error) {
	var found []byte
	for _, line := range strings.Split(string(manifest), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != asset {
			continue
		}
		value, err := hex.DecodeString(fields[0])
		if err != nil || len(value) != sha256.Size || found != nil {
			return nil, fmt.Errorf("invalid or ambiguous release checksum")
		}
		found = value
	}
	if found == nil {
		return nil, fmt.Errorf("release checksum is missing")
	}
	return found, nil
}

func extractBinary(data []byte, name string) ([]byte, error) {
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid release archive")
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	var binary []byte
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid release archive")
		}
		if header.Name != name && header.Name != "sentiary-cli" {
			return nil, fmt.Errorf("unexpected release archive entry")
		}
		if header.Typeflag != tar.TypeReg || header.Size <= 0 || header.Size > maxAssetBytes || binary != nil {
			return nil, fmt.Errorf("invalid release binary entry")
		}
		binary, err = io.ReadAll(io.LimitReader(archive, maxAssetBytes+1))
		if err != nil || int64(len(binary)) != header.Size {
			return nil, fmt.Errorf("cannot read release binary")
		}
	}
	if binary == nil {
		return nil, fmt.Errorf("release binary is missing")
	}
	return binary, nil
}
