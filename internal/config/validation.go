package config

import (
	"fmt"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// Validate applies only static version-one declaration rules. It deliberately
// does not resolve selector or inheritance references.
func Validate(configuration Config, source string) error {
	if configuration.Version != CurrentVersion {
		return validationError(
			source,
			"version",
			fmt.Sprintf("must be %d, got %d", CurrentVersion, configuration.Version),
		)
	}
	if configuration.Preset != nil && configuration.Profile != nil {
		return validationError(source, "preset/profile", "selectors are mutually exclusive")
	}
	if err := validateOptionalID(configuration.Preset, source, "preset"); err != nil {
		return err
	}
	if err := validateOptionalID(configuration.Profile, source, "profile"); err != nil {
		return err
	}
	if err := validateBudgets(configuration.Budgets, source, "budgets"); err != nil {
		return err
	}

	ids := make([]string, 0, len(configuration.Profiles))
	for id := range configuration.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		prefix := joinField("profiles", id)
		if !ValidID(id) {
			return validationError(source, prefix, fmt.Sprintf("id %q does not match %s", id, IDPattern))
		}
		profile := configuration.Profiles[id]
		if err := validateProfile(profile, source, prefix); err != nil {
			return err
		}
	}

	return nil
}

func validateProfile(profile Profile, source, prefix string) error {
	if err := validateOptionalID(profile.Extends, source, joinField(prefix, "extends")); err != nil {
		return err
	}
	if profile.Platform != nil && *profile.Platform == "" {
		return validationError(source, joinField(prefix, "platform"), "must not be empty")
	}
	if profile.Renderer != nil && !ValidRenderer(*profile.Renderer) {
		return validationError(
			source,
			joinField(prefix, "renderer"),
			fmt.Sprintf("unsupported renderer %q", *profile.Renderer),
		)
	}
	if profile.TargetFPS != nil && *profile.TargetFPS < 0 {
		return validationError(source, joinField(prefix, "target_fps"), "must be non-negative")
	}
	if profile.Quality != nil && !ValidQuality(*profile.Quality) {
		return validationError(
			source,
			joinField(prefix, "quality"),
			fmt.Sprintf("unsupported quality %q", *profile.Quality),
		)
	}

	return validateBudgets(profile.Budgets, source, joinField(prefix, "budgets"))
}

func validateOptionalID(value *string, source, field string) error {
	if value == nil {
		return nil
	}
	if !ValidID(*value) {
		return validationError(
			source,
			field,
			fmt.Sprintf("id %q does not match %s", *value, IDPattern),
		)
	}

	return nil
}

func validateBudgets(limits budget.Limits, source, prefix string) error {
	for _, name := range metrics.OrderedNames() {
		value, configured := limits.Get(name)
		if configured && value < 0 {
			return validationError(source, joinField(prefix, string(name)), "must be non-negative")
		}
	}

	return nil
}

func validationError(source, field, detail string) *Error {
	return configError(ReasonValidation, source, field, detail, nil)
}
