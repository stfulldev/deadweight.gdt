package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
)

func newPresetsCommand(service Application) *cobra.Command {
	command := &cobra.Command{
		Use:   "presets",
		Short: "List built-in heuristic budget presets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := service.ListPresets()
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), renderPresetList(result.Catalog))
			return err
		},
	}

	command.AddCommand(&cobra.Command{
		Use:   "show <id>",
		Short: "Show one built-in preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := service.ShowPreset(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(cmd.OutOrStdout(), renderPreset(result.Preset))
			return err
		},
	})

	return command
}
func renderPresetList(catalog preset.Catalog) string {
	var output strings.Builder
	output.WriteString("Built-in presets (heuristic, experimental)\n\n")

	for _, item := range catalog {
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
	return output.String()
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

func displayRenderer(value string) string {
	switch value {
	case "forward_plus":
		return "Forward+"
	case "mobile":
		return "Mobile"
	case "compatibility":
		return "Compatibility"
	default:
		return displayTitle(value)
	}
}

func displayTitle(value string) string {
	if value == "" {
		return "Unspecified"
	}

	return strings.ToUpper(value[:1]) + strings.ReplaceAll(value[1:], "_", " ")
}

func formatInteger(value int64) string {
	raw := strconv.FormatInt(value, 10)
	if len(raw) <= 3 {
		return raw
	}

	first := len(raw) % 3
	if first == 0 {
		first = 3
	}

	var output strings.Builder
	output.WriteString(raw[:first])
	for index := first; index < len(raw); index += 3 {
		output.WriteByte(',')
		output.WriteString(raw[index : index+3])
	}

	return output.String()
}
