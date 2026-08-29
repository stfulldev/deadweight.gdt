package report

import (
	"fmt"
	"strings"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
)

// PresetList renders the built-in catalog in caller-provided product order.
func PresetList(result application.PresetListResult, _ Options) (string, error) {
	var output strings.Builder
	output.WriteString("Built-in presets (heuristic, experimental)\n\n")

	for _, item := range result.Catalog {
		fmt.Fprintf(&output, "%s\n", item.ID)
		fmt.Fprintf(&output, "  %s\n", item.Description)
		fmt.Fprintf(
			&output,
			"  Renderer: %s · Target: %d FPS · Quality: %s\n\n",
			displayRenderer(item.Renderer),
			item.TargetFPS,
			displayTitle(item.Quality),
		)
	}

	output.WriteString("Use `deadweight.gdt presets show <id>` to see budgets.\n")
	return output.String(), nil
}

// PresetShow renders metadata and budgets for one built-in preset.
func PresetShow(result application.PresetShowResult, _ Options) (string, error) {
	return renderPreset(result.Preset), nil
}

func renderPreset(item preset.Preset) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Preset:      %s\n", item.Name)
	fmt.Fprintf(&output, "ID:          %s\n", item.ID)
	fmt.Fprintf(&output, "Status:      %s\n", item.Status)
	fmt.Fprintf(&output, "Stability:   %s\n", item.Stability)
	fmt.Fprintf(&output, "Renderer:    %s\n", displayRenderer(item.Renderer))
	fmt.Fprintf(&output, "Target FPS:  %d\n", item.TargetFPS)
	fmt.Fprintf(&output, "Quality:     %s\n\n", displayTitle(item.Quality))
	output.WriteString("Budgets\n")

	for _, name := range metrics.OrderedNames() {
		limit, ok := item.Budgets.Get(name)
		if !ok {
			continue
		}
		fmt.Fprintf(&output, "  %-26s %8s\n", name.Label(), formatInteger(limit))
	}

	output.WriteString("\nThis preset is a starting guardrail, not a performance guarantee.\n")
	return output.String()
}
