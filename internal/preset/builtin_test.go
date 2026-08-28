package preset_test

import (
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/preset"
)

func TestBuiltinsAreFrozenAndOrdered(t *testing.T) {
	t.Parallel()

	catalog, err := preset.Builtins()
	if err != nil {
		t.Fatalf("Builtins() error = %v", err)
	}

	wantIDs := []string{"mobile", "steam-deck", "desktop"}
	gotIDs := make([]string, 0, len(catalog))
	for _, item := range catalog {
		gotIDs = append(gotIDs, item.ID)
		if item.Budgets.Count() != len(metrics.OrderedNames()) {
			t.Errorf("preset %q has %d budgets, want %d", item.ID, item.Budgets.Count(), len(metrics.OrderedNames()))
		}
	}

	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("preset IDs = %v, want %v", gotIDs, wantIDs)
	}

	steamDeck, ok := catalog.Find("steam-deck")
	if !ok {
		t.Fatal("steam-deck preset not found")
	}

	want := map[metrics.Name]int64{
		metrics.Nodes:             3000,
		metrics.TreeDepth:         20,
		metrics.SceneInstances:    250,
		metrics.MeshInstances:     1000,
		metrics.Lights:            32,
		metrics.ShadowLights:      8,
		metrics.ExternalResources: 300,
		metrics.SceneDependencies: 80,
	}
	for name, expected := range want {
		actual, configured := steamDeck.Budgets.Get(name)
		if !configured || actual != expected {
			t.Errorf("steam-deck %s = (%d, %t), want (%d, true)", name, actual, configured, expected)
		}
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
	nodes, ok := second[0].Budgets.Get(metrics.Nodes)
	if !ok || nodes != 1500 {
		t.Fatal("Builtins returned budget pointers that alias package state")
	}
}
