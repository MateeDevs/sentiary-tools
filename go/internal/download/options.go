package download

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// These limits are part of the resource-export-v1 wire contract.
const (
	ResourceExportCapability = "resource-export-v1"
	MaxExportListItems       = 10000
	MaxExportItemBytes       = 1024
)

// ExportOptions contains evaluated values, never local list paths.
// A non-nil empty Keys list is an explicit empty selection.
type ExportOptions struct {
	Keys         *[]string `json:"keys,omitempty"`
	Contexts     []string  `json:"contexts,omitempty"`
	RequiredKeys []string  `json:"requiredKeys,omitempty"`
	PluralBases  []string  `json:"pluralBases,omitempty"`
}

func (o ExportOptions) IsDefault() bool {
	return o.Keys == nil && len(o.Contexts) == 0 && len(o.RequiredKeys) == 0 && len(o.PluralBases) == 0
}

func (o ExportOptions) Validate(format string) error {
	lists := [][]string{o.Contexts, o.RequiredKeys, o.PluralBases}
	if o.Keys != nil {
		lists = append(lists, *o.Keys)
	}
	for _, values := range lists {
		if err := ValidateExportList(values); err != nil {
			return err
		}
	}
	if len(o.PluralBases) > 0 && format != "android" && format != "compose" {
		return fmt.Errorf("plural export requires android or compose format")
	}
	return nil
}

func ValidateExportList(values []string) error {
	if len(values) > MaxExportListItems {
		return fmt.Errorf("export list exceeds %d items", MaxExportListItems)
	}
	seen := map[string]bool{}
	for _, value := range values {
		if !utf8.ValidString(value) || strings.TrimSpace(value) == "" || len(value) > MaxExportItemBytes || strings.ContainsRune(value, 0) {
			return fmt.Errorf("invalid export list item (maximum %d UTF-8 bytes)", MaxExportItemBytes)
		}
		if seen[value] {
			return fmt.Errorf("duplicate export list item %q", value)
		}
		seen[value] = true
	}
	return nil
}

// Canonical returns detached values sorted only where order has no meaning.
func (o ExportOptions) Canonical() *ExportOptions {
	if o.IsDefault() {
		return nil
	}
	result := o
	if o.Keys != nil {
		keys := slices.Sorted(slices.Values(*o.Keys))
		if keys == nil {
			keys = []string{}
		}
		result.Keys = &keys
	}
	result.Contexts = slices.Sorted(slices.Values(o.Contexts))
	result.RequiredKeys = slices.Sorted(slices.Values(o.RequiredKeys))
	result.PluralBases = slices.Sorted(slices.Values(o.PluralBases))
	return &result
}
