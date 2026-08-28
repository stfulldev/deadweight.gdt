package preset

import "github.com/stfulldev/deadweight.gdt/internal/budget"

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

// Find returns a preset by its stable ID.
func (catalog Catalog) Find(id string) (Preset, bool) {
	for _, item := range catalog {
		if item.ID == id {
			return item, true
		}
	}

	return Preset{}, false
}
