package policy

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestListAndExplainProfilesReuseEffectiveResolutionWithProvenance(t *testing.T) {
	t.Parallel()

	configuration, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "fail_on_partial": false,
  "budgets": {"nodes": 90, "lights": 7},
  "profiles": {
    "shipping": {
      "name": "Shipping",
      "extends": "base",
      "renderer": "compatibility",
      "target_fps": 30,
      "quality": "low"
    },
    "ci": {"budgets": {"nodes": 1}},
    "base": {
      "extends": "steam-deck",
      "platform": "handheld",
      "budgets": {"nodes": 100}
    }
  }
}`), "profiles.json")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	listed, err := ListProfiles("profiles.json", configuration)
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if got := []string{listed[0].ID, listed[1].ID, listed[2].ID}; !reflect.DeepEqual(got, []string{"base", "ci", "shipping"}) {
		t.Fatalf("profile order = %v", got)
	}
	if listed[2].Name != "Shipping" || listed[2].Description == "" || listed[2].Extends != "base" {
		t.Fatalf("shipping summary = %#v", listed[2])
	}

	explanation, err := ExplainProfile("profiles.json", configuration, "shipping")
	if err != nil {
		t.Fatalf("ExplainProfile() error = %v", err)
	}
	effective, err := Resolve("profiles.json", configuration, Selector{Profile: "shipping"}, nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !reflect.DeepEqual(explanation.Effective, effective) {
		t.Fatalf("explained effective = %#v, ordinary = %#v", explanation.Effective, effective)
	}
	wantChain := []Layer{
		{Kind: LayerPreset, ID: "steam-deck"},
		{Kind: LayerProfile, ID: "base"},
		{Kind: LayerProfile, ID: "shipping"},
	}
	if !reflect.DeepEqual(explanation.Chain, wantChain) {
		t.Fatalf("chain = %#v, want %#v", explanation.Chain, wantChain)
	}
	assertLayer(t, explanation.MetadataSources.Name, LayerProfile, "shipping")
	assertLayer(t, explanation.MetadataSources.Description, LayerPreset, "steam-deck")
	assertLayer(t, explanation.MetadataSources.Platform, LayerProfile, "base")
	assertLayer(t, explanation.MetadataSources.Renderer, LayerProfile, "shipping")
	assertLayer(t, explanation.MetadataSources.TargetFPS, LayerProfile, "shipping")
	assertLayer(t, explanation.MetadataSources.Quality, LayerProfile, "shipping")
	assertLayer(t, explanation.MetadataSources.Status, LayerPreset, "steam-deck")
	assertLayer(t, explanation.MetadataSources.Stability, LayerPreset, "steam-deck")
	if explanation.FailOnPartial || explanation.FailOnPartialSource != (Layer{Kind: LayerProject}) {
		t.Fatalf("partial policy = %t / %#v", explanation.FailOnPartial, explanation.FailOnPartialSource)
	}

	for _, name := range metrics.OrderedNames() {
		_, present := explanation.Effective.Budgets.Get(name)
		source, sourced := explanation.BudgetSources.Get(name)
		if !present || !sourced {
			t.Fatalf("budget %s missing value/source", name)
		}
		if name == metrics.Nodes || name == metrics.Lights {
			assertLayer(t, source, LayerProject, "")
		} else {
			assertLayer(t, source, LayerPreset, "steam-deck")
		}
	}

	clone := explanation.Clone()
	*clone.Effective.Budgets.Nodes = 999
	clone.Chain[0].ID = "changed"
	clone.BudgetSources.Nodes.Kind = LayerDefault
	nodes, _ := explanation.Effective.Budgets.Get(metrics.Nodes)
	nodeSource, _ := explanation.BudgetSources.Get(metrics.Nodes)
	if nodes != 90 || explanation.Chain[0].ID != "steam-deck" || nodeSource.Kind != LayerProject {
		t.Fatalf("Clone() aliases explanation: %#v", explanation)
	}
}

func TestExplainRootDefaultsAndMissingBudgetSources(t *testing.T) {
	t.Parallel()

	configuration, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "profiles": {"minimal": {"name": "Minimal", "budgets": {"nodes": 0}}}
}`), "minimal.json")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	explanation, err := ExplainProfile("minimal.json", configuration, "minimal")
	if err != nil {
		t.Fatalf("ExplainProfile() error = %v", err)
	}
	assertLayer(t, explanation.MetadataSources.Name, LayerProfile, "minimal")
	for _, source := range []Layer{
		explanation.MetadataSources.Description,
		explanation.MetadataSources.Platform,
		explanation.MetadataSources.Renderer,
		explanation.MetadataSources.TargetFPS,
		explanation.MetadataSources.Quality,
		explanation.MetadataSources.Status,
		explanation.MetadataSources.Stability,
		explanation.FailOnPartialSource,
	} {
		assertLayer(t, source, LayerDefault, "")
	}
	if source, present := explanation.BudgetSources.Get(metrics.Lights); present || source != (Layer{}) {
		t.Fatalf("absent lights source = %#v / %t", source, present)
	}
}

func TestProfileInspectionRejectsWholeGraphAndUnknownSelection(t *testing.T) {
	t.Parallel()

	configuration, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "profiles": {
    "valid": {"budgets": {"nodes": 1}},
    "a": {"extends": "b"},
    "b": {"extends": "a"}
  }
}`), "cycle.json")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if _, err := ListProfiles("cycle.json", configuration); err == nil || !strings.Contains(err.Error(), "a -> b -> a") {
		t.Fatalf("ListProfiles() cycle error = %v", err)
	}
	if _, err := ExplainProfile("cycle.json", configuration, "valid"); err == nil || !strings.Contains(err.Error(), "a -> b -> a") {
		t.Fatalf("ExplainProfile() whole-graph error = %v", err)
	}

	valid, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "profiles": {"valid": {"budgets": {"nodes": 1}}}
}`), "valid.json")
	if err != nil {
		t.Fatalf("Decode(valid) error = %v", err)
	}
	if _, err := ExplainProfile("valid.json", valid, "mobile"); err == nil || !strings.Contains(err.Error(), "unknown custom profile") {
		t.Fatalf("ExplainProfile(unknown) error = %v", err)
	}
}

func assertLayer(t *testing.T, got Layer, kind LayerKind, id string) {
	t.Helper()
	want := Layer{Kind: kind, ID: id}
	if got != want {
		t.Fatalf("layer = %#v, want %#v", got, want)
	}
}
