package policy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
)

const maxProfileDepth = 32

type visitState uint8

const (
	visitUnseen visitState = iota
	visitActive
	visitDone
)

type selectedBase struct {
	kind  Kind
	id    string
	field string
}

type resolvedProfile struct {
	metadata Metadata
	budgets  budget.Limits
}

func (resolved resolvedProfile) clone() resolvedProfile {
	resolved.budgets = resolved.budgets.Clone()
	return resolved
}

type graphResolver struct {
	source     string
	profiles   map[string]config.Profile
	builtins   map[string]preset.Preset
	builtinIDs []string
	profileIDs []string
	state      map[string]visitState
	memo       map[string]resolvedProfile
	stack      []string
	stackIndex map[string]int
}

// Resolve validates all custom profiles and resolves selectors and overrides
// into one owned effective policy. It returns the zero policy on every error.
func Resolve(
	source string,
	configuration config.Config,
	cli Selector,
	cliBudgetValues []string,
) (Effective, error) {
	configuration = configuration.Clone()
	catalog, err := preset.Builtins()
	if err != nil {
		return Effective{}, fmt.Errorf("load built-in presets: %w", err)
	}

	resolver, err := newGraphResolver(source, configuration.Profiles, catalog)
	if err != nil {
		return Effective{}, err
	}
	if err := resolver.resolveAll(); err != nil {
		return Effective{}, err
	}

	selected, err := selectBase(source, configuration, cli)
	if err != nil {
		return Effective{}, err
	}

	effective, err := resolver.resolveSelected(selected)
	if err != nil {
		return Effective{}, err
	}
	effective.Budgets = mergeLimits(effective.Budgets, configuration.Budgets)

	cliBudgets, err := ParseBudgetOverrides(source, cliBudgetValues)
	if err != nil {
		return Effective{}, err
	}
	effective.Budgets = mergeLimits(effective.Budgets, cliBudgets)
	if effective.Budgets.Count() == 0 {
		return Effective{}, policyError(
			source,
			"budgets",
			"effective policy has no budget; select a preset/profile or provide a config/CLI budget",
		)
	}

	return effective.Clone(), nil
}

func newGraphResolver(
	source string,
	profiles map[string]config.Profile,
	catalog preset.Catalog,
) (*graphResolver, error) {
	resolver := &graphResolver{
		source:     source,
		profiles:   profiles,
		builtins:   make(map[string]preset.Preset, len(catalog)),
		builtinIDs: make([]string, 0, len(catalog)),
		profileIDs: make([]string, 0, len(profiles)),
		state:      make(map[string]visitState, len(profiles)),
		memo:       make(map[string]resolvedProfile, len(profiles)),
		stackIndex: make(map[string]int, maxProfileDepth),
	}

	for _, item := range catalog {
		resolver.builtins[item.ID] = item
		resolver.builtinIDs = append(resolver.builtinIDs, item.ID)
	}
	for id := range profiles {
		resolver.profileIDs = append(resolver.profileIDs, id)
	}
	sort.Strings(resolver.profileIDs)

	for _, id := range resolver.profileIDs {
		if _, collision := resolver.builtins[id]; collision {
			return nil, policyError(
				source,
				"profiles."+id,
				fmt.Sprintf("custom profile %q collides with a built-in preset ID", id),
			)
		}
	}
	for _, id := range resolver.profileIDs {
		parent := profiles[id].Extends
		if parent == nil {
			continue
		}
		if _, builtIn := resolver.builtins[*parent]; builtIn {
			continue
		}
		if _, custom := profiles[*parent]; custom {
			continue
		}

		return nil, policyError(
			source,
			"profiles."+id+".extends",
			fmt.Sprintf("custom profile %q extends unknown parent %q", id, *parent),
		)
	}

	return resolver, nil
}

func (resolver *graphResolver) resolveAll() error {
	for _, id := range resolver.profileIDs {
		if _, err := resolver.resolveProfile(id); err != nil {
			return err
		}
	}

	return nil
}

func (resolver *graphResolver) resolveProfile(id string) (resolvedProfile, error) {
	switch resolver.state[id] {
	case visitDone:
		return resolver.memo[id].clone(), nil
	case visitActive:
		start := resolver.stackIndex[id]
		cycle := append(append([]string(nil), resolver.stack[start:]...), id)
		fieldID := resolver.stack[len(resolver.stack)-1]
		return resolvedProfile{}, policyError(
			resolver.source,
			"profiles."+fieldID+".extends",
			"profile inheritance cycle: "+strings.Join(cycle, " -> "),
		)
	}

	if len(resolver.stack) >= maxProfileDepth {
		chain := append(append([]string(nil), resolver.stack...), id)
		fieldID := resolver.stack[len(resolver.stack)-1]
		return resolvedProfile{}, policyError(
			resolver.source,
			"profiles."+fieldID+".extends",
			fmt.Sprintf(
				"profile inheritance depth exceeds %d: %s",
				maxProfileDepth,
				strings.Join(chain, " -> "),
			),
		)
	}

	resolver.state[id] = visitActive
	resolver.stackIndex[id] = len(resolver.stack)
	resolver.stack = append(resolver.stack, id)

	profile := resolver.profiles[id]
	base := rootCustomProfile()
	if profile.Extends != nil {
		if item, builtIn := resolver.builtins[*profile.Extends]; builtIn {
			base = profileFromPreset(item)
		} else {
			parent, err := resolver.resolveProfile(*profile.Extends)
			if err != nil {
				return resolvedProfile{}, err
			}
			base = parent
		}
	}

	resolved := mergeProfile(base, profile)
	resolver.stack = resolver.stack[:len(resolver.stack)-1]
	delete(resolver.stackIndex, id)
	resolver.state[id] = visitDone
	resolver.memo[id] = resolved.clone()

	return resolved.clone(), nil
}

func (resolver *graphResolver) resolveSelected(selected selectedBase) (Effective, error) {
	switch selected.kind {
	case KindNone:
		return Effective{Kind: KindNone}, nil
	case KindPreset:
		item, ok := resolver.builtins[selected.id]
		if !ok {
			return Effective{}, policyError(
				resolver.source,
				selected.field,
				fmt.Sprintf(
					"unknown built-in preset %q; available presets: %s",
					selected.id,
					strings.Join(resolver.builtinIDs, ", "),
				),
			)
		}
		resolved := profileFromPreset(item)
		return Effective{
			Kind:     KindPreset,
			ID:       selected.id,
			Metadata: resolved.metadata,
			Budgets:  resolved.budgets.Clone(),
		}, nil
	case KindProfile:
		resolved, ok := resolver.memo[selected.id]
		if !ok {
			available := "none"
			if len(resolver.profileIDs) > 0 {
				available = strings.Join(resolver.profileIDs, ", ")
			}
			return Effective{}, policyError(
				resolver.source,
				selected.field,
				fmt.Sprintf(
					"unknown custom profile %q; available profiles: %s",
					selected.id,
					available,
				),
			)
		}
		return Effective{
			Kind:     KindProfile,
			ID:       selected.id,
			Metadata: resolved.metadata,
			Budgets:  resolved.budgets.Clone(),
		}, nil
	default:
		return Effective{}, policyError(
			resolver.source,
			selected.field,
			fmt.Sprintf("unknown policy selector kind %q", selected.kind),
		)
	}
}

func selectBase(source string, configuration config.Config, cli Selector) (selectedBase, error) {
	if cli.Preset != "" && cli.Profile != "" {
		return selectedBase{}, policyError(
			source,
			"cli.preset/profile",
			"CLI preset and profile selectors are mutually exclusive",
		)
	}
	if cli.Preset != "" {
		return selectedBase{kind: KindPreset, id: cli.Preset, field: "cli.preset"}, nil
	}
	if cli.Profile != "" {
		return selectedBase{kind: KindProfile, id: cli.Profile, field: "cli.profile"}, nil
	}

	if configuration.Preset != nil && configuration.Profile != nil {
		return selectedBase{}, policyError(
			source,
			"preset/profile",
			"config preset and profile selectors are mutually exclusive",
		)
	}
	if configuration.Preset != nil {
		return selectedBase{kind: KindPreset, id: *configuration.Preset, field: "preset"}, nil
	}
	if configuration.Profile != nil {
		return selectedBase{kind: KindProfile, id: *configuration.Profile, field: "profile"}, nil
	}

	return selectedBase{kind: KindNone}, nil
}

func rootCustomProfile() resolvedProfile {
	return resolvedProfile{metadata: Metadata{
		Platform:  "custom",
		Renderer:  "unspecified",
		TargetFPS: 0,
		Quality:   "custom",
		Status:    "custom",
	}}
}

func profileFromPreset(item preset.Preset) resolvedProfile {
	return resolvedProfile{
		metadata: Metadata{
			Name:        item.Name,
			Description: item.Description,
			Platform:    item.Platform,
			Renderer:    item.Renderer,
			TargetFPS:   int64(item.TargetFPS),
			Quality:     item.Quality,
			Status:      item.Status,
			Stability:   item.Stability,
		},
		budgets: item.Budgets.Clone(),
	}
}

func mergeProfile(base resolvedProfile, profile config.Profile) resolvedProfile {
	merged := base.clone()
	if profile.Name != nil {
		merged.metadata.Name = *profile.Name
	}
	if profile.Description != nil {
		merged.metadata.Description = *profile.Description
	}
	if profile.Platform != nil {
		merged.metadata.Platform = *profile.Platform
	}
	if profile.Renderer != nil {
		merged.metadata.Renderer = *profile.Renderer
	}
	if profile.TargetFPS != nil {
		merged.metadata.TargetFPS = *profile.TargetFPS
	}
	if profile.Quality != nil {
		merged.metadata.Quality = *profile.Quality
	}
	merged.budgets = mergeLimits(merged.budgets, profile.Budgets)

	return merged
}
