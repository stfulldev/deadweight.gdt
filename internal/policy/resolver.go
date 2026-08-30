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
	metadata        Metadata
	budgets         budget.Limits
	metadataSources MetadataSources
	budgetSources   LimitSources
	chain           []Layer
}

func (resolved resolvedProfile) clone() resolvedProfile {
	resolved.budgets = resolved.budgets.Clone()
	resolved.budgetSources = resolved.budgetSources.Clone()
	resolved.chain = append([]Layer(nil), resolved.chain...)
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
	resolver, err := resolveGraph(source, configuration.Profiles)
	if err != nil {
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

// ListProfiles validates the complete graph and returns custom profiles in
// canonical ID order with effective display metadata.
func ListProfiles(source string, configuration config.Config) ([]ProfileSummary, error) {
	configuration = configuration.Clone()
	resolver, err := resolveGraph(source, configuration.Profiles)
	if err != nil {
		return nil, err
	}

	result := make([]ProfileSummary, 0, len(resolver.profileIDs))
	for _, id := range resolver.profileIDs {
		declaration := configuration.Profiles[id]
		resolved := resolver.memo[id]
		extends := ""
		if declaration.Extends != nil {
			extends = *declaration.Extends
		}
		result = append(result, ProfileSummary{
			ID:          id,
			Extends:     extends,
			Name:        resolved.metadata.Name,
			Description: resolved.metadata.Description,
		})
	}

	return result, nil
}

// ExplainProfile resolves one custom profile with the same effective merge
// semantics as check and returns field-level provenance.
func ExplainProfile(source string, configuration config.Config, id string) (Explanation, error) {
	configuration = configuration.Clone()
	resolver, err := resolveGraph(source, configuration.Profiles)
	if err != nil {
		return Explanation{}, err
	}

	resolved, ok := resolver.memo[id]
	if !ok {
		available := "none"
		if len(resolver.profileIDs) > 0 {
			available = strings.Join(resolver.profileIDs, ", ")
		}
		return Explanation{}, policyError(
			source,
			"profile",
			fmt.Sprintf("unknown custom profile %q; available profiles: %s", id, available),
		)
	}
	resolved = mergeResolvedLimits(resolved, configuration.Budgets, Layer{Kind: LayerProject})
	if resolved.budgets.Count() == 0 {
		return Explanation{}, policyError(
			source,
			"budgets",
			"effective policy has no budget; select a profile with budgets or provide a config budget",
		)
	}

	failSource := Layer{Kind: LayerDefault}
	if configuration.FailOnPartialDeclared() {
		failSource = Layer{Kind: LayerProject}
	}
	explanation := Explanation{
		Effective: Effective{
			Kind:     KindProfile,
			ID:       id,
			Metadata: resolved.metadata,
			Budgets:  resolved.budgets.Clone(),
		},
		FailOnPartial:       configuration.FailOnPartial,
		FailOnPartialSource: failSource,
		Chain:               append([]Layer(nil), resolved.chain...),
		MetadataSources:     resolved.metadataSources,
		BudgetSources:       resolved.budgetSources.Clone(),
	}

	return explanation.Clone(), nil
}

func resolveGraph(source string, profiles map[string]config.Profile) (*graphResolver, error) {
	catalog, err := preset.Builtins()
	if err != nil {
		return nil, fmt.Errorf("load built-in presets: %w", err)
	}
	resolver, err := newGraphResolver(source, profiles, catalog)
	if err != nil {
		return nil, err
	}
	if err := resolver.resolveAll(); err != nil {
		return nil, err
	}

	return resolver, nil
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

	resolved := mergeProfile(base, profile, id)
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
	defaultLayer := Layer{Kind: LayerDefault}
	return resolvedProfile{
		metadata: Metadata{
			Platform:  "custom",
			Renderer:  "unspecified",
			TargetFPS: 0,
			Quality:   "custom",
			Status:    "custom",
		},
		metadataSources: MetadataSources{
			Name:        defaultLayer,
			Description: defaultLayer,
			Platform:    defaultLayer,
			Renderer:    defaultLayer,
			TargetFPS:   defaultLayer,
			Quality:     defaultLayer,
			Status:      defaultLayer,
			Stability:   defaultLayer,
		},
	}
}

func profileFromPreset(item preset.Preset) resolvedProfile {
	layer := Layer{Kind: LayerPreset, ID: item.ID}
	resolved := resolvedProfile{
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
		metadataSources: MetadataSources{
			Name:        layer,
			Description: layer,
			Platform:    layer,
			Renderer:    layer,
			TargetFPS:   layer,
			Quality:     layer,
			Status:      layer,
			Stability:   layer,
		},
		chain: []Layer{layer},
	}
	return mergeResolvedLimits(resolved, item.Budgets, layer)
}

func mergeProfile(base resolvedProfile, profile config.Profile, id string) resolvedProfile {
	merged := base.clone()
	layer := Layer{Kind: LayerProfile, ID: id}
	if profile.Name != nil {
		merged.metadata.Name = *profile.Name
		merged.metadataSources.Name = layer
	}
	if profile.Description != nil {
		merged.metadata.Description = *profile.Description
		merged.metadataSources.Description = layer
	}
	if profile.Platform != nil {
		merged.metadata.Platform = *profile.Platform
		merged.metadataSources.Platform = layer
	}
	if profile.Renderer != nil {
		merged.metadata.Renderer = *profile.Renderer
		merged.metadataSources.Renderer = layer
	}
	if profile.TargetFPS != nil {
		merged.metadata.TargetFPS = *profile.TargetFPS
		merged.metadataSources.TargetFPS = layer
	}
	if profile.Quality != nil {
		merged.metadata.Quality = *profile.Quality
		merged.metadataSources.Quality = layer
	}
	merged = mergeResolvedLimits(merged, profile.Budgets, layer)
	merged.chain = append(merged.chain, layer)

	return merged
}
