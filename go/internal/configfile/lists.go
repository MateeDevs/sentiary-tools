package configfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/MateeDevs/sentiary-tools/go/internal/download"
)

func readExportList(root string, filename *string) ([]string, error) {
	if filename == nil {
		return nil, nil
	}
	if strings.TrimSpace(*filename) == "" {
		return nil, fmt.Errorf("export list path must not be empty")
	}
	path := *filename
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Bound the file independently of line lengths and blank-line count.
	const maxFileBytes = download.MaxExportListItems * (download.MaxExportItemBytes + 2)
	limited := &io.LimitedReader{R: file, N: maxFileBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 4096), maxFileBytes+1)
	values := []string{}
	seen := map[string]int{}
	line := 0
	for scanner.Scan() {
		line++
		if !utf8.Valid(scanner.Bytes()) {
			return nil, fmt.Errorf("%s:%d: invalid UTF-8", path, line)
		}
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			continue
		}
		if previous, ok := seen[value]; ok {
			return nil, fmt.Errorf("%s:%d: duplicate key (first at line %d)", path, line, previous)
		}
		if err := download.ValidateExportList([]string{value}); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		seen[value] = line
		values = append(values, value)
		if len(values) > download.MaxExportListItems {
			return nil, fmt.Errorf("%s:%d: too many export list items", path, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s:%d: %w", path, line+1, err)
	}
	if limited.N == 0 {
		return nil, fmt.Errorf("%s: export list file is too large", path)
	}
	return values, nil
}
