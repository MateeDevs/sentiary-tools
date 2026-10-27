package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MateeDevs/sentiary-tools/internal/download"
)

const integrationProtocol = 1

type integrationBinding struct {
	Output       string `json:"output"`
	Format       string `json:"format"`
	ResourceRoot string `json:"resourceRoot"`
}
type integrationRequest struct {
	ProtocolVersion int                  `json:"protocolVersion"`
	Bindings        []integrationBinding `json:"bindings"`
}

func readIntegration(path string, config download.Config) (*integrationRequest, []string, error) {
	if path == "" {
		return nil, nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var request integrationRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return nil, nil, fmt.Errorf("invalid integration request")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, nil, fmt.Errorf("invalid integration request trailer")
	}
	if request.ProtocolVersion != integrationProtocol {
		return nil, nil, fmt.Errorf("unsupported integration protocol; expected 1")
	}
	var selected []string
	seen := map[string]bool{}
	for _, binding := range request.Bindings {
		output, ok := config.Outputs[binding.Output]
		if !ok {
			return nil, nil, fmt.Errorf("unknown integration output %q", binding.Output)
		}
		if output.Format != binding.Format || output.Layout != download.PerLanguage {
			return nil, nil, fmt.Errorf("output %q: integration requires per_language format %q", binding.Output, binding.Format)
		}
		if !filepath.IsAbs(binding.ResourceRoot) {
			return nil, nil, fmt.Errorf("integration resourceRoot must be absolute")
		}
		if !seen[binding.Output] {
			selected = append(selected, binding.Output)
			seen[binding.Output] = true
		}
	}
	return &request, selected, nil
}

func (request *integrationRequest) validate(files []download.File) error {
	if request == nil {
		return nil
	}
	for _, binding := range request.Bindings {
		for _, file := range files {
			if file.Output != binding.Output {
				continue
			}
			relative, err := filepath.Rel(binding.ResourceRoot, file.Path)
			if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
				return fmt.Errorf("output %q: destination is outside integration resourceRoot", binding.Output)
			}
		}
	}
	return nil
}

type integrationFile struct {
	Output   string `json:"output"`
	Path     string `json:"path"`
	Checksum string `json:"sha256"`
}
type integrationResult struct {
	ProtocolVersion int               `json:"protocolVersion"`
	Files           []integrationFile `json:"files"`
}
