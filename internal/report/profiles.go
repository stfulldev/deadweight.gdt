package report

import (
	"fmt"
	"strconv"
	"strings"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
)

// ProfileList renders custom profiles in their caller-provided canonical order.
func ProfileList(result application.ProfileListResult, _ Options) (string, error) {
	if err := validateProfileContext(result.Project.Directory, result.ConfigSource.Path); err != nil {
		return "", err
	}
	if err := validateProfileSummaries(result.Profiles); err != nil {
		return "", err
	}

	var output strings.Builder
	output.WriteString("Custom profiles\n")
	fmt.Fprintf(&output, "Configuration: %s\n\n", profileConfigurationPath(result.Project.Directory, result.ConfigSource.Path))
	if len(result.Profiles) == 0 {
		output.WriteString("No custom profiles are declared.\n")
		return output.String(), nil
	}
	for _, item := range result.Profiles {
		fmt.Fprintf(&output, "%s\n", item.ID)
		if item.Name != "" {
			fmt.Fprintf(&output, "  Name: %s\n", item.Name)
		}
		if item.Description != "" {
			fmt.Fprintf(&output, "  %s\n", item.Description)
		}
		if item.Extends != "" {
			fmt.Fprintf(&output, "  Extends: %s\n", item.Extends)
		}
		output.WriteByte('\n')
	}
	output.WriteString("Use `deadweight.gdt profiles show <id>` to inspect effective values.\n")
	return output.String(), nil
}

// ProfileShow renders one effective custom profile and every value source.
func ProfileShow(result application.ProfileShowResult, _ Options) (string, error) {
	if err := validateProfileContext(result.Project.Directory, result.ConfigSource.Path); err != nil {
		return "", err
	}
	if err := validateExplanation(result.Explanation); err != nil {
		return "", err
	}

	explanation := result.Explanation
	metadata := explanation.Effective.Metadata
	sources := explanation.MetadataSources
	var output strings.Builder
	fmt.Fprintf(&output, "Profile:       %s\n", explanation.Effective.ID)
	fmt.Fprintf(&output, "Configuration: %s\n\n", profileConfigurationPath(result.Project.Directory, result.ConfigSource.Path))
	output.WriteString("Inheritance\n")
	for _, layer := range explanation.Chain {
		fmt.Fprintf(&output, "  %s\n", displayLayer(layer))
	}

	output.WriteString("\nMetadata\n")
	writeProfileString(&output, "name", metadata.Name, sources.Name)
	writeProfileString(&output, "description", metadata.Description, sources.Description)
	writeProfileString(&output, "platform", metadata.Platform, sources.Platform)
	writeProfileString(&output, "renderer", metadata.Renderer, sources.Renderer)
	writeProfileInteger(&output, "target_fps", metadata.TargetFPS, sources.TargetFPS)
	writeProfileString(&output, "quality", metadata.Quality, sources.Quality)
	writeProfileString(&output, "status", metadata.Status, sources.Status)
	writeProfileString(&output, "stability", metadata.Stability, sources.Stability)

	output.WriteString("\nBudgets\n")
	for _, name := range metrics.OrderedNames() {
		limit, present := explanation.Effective.Budgets.Get(name)
		if !present {
			continue
		}
		source, _ := explanation.BudgetSources.Get(name)
		fmt.Fprintf(&output, "  %-26s %10s  [%s]\n", string(name), formatInteger(limit), displayLayer(source))
	}

	output.WriteString("\nPartial policy\n")
	fmt.Fprintf(
		&output,
		"  %-26s %10t  [%s]\n",
		"fail_on_partial",
		explanation.FailOnPartial,
		displayLayer(explanation.FailOnPartialSource),
	)
	return output.String(), nil
}

func writeProfileString(output *strings.Builder, name, value string, source policy.Layer) {
	if value == "" {
		value = "<empty>"
	}
	fmt.Fprintf(output, "  %-26s %10s  [%s]\n", name, value, displayLayer(source))
}

func writeProfileInteger(output *strings.Builder, name string, value int64, source policy.Layer) {
	fmt.Fprintf(output, "  %-26s %10s  [%s]\n", name, formatInteger(value), displayLayer(source))
}

func displayLayer(layer policy.Layer) string {
	if layer.ID == "" {
		return string(layer.Kind)
	}
	return string(layer.Kind) + ":" + layer.ID
}

func profileConfigurationPath(projectRoot, configPath string) string {
	if portable, ok := portableProjectPath(projectRoot, configPath); ok {
		return portable
	}
	return "<external>"
}

func validateProfileContext(projectRoot, configPath string) error {
	if strings.TrimSpace(projectRoot) == "" {
		return fmt.Errorf("profile result requires a project root")
	}
	if strings.TrimSpace(configPath) == "" {
		return fmt.Errorf("profile result requires a configuration source")
	}
	return nil
}

func validateProfileSummaries(profiles []policy.ProfileSummary) error {
	previous := ""
	for index, profile := range profiles {
		if profile.ID == "" {
			return fmt.Errorf("profile summary %d has an empty ID", index)
		}
		if previous != "" && profile.ID <= previous {
			return fmt.Errorf("profile summaries are not in canonical ID order")
		}
		previous = profile.ID
	}
	return nil
}

func validateExplanation(explanation policy.Explanation) error {
	if explanation.Effective.Kind != policy.KindProfile || explanation.Effective.ID == "" {
		return fmt.Errorf("profile explanation requires a selected custom profile")
	}
	if explanation.Effective.Metadata.TargetFPS < 0 {
		return fmt.Errorf("profile target_fps must be non-negative")
	}
	for _, layer := range explanation.Chain {
		if err := validateLayer(layer); err != nil {
			return err
		}
	}
	metadataSources := []policy.Layer{
		explanation.MetadataSources.Name,
		explanation.MetadataSources.Description,
		explanation.MetadataSources.Platform,
		explanation.MetadataSources.Renderer,
		explanation.MetadataSources.TargetFPS,
		explanation.MetadataSources.Quality,
		explanation.MetadataSources.Status,
		explanation.MetadataSources.Stability,
		explanation.FailOnPartialSource,
	}
	for _, source := range metadataSources {
		if err := validateLayer(source); err != nil {
			return err
		}
	}
	for _, name := range metrics.OrderedNames() {
		limit, present := explanation.Effective.Budgets.Get(name)
		source, sourced := explanation.BudgetSources.Get(name)
		if present != sourced {
			return fmt.Errorf("budget %s value/source presence mismatch", name)
		}
		if present {
			if limit < 0 {
				return fmt.Errorf("budget %s must be non-negative, got %s", name, strconv.FormatInt(limit, 10))
			}
			if err := validateLayer(source); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLayer(layer policy.Layer) error {
	if !layer.Kind.Valid() {
		return fmt.Errorf("invalid policy source layer %q", layer.Kind)
	}
	if (layer.Kind == policy.LayerPreset || layer.Kind == policy.LayerProfile) != (layer.ID != "") {
		return fmt.Errorf("policy source layer %q has invalid ID %q", layer.Kind, layer.ID)
	}
	return nil
}
