package config

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
)

func TestDecodeAcceptsMinimalAndFullVersionOneDocuments(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		got, err := Decode(strings.NewReader(`{"version":1}`), "minimal.json")
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		want := Config{Version: CurrentVersion, Profiles: map[string]Profile{}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Decode() = %#v, want %#v", got, want)
		}
	})

	t.Run("full", func(t *testing.T) {
		got, err := Decode(strings.NewReader(`{
  "version": 1,
  "profile": "shipping",
  "fail_on_partial": true,
  "budgets": {
    "nodes": 0,
    "tree_depth": 0,
    "scene_instances": 0,
    "mesh_instances": 0,
    "lights": 0,
    "shadow_lights": 0,
    "external_resources": 0,
    "scene_dependencies": 9223372036854775807
  },
  "profiles": {
    "shipping": {
      "name": "Shipping",
      "description": "Project policy",
      "extends": "future-base",
      "platform": "office_laptop",
      "renderer": "forward_plus",
      "target_fps": 0,
      "quality": "balanced",
      "budgets": { "nodes": 0 }
    }
  }
}`), "full.json")
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if got.Version != CurrentVersion || got.Profile == nil || *got.Profile != "shipping" || got.Preset != nil || !got.FailOnPartial {
			t.Fatalf("config metadata = %#v", got)
		}
		if got.Budgets.Count() != 8 {
			t.Fatalf("budget count = %d, want 8", got.Budgets.Count())
		}
		for _, field := range []struct {
			name  string
			value *int64
		}{
			{name: "nodes", value: got.Budgets.Nodes},
			{name: "tree_depth", value: got.Budgets.TreeDepth},
			{name: "scene_instances", value: got.Budgets.SceneInstances},
			{name: "mesh_instances", value: got.Budgets.MeshInstances},
			{name: "lights", value: got.Budgets.Lights},
			{name: "shadow_lights", value: got.Budgets.ShadowLights},
			{name: "external_resources", value: got.Budgets.ExternalResources},
		} {
			if field.value == nil || *field.value != 0 {
				t.Errorf("budget %s = %v, want configured zero", field.name, field.value)
			}
		}
		if got.Budgets.SceneDependencies == nil || *got.Budgets.SceneDependencies != math.MaxInt64 {
			t.Fatalf("scene_dependencies = %v", got.Budgets.SceneDependencies)
		}
		profile := got.Profiles["shipping"]
		if profile.Name == nil || *profile.Name != "Shipping" ||
			profile.Description == nil || *profile.Description != "Project policy" ||
			profile.Extends == nil || *profile.Extends != "future-base" ||
			profile.Platform == nil || *profile.Platform != "office_laptop" ||
			profile.Renderer == nil || *profile.Renderer != "forward_plus" ||
			profile.TargetFPS == nil || *profile.TargetFPS != 0 ||
			profile.Quality == nil || *profile.Quality != "balanced" ||
			profile.Budgets.Nodes == nil || *profile.Budgets.Nodes != 0 {
			t.Fatalf("profile = %#v", profile)
		}
	})
}

func TestDecodeRetainsPatternValidUnknownDynamicReferences(t *testing.T) {
	got, err := Decode(strings.NewReader(`{
  "version": 1,
  "preset": "future-preset",
  "profiles": {"shipping":{"extends":"future-parent"}}
}`), "dynamic.json")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got.Preset == nil || *got.Preset != "future-preset" || got.Profiles["shipping"].Extends == nil || *got.Profiles["shipping"].Extends != "future-parent" {
		t.Fatalf("dynamic references = %#v", got)
	}
}

func TestDecodeRejectsInvalidJSONAndStaticDeclarations(t *testing.T) {
	longID := "a" + strings.Repeat("b", 64)
	tests := []struct {
		name   string
		input  string
		field  string
		reason ErrorReason
	}{
		{name: "empty", input: ``, reason: ReasonDecode},
		{name: "malformed", input: `{`, reason: ReasonDecode},
		{name: "root null", input: `null`, reason: ReasonDecode},
		{name: "root array", input: `[]`, reason: ReasonDecode},
		{name: "trailing object", input: `{"version":1}{}`, reason: ReasonDecode},
		{name: "trailing garbage", input: `{"version":1} nope`, reason: ReasonDecode},
		{name: "unknown top-level", input: `{"version":1,"unexpected":true}`, field: "unexpected", reason: ReasonDecode},
		{name: "missing version", input: `{}`, field: "version", reason: ReasonValidation},
		{name: "null version", input: `{"version":null}`, field: "version", reason: ReasonDecode},
		{name: "string version", input: `{"version":"1"}`, field: "version", reason: ReasonDecode},
		{name: "float version", input: `{"version":1.0}`, field: "version", reason: ReasonDecode},
		{name: "unsupported version", input: `{"version":2}`, field: "version", reason: ReasonValidation},
		{name: "null preset", input: `{"version":1,"preset":null}`, field: "preset", reason: ReasonDecode},
		{name: "null profile", input: `{"version":1,"profile":null}`, field: "profile", reason: ReasonDecode},
		{name: "null partial flag", input: `{"version":1,"fail_on_partial":null}`, field: "fail_on_partial", reason: ReasonDecode},
		{name: "string partial flag", input: `{"version":1,"fail_on_partial":"true"}`, field: "fail_on_partial", reason: ReasonDecode},
		{name: "null budgets", input: `{"version":1,"budgets":null}`, field: "budgets", reason: ReasonDecode},
		{name: "array budgets", input: `{"version":1,"budgets":[]}`, field: "budgets", reason: ReasonDecode},
		{name: "unknown metric", input: `{"version":1,"budgets":{"triangles":1}}`, field: "budgets.triangles", reason: ReasonDecode},
		{name: "negative budget", input: `{"version":1,"budgets":{"nodes":-1}}`, field: "budgets.nodes", reason: ReasonValidation},
		{name: "float budget", input: `{"version":1,"budgets":{"nodes":1.5}}`, field: "budgets.nodes", reason: ReasonDecode},
		{name: "string budget", input: `{"version":1,"budgets":{"nodes":"1"}}`, field: "budgets.nodes", reason: ReasonDecode},
		{name: "boolean budget", input: `{"version":1,"budgets":{"nodes":true}}`, field: "budgets.nodes", reason: ReasonDecode},
		{name: "null budget", input: `{"version":1,"budgets":{"nodes":null}}`, field: "budgets.nodes", reason: ReasonDecode},
		{name: "overflow budget", input: `{"version":1,"budgets":{"nodes":9223372036854775808}}`, field: "budgets.nodes", reason: ReasonDecode},
		{name: "selector conflict", input: `{"version":1,"preset":"mobile","profile":"shipping"}`, field: "preset/profile", reason: ReasonValidation},
		{name: "empty preset id", input: `{"version":1,"preset":""}`, field: "preset", reason: ReasonValidation},
		{name: "uppercase profile id", input: `{"version":1,"profile":"Shipping"}`, field: "profile", reason: ReasonValidation},
		{name: "long selector id", input: `{"version":1,"preset":"` + longID + `"}`, field: "preset", reason: ReasonValidation},
		{name: "null profiles", input: `{"version":1,"profiles":null}`, field: "profiles", reason: ReasonDecode},
		{name: "array profiles", input: `{"version":1,"profiles":[]}`, field: "profiles", reason: ReasonDecode},
		{name: "invalid profile key", input: `{"version":1,"profiles":{"Bad":{}}}`, field: "profiles.Bad", reason: ReasonValidation},
		{name: "null profile object", input: `{"version":1,"profiles":{"shipping":null}}`, field: "profiles.shipping", reason: ReasonDecode},
		{name: "unknown profile field", input: `{"version":1,"profiles":{"shipping":{"status":"stable"}}}`, field: "profiles.shipping.status", reason: ReasonDecode},
		{name: "null name", input: `{"version":1,"profiles":{"shipping":{"name":null}}}`, field: "profiles.shipping.name", reason: ReasonDecode},
		{name: "null description", input: `{"version":1,"profiles":{"shipping":{"description":null}}}`, field: "profiles.shipping.description", reason: ReasonDecode},
		{name: "invalid extends", input: `{"version":1,"profiles":{"shipping":{"extends":"Bad Parent"}}}`, field: "profiles.shipping.extends", reason: ReasonValidation},
		{name: "empty platform", input: `{"version":1,"profiles":{"shipping":{"platform":""}}}`, field: "profiles.shipping.platform", reason: ReasonValidation},
		{name: "null renderer", input: `{"version":1,"profiles":{"shipping":{"renderer":null}}}`, field: "profiles.shipping.renderer", reason: ReasonDecode},
		{name: "invalid renderer", input: `{"version":1,"profiles":{"shipping":{"renderer":"deferred"}}}`, field: "profiles.shipping.renderer", reason: ReasonValidation},
		{name: "negative fps", input: `{"version":1,"profiles":{"shipping":{"target_fps":-1}}}`, field: "profiles.shipping.target_fps", reason: ReasonValidation},
		{name: "float fps", input: `{"version":1,"profiles":{"shipping":{"target_fps":60.5}}}`, field: "profiles.shipping.target_fps", reason: ReasonDecode},
		{name: "null quality", input: `{"version":1,"profiles":{"shipping":{"quality":null}}}`, field: "profiles.shipping.quality", reason: ReasonDecode},
		{name: "invalid quality", input: `{"version":1,"profiles":{"shipping":{"quality":"ultra"}}}`, field: "profiles.shipping.quality", reason: ReasonValidation},
		{name: "null profile budgets", input: `{"version":1,"profiles":{"shipping":{"budgets":null}}}`, field: "profiles.shipping.budgets", reason: ReasonDecode},
		{name: "invalid profile budget", input: `{"version":1,"profiles":{"shipping":{"budgets":{"lights":-1}}}}`, field: "profiles.shipping.budgets.lights", reason: ReasonValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Decode(strings.NewReader(test.input), "invalid.json")
			if err == nil || !reflect.DeepEqual(got, Config{}) {
				t.Fatalf("Decode() = %#v, %v; want zero result and error", got, err)
			}
			configErr := requireConfigError(t, err, test.reason, "invalid.json", test.field)
			if configErr.Detail == "" {
				t.Fatalf("error has no actionable detail: %#v", configErr)
			}
		})
	}
}

func TestValidateRejectsEveryNegativeBudgetAndSortsProfileFailures(t *testing.T) {
	negative := int64(-1)
	limits := budget.Limits{
		Nodes:             &negative,
		TreeDepth:         &negative,
		SceneInstances:    &negative,
		MeshInstances:     &negative,
		Lights:            &negative,
		ShadowLights:      &negative,
		ExternalResources: &negative,
		SceneDependencies: &negative,
	}
	fields := []string{
		"nodes", "tree_depth", "scene_instances", "mesh_instances", "lights",
		"shadow_lights", "external_resources", "scene_dependencies",
	}
	for _, field := range fields {
		field := field
		t.Run(field, func(t *testing.T) {
			selected := budget.Limits{}
			setBudgetTestField(&selected, field, &negative)
			err := Validate(Config{Version: CurrentVersion, Budgets: selected}, "constructed")
			_ = requireConfigError(t, err, ReasonValidation, "constructed", "budgets."+field)
		})
	}

	invalid := "Bad"
	err := Validate(Config{
		Version: CurrentVersion,
		Profiles: map[string]Profile{
			"z": {Extends: &invalid},
			"a": {Extends: &invalid},
		},
		Budgets: limits.Clone(),
	}, "sorted")
	configErr := requireConfigError(t, err, ReasonValidation, "sorted", "budgets.nodes")
	if configErr.Detail == "" {
		t.Fatal("missing detail")
	}

	err = Validate(Config{
		Version: CurrentVersion,
		Profiles: map[string]Profile{
			"z": {Extends: &invalid},
			"a": {Extends: &invalid},
		},
	}, "sorted")
	_ = requireConfigError(t, err, ReasonValidation, "sorted", "profiles.a.extends")
}

func TestConfigCloneOwnsMapsAndOptionalValues(t *testing.T) {
	preset := "mobile"
	name := "Shipping"
	fps := int64(60)
	zero := int64(0)
	original := Config{
		Version: CurrentVersion,
		Preset:  &preset,
		Budgets: budget.Limits{Nodes: &zero},
		Profiles: map[string]Profile{
			"shipping": {Name: &name, TargetFPS: &fps, Budgets: budget.Limits{Lights: &zero}},
		},
	}
	cloned := original.Clone()
	*cloned.Preset = "desktop"
	*cloned.Budgets.Nodes = 5
	profile := cloned.Profiles["shipping"]
	*profile.Name = "Mutated"
	*profile.TargetFPS = 30
	*profile.Budgets.Lights = 4
	cloned.Profiles["shipping"] = profile
	delete(cloned.Profiles, "shipping")

	if *original.Preset != "mobile" || *original.Budgets.Nodes != 0 {
		t.Fatalf("top-level clone aliased original: %#v", original)
	}
	originalProfile := original.Profiles["shipping"]
	if *originalProfile.Name != "Shipping" || *originalProfile.TargetFPS != 60 || *originalProfile.Budgets.Lights != 0 {
		t.Fatalf("profile clone aliased original: %#v", originalProfile)
	}
}

func TestIdentifierAndEnumCatalogsAreFrozenAndOwned(t *testing.T) {
	validIDs := []string{"a", "a-b.c_d", "0", "a" + strings.Repeat("b", 63)}
	for _, id := range validIDs {
		if !ValidID(id) {
			t.Errorf("ValidID(%q) = false", id)
		}
	}
	for _, id := range []string{"", "A", "-a", "a/b", "a" + strings.Repeat("b", 64)} {
		if ValidID(id) {
			t.Errorf("ValidID(%q) = true", id)
		}
	}

	wantRenderers := []string{"forward_plus", "mobile", "compatibility", "unspecified"}
	wantQualities := []string{"low", "balanced", "high", "custom"}
	if !reflect.DeepEqual(RendererIDs(), wantRenderers) || !reflect.DeepEqual(QualityIDs(), wantQualities) {
		t.Fatalf("enum catalogs = %#v/%#v", RendererIDs(), QualityIDs())
	}
	for _, renderer := range wantRenderers {
		if !ValidRenderer(renderer) {
			t.Errorf("ValidRenderer(%q) = false", renderer)
		}
	}
	for _, quality := range wantQualities {
		if !ValidQuality(quality) {
			t.Errorf("ValidQuality(%q) = false", quality)
		}
	}
	renderers := RendererIDs()
	qualities := QualityIDs()
	renderers[0] = "mutated"
	qualities[0] = "mutated"
	if RendererIDs()[0] != "forward_plus" || QualityIDs()[0] != "low" {
		t.Fatal("caller mutation changed enum catalog")
	}
}

func TestErrorReasonsAndWrapping(t *testing.T) {
	for _, reason := range []ErrorReason{
		ReasonMissingExplicit, ReasonNotRegular, ReasonFilesystem, ReasonDecode, ReasonValidation,
	} {
		if !reason.Valid() {
			t.Errorf("%q.Valid() = false", reason)
		}
	}
	if ErrorReason("unknown").Valid() {
		t.Fatal("unknown reason is valid")
	}

	cause := errors.New("cause")
	err := configError(ReasonDecode, "config.json", "budgets.nodes", "invalid value", cause)
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "SB2003") || !strings.Contains(err.DiagnosticMessage(), "budgets.nodes") {
		t.Fatalf("error = %#v / %q", err, err.Error())
	}
}

func setBudgetTestField(limits *budget.Limits, field string, value *int64) {
	switch field {
	case "nodes":
		limits.Nodes = value
	case "tree_depth":
		limits.TreeDepth = value
	case "scene_instances":
		limits.SceneInstances = value
	case "mesh_instances":
		limits.MeshInstances = value
	case "lights":
		limits.Lights = value
	case "shadow_lights":
		limits.ShadowLights = value
	case "external_resources":
		limits.ExternalResources = value
	case "scene_dependencies":
		limits.SceneDependencies = value
	default:
		panic("unsupported budget field " + field)
	}
}
