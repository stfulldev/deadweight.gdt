package preset_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
)

func TestBuiltinsAreFrozenAndOrdered(t *testing.T) {
	t.Parallel()

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatalf("Builtins() error = %v", err)
	}

	want := preset.Catalog{
		{
			ID:          "mobile",
			Name:        "Mobile",
			Description: "Mobile-class 3D hardware",
			Platform:    "mobile",
			Renderer:    "mobile",
			TargetFPS:   30,
			Quality:     "low",
			Status:      "heuristic",
			Stability:   "experimental",
			Budgets:     frozenLimits(1500, 15, 100, 500, 16, 4, 150, 40),
		},
		{
			ID:          "steam-deck",
			Name:        "Steam Deck",
			Description: "Steam Deck-class hardware",
			Platform:    "steam_deck",
			Renderer:    "forward_plus",
			TargetFPS:   60,
			Quality:     "balanced",
			Status:      "heuristic",
			Stability:   "experimental",
			Budgets:     frozenLimits(3000, 20, 250, 1000, 32, 8, 300, 80),
		},
		{
			ID:          "desktop",
			Name:        "Desktop",
			Description: "Mid-range desktop hardware",
			Platform:    "desktop",
			Renderer:    "forward_plus",
			TargetFPS:   60,
			Quality:     "high",
			Status:      "heuristic",
			Stability:   "experimental",
			Budgets:     frozenLimits(6000, 30, 500, 2500, 64, 16, 600, 160),
		},
	}

	if !reflect.DeepEqual(catalog, want) {
		t.Fatalf("frozen catalog mismatch\ngot:\n%s\nwant:\n%s", formatCatalog(t, catalog), formatCatalog(t, want))
	}
}

func TestBuiltinsLoadFromEmbeddedData(t *testing.T) {
	t.Chdir(t.TempDir())

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatalf("Builtins() outside repository working directory: %v", err)
	}
	if len(catalog) != 3 {
		t.Fatalf("len(Builtins()) = %d, want 3", len(catalog))
	}
}

func TestBuiltinsReturnsDefensiveCatalogCopy(t *testing.T) {
	t.Parallel()

	first, err := preset.Builtins()
	if err != nil {
		t.Fatal(err)
	}
	first[0].ID = "mutated"
	*first[0].Budgets.Nodes = 999999

	second, err := preset.Builtins()
	if err != nil {
		t.Fatal(err)
	}
	if second[0].ID != "mobile" {
		t.Fatal("Builtins returned mutable package catalog state")
	}
	if nodes := *second[0].Budgets.Nodes; nodes != 1500 {
		t.Fatal("Builtins returned budget pointers that alias package state")
	}
}

func TestCatalogFindReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatal(err)
	}

	first, err := catalog.Find("mobile")
	if err != nil {
		t.Fatalf("Find(mobile): %v", err)
	}
	first.ID = "mutated"
	*first.Budgets.Nodes = 999999

	second, err := catalog.Find("mobile")
	if err != nil {
		t.Fatalf("Find(mobile) after mutation: %v", err)
	}
	if second.ID != "mobile" || *second.Budgets.Nodes != 1500 {
		t.Fatalf("second lookup = %#v, want unmodified mobile preset", second)
	}
	if catalog[0].ID != "mobile" || *catalog[0].Budgets.Nodes != 1500 {
		t.Fatal("Find returned values that alias the source catalog")
	}
}

func TestCatalogFindReturnsActionableError(t *testing.T) {
	t.Parallel()

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatal(err)
	}

	_, err = catalog.Find("unknown")
	if err == nil {
		t.Fatal("Find(unknown) error = nil")
	}
	want := `unknown preset "unknown"; available presets: mobile, steam-deck, desktop`
	if err.Error() != want {
		t.Fatalf("Find(unknown) error = %q, want %q", err, want)
	}
}

func frozenLimits(nodes, treeDepth, sceneInstances, meshInstances, lights, shadowLights, externalResources, sceneDependencies int64) budget.Limits {
	return budget.Limits{
		Nodes:             int64Pointer(nodes),
		TreeDepth:         int64Pointer(treeDepth),
		SceneInstances:    int64Pointer(sceneInstances),
		MeshInstances:     int64Pointer(meshInstances),
		Lights:            int64Pointer(lights),
		ShadowLights:      int64Pointer(shadowLights),
		ExternalResources: int64Pointer(externalResources),
		SceneDependencies: int64Pointer(sceneDependencies),
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}

func formatCatalog(t *testing.T, catalog preset.Catalog) string {
	t.Helper()

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		t.Fatalf("marshal catalog: %v", err)
	}
	return string(data)
}
