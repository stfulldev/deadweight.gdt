package preset

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

//go:embed data/*.json
var builtinFiles embed.FS

var builtinIDs = []string{"mobile", "steam-deck", "desktop"}

var (
	builtinOnce sync.Once
	builtins    Catalog
	builtinErr  error
)

// Builtins loads and validates the version-controlled built-in presets.
// The returned catalog is a defensive copy in product display order.
func Builtins() (Catalog, error) {
	builtinOnce.Do(loadBuiltins)
	if builtinErr != nil {
		return nil, builtinErr
	}

	return builtins.clone(), nil
}

func loadBuiltins() {
	builtins, builtinErr = loadCatalog(builtinFiles, builtinIDs)
}

func loadCatalog(files fs.FS, ids []string) (Catalog, error) {
	catalog := make(Catalog, 0, len(ids))
	seenIDs := make(map[string]struct{}, len(ids))

	for _, id := range ids {
		path := fmt.Sprintf("data/%s.json", id)
		data, err := fs.ReadFile(files, path)
		if err != nil {
			return nil, fmt.Errorf("read built-in preset %q: %w", id, err)
		}

		item, err := decodePreset(data, id)
		if err != nil {
			return nil, err
		}

		if _, duplicate := seenIDs[item.ID]; duplicate {
			return nil, fmt.Errorf("built-in preset %q has duplicate id", item.ID)
		}
		seenIDs[item.ID] = struct{}{}
		catalog = append(catalog, item)
	}

	return catalog, nil
}

func decodePreset(data []byte, expectedID string) (Preset, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var item Preset
	if err := decoder.Decode(&item); err != nil {
		return Preset{}, fmt.Errorf("decode built-in preset %q: %w", expectedID, err)
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Preset{}, fmt.Errorf("decode built-in preset %q: trailing JSON data", expectedID)
	}

	if err := validatePreset(item, expectedID); err != nil {
		return Preset{}, err
	}

	return item, nil
}

func validatePreset(item Preset, expectedID string) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "id", value: item.ID},
		{name: "name", value: item.Name},
		{name: "description", value: item.Description},
		{name: "platform", value: item.Platform},
		{name: "renderer", value: item.Renderer},
		{name: "quality", value: item.Quality},
		{name: "status", value: item.Status},
		{name: "stability", value: item.Stability},
	}
	for _, field := range required {
		if field.value == "" {
			return fmt.Errorf("built-in preset %q is missing required field %q", expectedID, field.name)
		}
	}

	if item.ID != expectedID {
		return fmt.Errorf("built-in preset %q contains id %q", expectedID, item.ID)
	}
	if item.TargetFPS <= 0 {
		return fmt.Errorf("built-in preset %q field %q must be greater than zero", expectedID, "target_fps")
	}
	if item.Status != "heuristic" {
		return fmt.Errorf("built-in preset %q field %q must be %q", expectedID, "status", "heuristic")
	}
	if item.Stability != "experimental" {
		return fmt.Errorf("built-in preset %q field %q must be %q", expectedID, "stability", "experimental")
	}
	if !validRenderer(item.Renderer) {
		return fmt.Errorf("built-in preset %q has invalid renderer %q", expectedID, item.Renderer)
	}
	if !validQuality(item.Quality) {
		return fmt.Errorf("built-in preset %q has invalid quality %q", expectedID, item.Quality)
	}

	for _, name := range metrics.OrderedNames() {
		limit, configured := item.Budgets.Get(name)
		if !configured {
			return fmt.Errorf("built-in preset %q is missing budget %q", expectedID, name)
		}
		if limit < 0 {
			return fmt.Errorf("built-in preset %q budget %q must be non-negative", expectedID, name)
		}
	}

	return nil
}

func validRenderer(value string) bool {
	switch value {
	case "forward_plus", "mobile", "compatibility", "unspecified":
		return true
	default:
		return false
	}
}

func validQuality(value string) bool {
	switch value {
	case "low", "balanced", "high", "custom":
		return true
	default:
		return false
	}
}
