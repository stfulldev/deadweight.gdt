package analysis

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func TestInvocationCacheMemoizesSuccessfulSceneEffects(t *testing.T) {
	path := testScenePath(t.TempDir(), "root.tscn")
	effects := &memorySceneEffects{sources: map[string]string{
		path.Canonical: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node3D\"]\n",
	}}
	cache := newInvocationCache()

	first, err := cache.loadDocument(path, effects.open, effects.parse)
	if err != nil {
		t.Fatalf("first loadDocument() error = %v", err)
	}
	second, err := cache.loadDocument(path, effects.open, effects.parse)
	if err != nil {
		t.Fatalf("second loadDocument() error = %v", err)
	}
	if first == nil || first != second {
		t.Fatalf("cached documents = %p/%p, want same non-nil document", first, second)
	}
	requireMemorySceneEffects(t, effects, path, 1)
	if parsed, countErr := cache.parsedSceneFiles(); countErr != nil || parsed != 1 {
		t.Fatalf("parsedSceneFiles() = %d, %v; want 1", parsed, countErr)
	}
}

func TestInvocationCacheMemoizesSceneFailures(t *testing.T) {
	openCause := errors.New("open failed")
	closeCause := errors.New("close failed")
	tests := []struct {
		name        string
		source      string
		openErr     error
		closeErr    error
		wantCode    diagnostic.Code
		wantOpen    int
		wantParse   int
		wantClose   int
		wantWrapped bool
	}{
		{
			name:        "open",
			openErr:     openCause,
			wantOpen:    1,
			wantWrapped: true,
		},
		{
			name:      "parse",
			source:    "not a tscn document",
			wantCode:  diagnostic.CodeInvalidTSCNRoot,
			wantOpen:  1,
			wantParse: 1,
			wantClose: 1,
		},
		{
			name:        "close",
			source:      "[gd_scene format=3]\n[node name=\"Root\" type=\"Node3D\"]\n",
			closeErr:    closeCause,
			wantOpen:    1,
			wantParse:   1,
			wantClose:   1,
			wantWrapped: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := testScenePath(t.TempDir(), "failure.tscn")
			effects := &memorySceneEffects{
				sources:     map[string]string{path.Canonical: test.source},
				errors:      map[string]error{path.Canonical: test.openErr},
				closeErrors: map[string]error{path.Canonical: test.closeErr},
			}
			cache := newInvocationCache()

			firstDocument, firstErr := cache.loadDocument(path, effects.open, effects.parse)
			secondDocument, secondErr := cache.loadDocument(path, effects.open, effects.parse)
			if firstErr == nil || firstErr != secondErr || firstDocument != nil || secondDocument != nil {
				t.Fatalf("cached failure = (%p, %v)/(%p, %v)", firstDocument, firstErr, secondDocument, secondErr)
			}
			if code, coded := diagnostic.CodeOf(firstErr); code != test.wantCode || coded != (test.wantCode != "") {
				t.Fatalf("diagnostic code = %q, %v; want %q", code, coded, test.wantCode)
			}
			var loadErr *sceneLoadError
			if errors.As(firstErr, &loadErr) != test.wantWrapped {
				t.Fatalf("sceneLoadError wrapping = %v, want %v", loadErr != nil, test.wantWrapped)
			}
			if test.openErr != nil && !errors.Is(firstErr, test.openErr) {
				t.Fatalf("open error chain = %v, want %v", firstErr, test.openErr)
			}
			if test.closeErr != nil && !errors.Is(firstErr, test.closeErr) {
				t.Fatalf("close error chain = %v, want %v", firstErr, test.closeErr)
			}
			if effects.calls[path.Canonical] != test.wantOpen ||
				effects.parses[path.Canonical] != test.wantParse ||
				effects.closes[path.Canonical] != test.wantClose {
				t.Fatalf(
					"effects = open %d, parse %d, close %d; want %d/%d/%d",
					effects.calls[path.Canonical],
					effects.parses[path.Canonical],
					effects.closes[path.Canonical],
					test.wantOpen,
					test.wantParse,
					test.wantClose,
				)
			}
			if parsed, countErr := cache.parsedSceneFiles(); countErr != nil || parsed != 0 {
				t.Fatalf("parsedSceneFiles() = %d, %v; want zero", parsed, countErr)
			}
		})
	}
}

func TestInvocationCacheOwnsStoredDomainValues(t *testing.T) {
	canonical := testScenePath(t.TempDir(), "cached.tscn").Canonical
	cache := newInvocationCache()
	local := LocalSummary{
		Mounts:            []InstanceMount{{NodeRecord: NodeRecord{Name: "Child"}, Candidate: &ExternalResource{ID: "1_child", Path: "child.tscn"}}},
		ExternalResources: []ExternalResource{{ID: "1_asset", Path: "asset.res"}},
	}
	expanded := ExpandedSummary{
		Dependencies:      []string{"child.tscn"},
		ExternalResources: []ResourceIdentity{{Resolved: true, Canonical: "asset.res"}},
		Unresolved:        []UnresolvedInstance{{RawTarget: "missing.tscn", Occurrences: 1}},
	}
	resources := map[string]resourceResolution{
		"1_child": {resource: ExternalResource{ID: "1_child", Path: "child.tscn"}},
	}
	cache.storeLocalSummary(canonical, local)
	cache.storeExpandedSummary(canonical, expanded)
	cache.storeResources(canonical, resources)

	local.Mounts[0].Name = "mutated input"
	expanded.Dependencies[0] = "mutated input"
	resources["1_child"] = resourceResolution{}

	firstLocal, _, _ := cache.localSummary(canonical)
	firstExpanded, _ := cache.expandedSummary(canonical)
	firstResources, _ := cache.resources(canonical)
	firstLocal.Mounts[0].Name = "mutated output"
	firstLocal.Mounts[0].Candidate.Path = "mutated output"
	firstExpanded.Dependencies[0] = "mutated output"
	firstExpanded.Unresolved[0].RawTarget = "mutated output"
	firstResources["1_child"] = resourceResolution{}

	secondLocal, _, _ := cache.localSummary(canonical)
	secondExpanded, _ := cache.expandedSummary(canonical)
	secondResources, _ := cache.resources(canonical)
	if !reflect.DeepEqual(secondLocal, LocalSummary{
		Mounts:            []InstanceMount{{NodeRecord: NodeRecord{Name: "Child"}, Candidate: &ExternalResource{ID: "1_child", Path: "child.tscn"}}},
		ExternalResources: []ExternalResource{{ID: "1_asset", Path: "asset.res"}},
	}) {
		t.Fatalf("cached local summary changed: %#v", secondLocal)
	}
	if !reflect.DeepEqual(secondExpanded, ExpandedSummary{
		Dependencies:      []string{"child.tscn"},
		ExternalResources: []ResourceIdentity{{Resolved: true, Canonical: "asset.res"}},
		Unresolved:        []UnresolvedInstance{{RawTarget: "missing.tscn", Occurrences: 1}},
	}) {
		t.Fatalf("cached expanded summary changed: %#v", secondExpanded)
	}
	if secondResources["1_child"].resource.ID != "1_child" {
		t.Fatalf("cached resource resolutions changed: %#v", secondResources)
	}
}

func TestInvocationCacheMemoizesLocalSummaryFailure(t *testing.T) {
	canonical := testScenePath(t.TempDir(), "local.tscn").Canonical
	want := errors.New("local summary failed")
	cache := newInvocationCache()
	cache.storeLocalSummaryError(canonical, want)

	first, firstErr, firstExists := cache.localSummary(canonical)
	second, secondErr, secondExists := cache.localSummary(canonical)
	if !reflect.DeepEqual(first, LocalSummary{}) || !reflect.DeepEqual(second, LocalSummary{}) ||
		!firstExists || !secondExists || firstErr != want || secondErr != want {
		t.Fatalf("cached local errors = %#v/%v/%v and %#v/%v/%v", first, firstErr, firstExists, second, secondErr, secondExists)
	}
}

var _ SceneParser = tscn.Parse
