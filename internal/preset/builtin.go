package preset

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/*.json
var builtinFiles embed.FS

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

	result := make(Catalog, len(builtins))
	for index, item := range builtins {
		result[index] = item
		result[index].Budgets = item.Budgets.Clone()
	}

	return result, nil
}

func loadBuiltins() {
	ids := []string{"mobile", "steam-deck", "desktop"}
	builtins = make(Catalog, 0, len(ids))

	for _, id := range ids {
		path := fmt.Sprintf("data/%s.json", id)
		data, err := builtinFiles.ReadFile(path)
		if err != nil {
			builtinErr = fmt.Errorf("read built-in preset %q: %w", id, err)
			return
		}

		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()

		var item Preset
		if err := decoder.Decode(&item); err != nil {
			builtinErr = fmt.Errorf("decode built-in preset %q: %w", id, err)
			return
		}

		if item.ID != id {
			builtinErr = fmt.Errorf("built-in preset file %q contains id %q", id, item.ID)
			return
		}
		if item.Status != "heuristic" || item.Stability != "experimental" {
			builtinErr = fmt.Errorf("built-in preset %q must be heuristic and experimental", id)
			return
		}
		if item.TargetFPS <= 0 {
			builtinErr = fmt.Errorf("built-in preset %q has invalid target_fps", id)
			return
		}
		if item.Budgets.Count() != 8 {
			builtinErr = fmt.Errorf("built-in preset %q must define all eight budgets", id)
			return
		}

		builtins = append(builtins, item)
	}
}
