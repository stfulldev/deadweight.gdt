// Package config discovers and strictly decodes version-one project policy.
package config

import "github.com/stfulldev/deadweight.gdt/internal/budget"

const (
	// CurrentVersion is the only configuration version supported by MVP 0.1.
	CurrentVersion = 1
	// DefaultFilename is the implicit project-local configuration name.
	DefaultFilename = ".deadweight.gdt.json"
)

// Config is one owned, statically validated version-one declaration set.
// Dynamic selector and inheritance resolution belongs to the policy layer.
type Config struct {
	Version       int
	Preset        *string
	Profile       *string
	FailOnPartial bool
	Budgets       budget.Limits
	Profiles      map[string]Profile
}

// Profile preserves optional custom-profile declarations for later merging.
type Profile struct {
	Name        *string
	Description *string
	Extends     *string
	Platform    *string
	Renderer    *string
	TargetFPS   *int64
	Quality     *string
	Budgets     budget.Limits
}

// Clone returns a deep copy whose maps and optional values do not alias input.
func (configuration Config) Clone() Config {
	cloned := configuration
	cloned.Preset = cloneString(configuration.Preset)
	cloned.Profile = cloneString(configuration.Profile)
	cloned.Budgets = configuration.Budgets.Clone()
	if configuration.Profiles != nil {
		cloned.Profiles = make(map[string]Profile, len(configuration.Profiles))
		for id, profile := range configuration.Profiles {
			cloned.Profiles[id] = profile.clone()
		}
	}

	return cloned
}

func (profile Profile) clone() Profile {
	cloned := profile
	cloned.Name = cloneString(profile.Name)
	cloned.Description = cloneString(profile.Description)
	cloned.Extends = cloneString(profile.Extends)
	cloned.Platform = cloneString(profile.Platform)
	cloned.Renderer = cloneString(profile.Renderer)
	cloned.TargetFPS = cloneInt64(profile.TargetFPS)
	cloned.Quality = cloneString(profile.Quality)
	cloned.Budgets = profile.Budgets.Clone()

	return cloned
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
