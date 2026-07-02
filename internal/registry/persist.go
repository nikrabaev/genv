package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nikrabaev/genv/internal/core"
	genvio "github.com/nikrabaev/genv/internal/io"
)

const RegistryFilename = "genv.json"

// LoadRegistry reads genv.json from root, parses, and validates it.
func LoadRegistry(root string) (Registry, error) {
	path := filepath.Join(root, RegistryFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Registry{}, &core.GenvError{
				Code:    core.ErrNotFound,
				Message: fmt.Sprintf("no %s found in %s — run `genv init` first", RegistryFilename, root),
			}
		}
		return Registry{}, &core.GenvError{
			Code:    core.ErrVaultIO,
			Message: fmt.Sprintf("could not read %s: %v", RegistryFilename, err),
		}
	}

	var raw Registry
	if err := json.Unmarshal(data, &raw); err != nil {
		return Registry{}, &core.GenvError{
			Code:    core.ErrParse,
			Message: fmt.Sprintf("%s is not valid JSON: %v", RegistryFilename, err),
		}
	}

	issues := ValidateRegistry(raw)
	if len(issues) > 0 {
		parts := make([]string, len(issues))
		for i, iss := range issues {
			parts[i] = iss.String()
		}
		return Registry{}, &core.GenvError{
			Code:    core.ErrValidation,
			Message: fmt.Sprintf("%s is invalid — %s", RegistryFilename, strings.Join(parts, "; ")),
			Details: issues,
		}
	}

	return raw, nil
}

// SaveRegistry writes the registry to genv.json with canonical formatting:
// 2-space indent + trailing newline, so diffs stay minimal and byte-stable.
func SaveRegistry(root string, r Registry) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return &core.GenvError{
			Code:    core.ErrVaultIO,
			Message: fmt.Sprintf("could not serialize registry: %v", err),
		}
	}
	content := append(data, '\n')
	return genvio.WriteFileAtomic(root, RegistryFilename, content)
}
