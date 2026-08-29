package policy

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestKindAndEffectiveClone(t *testing.T) {
	t.Parallel()

	for _, kind := range []Kind{KindNone, KindPreset, KindProfile} {
		if !kind.Valid() {
			t.Errorf("%q.Valid() = false", kind)
		}
	}
	if Kind("unknown").Valid() {
		t.Fatal("unknown kind is valid")
	}

	original := Effective{Budgets: limitsOf(map[metrics.Name]int64{metrics.Nodes: 7})}
	cloned := original.Clone()
	*cloned.Budgets.Nodes = 99
	if got, _ := original.Budgets.Get(metrics.Nodes); got != 7 {
		t.Fatalf("original nodes = %d, want 7", got)
	}
}

func TestMergeLimitsCoversEveryMetricAndZero(t *testing.T) {
	t.Parallel()

	for index, name := range metrics.OrderedNames() {
		lower := limitsOf(map[metrics.Name]int64{name: int64(index + 1)})
		higher := limitsOf(map[metrics.Name]int64{name: 0})
		merged := mergeLimits(lower, higher)
		if got, configured := merged.Get(name); !configured || got != 0 {
			t.Errorf("merge %q = %d, %v; want 0, true", name, got, configured)
		}
		*limitPointer(t, merged, name) = 100
		if got, _ := lower.Get(name); got != int64(index+1) {
			t.Errorf("lower %q mutated to %d", name, got)
		}
	}

	lower := limitsOf(map[metrics.Name]int64{metrics.Nodes: 3})
	if got := mergeLimits(lower, budget.Limits{}); !reflect.DeepEqual(got, lower) {
		t.Fatalf("absent override = %#v, want %#v", got, lower)
	}
}

func TestResolveSelectorPrecedenceAndDomains(t *testing.T) {
	t.Parallel()

	shipping := "shipping"
	mobile := "mobile"
	configuration := config.Config{
		Version: config.CurrentVersion,
		Profile: &shipping,
		Profiles: map[string]config.Profile{
			shipping: {Extends: &mobile},
		},
	}

	effective, err := Resolve("config.json", configuration, Selector{Preset: "desktop"}, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if effective.Kind != KindPreset || effective.ID != "desktop" || effective.Metadata.Platform != "desktop" {
		t.Fatalf("CLI selected policy = %#v", effective)
	}

	effective, err = Resolve("config.json", configuration, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(config profile) error = %v", err)
	}
	if effective.Kind != KindProfile || effective.ID != shipping || effective.Metadata.Platform != "mobile" {
		t.Fatalf("config selected policy = %#v", effective)
	}

	noBase := config.Config{
		Version: config.CurrentVersion,
		Budgets: limitsOf(map[metrics.Name]int64{metrics.Nodes: 1}),
	}
	effective, err = Resolve("config.json", noBase, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(no base) error = %v", err)
	}
	if effective.Kind != KindNone || effective.ID != "" || effective.Metadata != (Metadata{}) {
		t.Fatalf("no-base policy = %#v", effective)
	}

	tests := []struct {
		name          string
		configuration config.Config
		selector      Selector
		field         string
		detail        string
	}{
		{
			name: "CLI conflict", configuration: noBase,
			selector: Selector{Preset: "mobile", Profile: "shipping"},
			field:    "cli.preset/profile", detail: "mutually exclusive",
		},
		{
			name: "config conflict",
			configuration: config.Config{
				Version: config.CurrentVersion,
				Preset:  stringPointer("mobile"), Profile: stringPointer("shipping"),
				Budgets: noBase.Budgets,
			},
			field: "preset/profile", detail: "mutually exclusive",
		},
		{
			name: "unknown built-in", configuration: noBase,
			selector: Selector{Preset: "unknown"},
			field:    "cli.preset", detail: "available presets: mobile, steam-deck, desktop",
		},
		{
			name: "preset does not fall through to profile",
			configuration: config.Config{
				Version:  config.CurrentVersion,
				Profiles: map[string]config.Profile{"shipping": {}},
				Budgets:  noBase.Budgets,
			},
			selector: Selector{Preset: "shipping"},
			field:    "cli.preset", detail: "unknown built-in preset",
		},
		{
			name: "profile does not fall through to preset", configuration: noBase,
			selector: Selector{Profile: "mobile"},
			field:    "cli.profile", detail: "unknown custom profile",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, resolveErr := Resolve("config.json", test.configuration, test.selector, nil)
			if got != (Effective{}) {
				t.Fatalf("result = %#v, want zero", got)
			}
			_ = requirePolicyError(t, resolveErr, "config.json", test.field, test.detail)
		})
	}
}

func TestResolveValidatesCompleteProfileGraphDeterministically(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		profiles map[string]config.Profile
		field    string
		detail   string
	}{
		{
			name:     "built-in collision",
			profiles: map[string]config.Profile{"steam-deck": {}},
			field:    "profiles.steam-deck",
			detail:   "collides with a built-in preset ID",
		},
		{
			name: "unselected missing parent",
			profiles: map[string]config.Profile{
				"valid":  {},
				"unused": {Extends: stringPointer("missing")},
			},
			field:  "profiles.unused.extends",
			detail: `extends unknown parent "missing"`,
		},
		{
			name: "sorted missing parent",
			profiles: map[string]config.Profile{
				"z": {Extends: stringPointer("missing-z")},
				"a": {Extends: stringPointer("missing-a")},
			},
			field:  "profiles.a.extends",
			detail: `extends unknown parent "missing-a"`,
		},
		{
			name: "full cycle",
			profiles: map[string]config.Profile{
				"a": {Extends: stringPointer("b")},
				"b": {Extends: stringPointer("c")},
				"c": {Extends: stringPointer("a")},
			},
			field:  "profiles.c.extends",
			detail: "profile inheritance cycle: a -> b -> c -> a",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			configuration := config.Config{
				Version:  config.CurrentVersion,
				Profiles: test.profiles,
				Budgets:  limitsOf(map[metrics.Name]int64{metrics.Nodes: 1}),
			}
			got, err := Resolve("graph.json", configuration, Selector{}, nil)
			if got != (Effective{}) {
				t.Fatalf("result = %#v, want zero", got)
			}
			_ = requirePolicyError(t, err, "graph.json", test.field, test.detail)
		})
	}
}

func TestResolveProfileDepthBoundary(t *testing.T) {
	t.Parallel()

	profiles32 := profileChain(32, "mobile")
	effective, err := Resolve("depth.json", config.Config{
		Version:  config.CurrentVersion,
		Profile:  stringPointer("p01"),
		Profiles: profiles32,
	}, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(32) error = %v", err)
	}
	if effective.Kind != KindProfile || effective.ID != "p01" || effective.Metadata.Platform != "mobile" {
		t.Fatalf("Resolve(32) = %#v", effective)
	}

	profiles33 := profileChain(33, "mobile")
	got, err := Resolve("depth.json", config.Config{
		Version:  config.CurrentVersion,
		Profile:  stringPointer("p01"),
		Profiles: profiles33,
	}, Selector{}, nil)
	if got != (Effective{}) {
		t.Fatalf("Resolve(33) = %#v, want zero", got)
	}
	configErr := requirePolicyError(t, err, "depth.json", "profiles.p32.extends", "depth exceeds 32")
	if !strings.Contains(configErr.Detail, "p01 -> p02") || !strings.Contains(configErr.Detail, "p32 -> p33") {
		t.Fatalf("depth detail = %q, want full chain", configErr.Detail)
	}
}

func TestResolveProfileMetadata(t *testing.T) {
	t.Parallel()

	effective, err := Resolve("metadata.json", config.Config{
		Version: config.CurrentVersion,
		Profile: stringPointer("shipping"),
		Profiles: map[string]config.Profile{
			"shipping": {
				Extends: stringPointer("steam-deck"),
				Name:    stringPointer("Shipping"),
				Quality: stringPointer("custom"),
			},
		},
	}, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(built-in parent) error = %v", err)
	}
	wantInherited := Metadata{
		Name:        "Shipping",
		Description: "Steam Deck-class hardware",
		Platform:    "steam_deck",
		Renderer:    "forward_plus",
		TargetFPS:   60,
		Quality:     "custom",
		Status:      "heuristic",
		Stability:   "experimental",
	}
	if effective.Metadata != wantInherited {
		t.Fatalf("metadata = %#v, want %#v", effective.Metadata, wantInherited)
	}

	root := config.Profile{
		Name:        stringPointer("Root"),
		Description: stringPointer("Root description"),
		Platform:    stringPointer("root-platform"),
		Renderer:    stringPointer("compatibility"),
		TargetFPS:   int64Pointer(15),
		Quality:     stringPointer("low"),
		Budgets:     limitsOf(map[metrics.Name]int64{metrics.Nodes: 1}),
	}
	effective, err = Resolve("metadata.json", config.Config{
		Version: config.CurrentVersion,
		Profile: stringPointer("child"),
		Profiles: map[string]config.Profile{
			"root": root,
			"child": {
				Extends:     stringPointer("root"),
				Description: stringPointer("Child description"),
				Platform:    stringPointer("child-platform"),
				Renderer:    stringPointer("mobile"),
				TargetFPS:   int64Pointer(0),
				Quality:     stringPointer("high"),
			},
		},
	}, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(custom parent) error = %v", err)
	}
	wantCustom := Metadata{
		Name:        "Root",
		Description: "Child description",
		Platform:    "child-platform",
		Renderer:    "mobile",
		TargetFPS:   0,
		Quality:     "high",
		Status:      "custom",
	}
	if effective.Metadata != wantCustom {
		t.Fatalf("custom metadata = %#v, want %#v", effective.Metadata, wantCustom)
	}

	effective, err = Resolve("metadata.json", config.Config{
		Version: config.CurrentVersion,
		Profile: stringPointer("root"),
		Profiles: map[string]config.Profile{
			"root": {Budgets: limitsOf(map[metrics.Name]int64{metrics.Nodes: 1})},
		},
	}, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(root defaults) error = %v", err)
	}
	wantDefaults := Metadata{
		Platform: "custom", Renderer: "unspecified", TargetFPS: 0,
		Quality: "custom", Status: "custom",
	}
	if effective.Metadata != wantDefaults {
		t.Fatalf("root defaults = %#v, want %#v", effective.Metadata, wantDefaults)
	}
}

func TestResolveFourLayerBudgetsAndOwnership(t *testing.T) {
	t.Parallel()

	configuration := config.Config{
		Version: config.CurrentVersion,
		Profile: stringPointer("shipping"),
		Budgets: limitsOf(map[metrics.Name]int64{metrics.ShadowLights: 6}),
		Profiles: map[string]config.Profile{
			"shipping": {
				Extends: stringPointer("steam-deck"),
				Budgets: limitsOf(map[metrics.Name]int64{metrics.MeshInstances: 1600}),
			},
		},
	}

	effective, err := Resolve("four-layer.json", configuration, Selector{}, []string{"nodes=4000"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := map[metrics.Name]int64{
		metrics.Nodes:             4000,
		metrics.TreeDepth:         20,
		metrics.SceneInstances:    250,
		metrics.MeshInstances:     1600,
		metrics.Lights:            32,
		metrics.ShadowLights:      6,
		metrics.ExternalResources: 300,
		metrics.SceneDependencies: 80,
	}
	for _, name := range metrics.OrderedNames() {
		got, configured := effective.Budgets.Get(name)
		if !configured || got != want[name] {
			t.Errorf("budget %q = %d, %v; want %d, true", name, got, configured, want[name])
		}
	}

	*effective.Budgets.Nodes = 999
	again, err := Resolve("four-layer.json", configuration, Selector{}, []string{"nodes=4000"})
	if err != nil {
		t.Fatalf("second Resolve() error = %v", err)
	}
	if got, _ := again.Budgets.Get(metrics.Nodes); got != 4000 {
		t.Fatalf("second nodes = %d, want 4000", got)
	}
	if got, _ := configuration.Profiles["shipping"].Budgets.Get(metrics.MeshInstances); got != 1600 {
		t.Fatalf("input profile mutated to %d", got)
	}
}

func TestResolveNoBaseBudgetSourcesAndEmptyPolicy(t *testing.T) {
	t.Parallel()

	projectOnly, err := Resolve("policy.json", config.Config{
		Version: config.CurrentVersion,
		Budgets: limitsOf(map[metrics.Name]int64{metrics.Nodes: 0}),
	}, Selector{}, nil)
	if err != nil {
		t.Fatalf("Resolve(project-only) error = %v", err)
	}
	if got, configured := projectOnly.Budgets.Get(metrics.Nodes); !configured || got != 0 {
		t.Fatalf("project-only nodes = %d, %v; want 0, true", got, configured)
	}

	cliOnly, err := Resolve("policy.json", config.Config{Version: config.CurrentVersion}, Selector{}, []string{
		"nodes=1", "lights=8", "nodes=2",
	})
	if err != nil {
		t.Fatalf("Resolve(CLI-only) error = %v", err)
	}
	if got, _ := cliOnly.Budgets.Get(metrics.Nodes); got != 2 {
		t.Fatalf("CLI-only nodes = %d, want 2", got)
	}
	if got, _ := cliOnly.Budgets.Get(metrics.Lights); got != 8 {
		t.Fatalf("CLI-only lights = %d, want 8", got)
	}

	got, err := Resolve("policy.json", config.Config{Version: config.CurrentVersion}, Selector{}, nil)
	if got != (Effective{}) {
		t.Fatalf("empty policy result = %#v, want zero", got)
	}
	_ = requirePolicyError(t, err, "policy.json", "budgets", "select a preset/profile or provide a config/CLI budget")
}

func profileChain(count int, terminal string) map[string]config.Profile {
	profiles := make(map[string]config.Profile, count)
	for index := 1; index <= count; index++ {
		id := fmt.Sprintf("p%02d", index)
		parent := terminal
		if index < count {
			parent = fmt.Sprintf("p%02d", index+1)
		}
		profiles[id] = config.Profile{Extends: stringPointer(parent)}
	}
	return profiles
}

func limitsOf(values map[metrics.Name]int64) budget.Limits {
	var limits budget.Limits
	for name, value := range values {
		setLimit(&limits, name, value)
	}
	return limits
}

func limitPointer(t *testing.T, limits budget.Limits, name metrics.Name) *int64 {
	t.Helper()
	switch name {
	case metrics.Nodes:
		return limits.Nodes
	case metrics.TreeDepth:
		return limits.TreeDepth
	case metrics.SceneInstances:
		return limits.SceneInstances
	case metrics.MeshInstances:
		return limits.MeshInstances
	case metrics.Lights:
		return limits.Lights
	case metrics.ShadowLights:
		return limits.ShadowLights
	case metrics.ExternalResources:
		return limits.ExternalResources
	case metrics.SceneDependencies:
		return limits.SceneDependencies
	default:
		t.Fatalf("unknown metric %q", name)
		return nil
	}
}

func stringPointer(value string) *string {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}

func requirePolicyError(t *testing.T, err error, source, field, detail string) *config.Error {
	t.Helper()

	var configErr *config.Error
	if !errors.As(err, &configErr) {
		t.Fatalf("error = %T %v, want *config.Error", err, err)
	}
	if configErr.Reason != config.ReasonValidation || configErr.Source != source || configErr.Field != field {
		t.Fatalf("config error = %#v", configErr)
	}
	if !strings.Contains(configErr.Detail, detail) {
		t.Fatalf("detail = %q, want substring %q", configErr.Detail, detail)
	}
	if configErr.DiagnosticCode() != diagnostic.CodeInvalidConfiguration {
		t.Fatalf("diagnostic code = %q", configErr.DiagnosticCode())
	}
	if !strings.Contains(configErr.Error(), "SB2003") || diagnostic.MessageOf(err) != configErr.DiagnosticMessage() {
		t.Fatalf("diagnostic contract mismatch: %q / %q", configErr.Error(), diagnostic.MessageOf(err))
	}
	return configErr
}
