package preset

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestDecodePresetRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Preset)
		want   string
	}{
		{name: "missing id", mutate: func(item *Preset) { item.ID = "" }, want: `field "id"`},
		{name: "mismatched id", mutate: func(item *Preset) { item.ID = "desktop" }, want: `contains id "desktop"`},
		{name: "missing name", mutate: func(item *Preset) { item.Name = "" }, want: `field "name"`},
		{name: "missing description", mutate: func(item *Preset) { item.Description = "" }, want: `field "description"`},
		{name: "missing platform", mutate: func(item *Preset) { item.Platform = "" }, want: `field "platform"`},
		{name: "missing renderer", mutate: func(item *Preset) { item.Renderer = "" }, want: `field "renderer"`},
		{name: "missing quality", mutate: func(item *Preset) { item.Quality = "" }, want: `field "quality"`},
		{name: "missing status", mutate: func(item *Preset) { item.Status = "" }, want: `field "status"`},
		{name: "invalid status", mutate: func(item *Preset) { item.Status = "certified" }, want: `field "status"`},
		{name: "missing stability", mutate: func(item *Preset) { item.Stability = "" }, want: `field "stability"`},
		{name: "invalid stability", mutate: func(item *Preset) { item.Stability = "stable" }, want: `field "stability"`},
		{name: "zero target fps", mutate: func(item *Preset) { item.TargetFPS = 0 }, want: `field "target_fps"`},
		{name: "negative target fps", mutate: func(item *Preset) { item.TargetFPS = -1 }, want: `field "target_fps"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := validTestPreset()
			test.mutate(&item)

			_, err := decodePreset(marshalTestPreset(t, item), "mobile")
			requireErrorContains(t, err, `built-in preset "mobile"`, test.want)
		})
	}
}

func TestDecodePresetRejectsInvalidJSONShape(t *testing.T) {
	valid := string(marshalTestPreset(t, validTestPreset()))
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "unknown field",
			data: strings.TrimSuffix(valid, "}") + `,"unexpected":true}`,
			want: `unknown field "unexpected"`,
		},
		{
			name: "trailing object",
			data: valid + `{}`,
			want: "trailing JSON data",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodePreset([]byte(test.data), "mobile")
			requireErrorContains(t, err, `built-in preset "mobile"`, test.want)
		})
	}
}

func TestLoadCatalogRejectsDuplicateIDs(t *testing.T) {
	files := fstest.MapFS{
		"data/mobile.json": {Data: marshalTestPreset(t, validTestPreset())},
	}

	_, err := loadCatalog(files, []string{"mobile", "mobile"})
	requireErrorContains(t, err, `built-in preset "mobile"`, "duplicate id")
}

func TestValidatePresetAcceptsRendererAndQualityIDs(t *testing.T) {
	for _, renderer := range []string{"forward_plus", "mobile", "compatibility", "unspecified"} {
		t.Run("renderer_"+renderer, func(t *testing.T) {
			item := validTestPreset()
			item.Renderer = renderer
			if err := validatePreset(item, item.ID); err != nil {
				t.Fatalf("validatePreset(renderer=%q): %v", renderer, err)
			}
		})
	}

	for _, quality := range []string{"low", "balanced", "high", "custom"} {
		t.Run("quality_"+quality, func(t *testing.T) {
			item := validTestPreset()
			item.Quality = quality
			if err := validatePreset(item, item.ID); err != nil {
				t.Fatalf("validatePreset(quality=%q): %v", quality, err)
			}
		})
	}
}

func TestValidatePresetRejectsUnsupportedRendererAndQuality(t *testing.T) {
	t.Run("renderer", func(t *testing.T) {
		item := validTestPreset()
		item.Renderer = "deferred"
		err := validatePreset(item, item.ID)
		requireErrorContains(t, err, `built-in preset "mobile"`, "renderer", "deferred")
	})

	t.Run("quality", func(t *testing.T) {
		item := validTestPreset()
		item.Quality = "ultra"
		err := validatePreset(item, item.ID)
		requireErrorContains(t, err, `built-in preset "mobile"`, "quality", "ultra")
	})
}

func TestValidatePresetRejectsInvalidBudgets(t *testing.T) {
	for _, name := range metrics.OrderedNames() {
		name := name

		t.Run("missing_"+string(name), func(t *testing.T) {
			item := validTestPreset()
			setTestBudget(&item.Budgets, name, nil)

			err := validatePreset(item, item.ID)
			requireErrorContains(t, err, `built-in preset "mobile"`, "missing budget", string(name))
		})

		t.Run("negative_"+string(name), func(t *testing.T) {
			item := validTestPreset()
			setTestBudget(&item.Budgets, name, testInt64Pointer(-1))

			err := validatePreset(item, item.ID)
			requireErrorContains(t, err, `built-in preset "mobile"`, string(name), "non-negative")
		})

		t.Run("zero_"+string(name), func(t *testing.T) {
			item := validTestPreset()
			setTestBudget(&item.Budgets, name, testInt64Pointer(0))

			if err := validatePreset(item, item.ID); err != nil {
				t.Fatalf("validatePreset(%s=0): %v", name, err)
			}
		})
	}
}

func TestDecodePresetRejectsNonIntegerBudgets(t *testing.T) {
	valid := string(marshalTestPreset(t, validTestPreset()))
	for _, replacement := range []string{`"nodes":1.5`, `"nodes":"10"`} {
		data := strings.Replace(valid, `"nodes":10`, replacement, 1)
		if data == valid {
			t.Fatal("test fixture did not contain nodes budget")
		}

		_, err := decodePreset([]byte(data), "mobile")
		requireErrorContains(t, err, `built-in preset "mobile"`, "nodes")
	}
}

func validTestPreset() Preset {
	return Preset{
		ID:          "mobile",
		Name:        "Mobile",
		Description: "Mobile-class 3D hardware",
		Platform:    "mobile",
		Renderer:    "mobile",
		TargetFPS:   30,
		Quality:     "low",
		Status:      "heuristic",
		Stability:   "experimental",
		Budgets: budget.Limits{
			Nodes:             testInt64Pointer(10),
			TreeDepth:         testInt64Pointer(10),
			SceneInstances:    testInt64Pointer(10),
			MeshInstances:     testInt64Pointer(10),
			Lights:            testInt64Pointer(10),
			ShadowLights:      testInt64Pointer(10),
			ExternalResources: testInt64Pointer(10),
			SceneDependencies: testInt64Pointer(10),
		},
	}
}

func marshalTestPreset(t *testing.T, item Preset) []byte {
	t.Helper()

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(Preset): %v", err)
	}
	return data
}

func setTestBudget(limits *budget.Limits, name metrics.Name, value *int64) {
	switch name {
	case metrics.Nodes:
		limits.Nodes = value
	case metrics.TreeDepth:
		limits.TreeDepth = value
	case metrics.SceneInstances:
		limits.SceneInstances = value
	case metrics.MeshInstances:
		limits.MeshInstances = value
	case metrics.Lights:
		limits.Lights = value
	case metrics.ShadowLights:
		limits.ShadowLights = value
	case metrics.ExternalResources:
		limits.ExternalResources = value
	case metrics.SceneDependencies:
		limits.SceneDependencies = value
	default:
		panic("unsupported test metric: " + name)
	}
}

func testInt64Pointer(value int64) *int64 {
	return &value
}

func requireErrorContains(t *testing.T, err error, fragments ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q does not contain %q", err, fragment)
		}
	}
}
