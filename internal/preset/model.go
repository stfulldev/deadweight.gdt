package preset

import (
	"fmt"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
)

// Preset is a built-in heuristic budget profile.
type Preset struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Platform    string        `json:"platform"`
	Renderer    string        `json:"renderer"`
	TargetFPS   int           `json:"target_fps"`
	Quality     string        `json:"quality"`
	Status      string        `json:"status"`
	Stability   string        `json:"stability"`
	Budgets     budget.Limits `json:"budgets"`
}

// Catalog preserves product order and provides lookup by stable ID.
type Catalog []Preset

// Find returns an independent preset value by its stable ID.
func (catalog Catalog) Find(id string) (Preset, error) {
	for _, item := range catalog {
		if item.ID == id {
			return item.clone(), nil
		}
	}

	ids := make([]string, 0, len(catalog))
	for _, item := range catalog {
		ids = append(ids, item.ID)
	}

	return Preset{}, fmt.Errorf("unknown preset %q; available presets: %s", id, strings.Join(ids, ", "))
}

func (item Preset) clone() Preset {
	item.Budgets = item.Budgets.Clone()
	return item
}

func (catalog Catalog) clone() Catalog {
	result := make(Catalog, len(catalog))
	for index, item := range catalog {
		result[index] = item.clone()
	}
	return result
}
