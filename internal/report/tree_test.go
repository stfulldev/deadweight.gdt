package report

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestTreeReportGoldens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result application.TreeResult
	}{
		{name: "tree_complete", result: completeTreeFixture("<PROJECT>")},
		{name: "tree_repeated_diamond", result: repeatedDiamondTreeFixture("<PROJECT>")},
		{name: "tree_partial", result: partialTreeFixture("<PROJECT>")},
		{name: "tree_inherited", result: inheritedTreeFixture("<PROJECT>")},
		{name: "tree_root_only", result: rootOnlyTreeFixture("<PROJECT>")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := Tree(test.result, Options{Version: "0.2.0-test"})
			if err != nil {
				t.Fatalf("Tree() error = %v", err)
			}
			assertTextFraming(t, got)
			goldenPath := filepath.Join("testdata", "golden", test.name+".golden")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
					t.Fatalf("update golden: %v", err)
				}
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
			}
		})
	}
}

func TestTreeJSONGoldensValidateAgainstSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result application.TreeResult
	}{
		{name: "tree_complete", result: completeTreeFixture("<PROJECT>")},
		{name: "tree_repeated_diamond", result: repeatedDiamondTreeFixture("<PROJECT>")},
		{name: "tree_partial", result: partialTreeFixture("<PROJECT>")},
		{name: "tree_inherited", result: inheritedTreeFixture("<PROJECT>")},
		{name: "tree_root_only", result: rootOnlyTreeFixture("<PROJECT>")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := TreeJSON(test.result, Options{Version: "0.2.0-test", Color: true})
			if err != nil {
				t.Fatalf("TreeJSON() error = %v", err)
			}
			assertJSONFraming(t, got)
			validateReportDocument(t, []byte(got))
			goldenPath := filepath.Join("testdata", "golden", "json", test.name+".golden")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(goldenPath, []byte(got), 0o600); err != nil {
					t.Fatalf("update golden: %v", err)
				}
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if got != string(want) {
				t.Fatalf("golden mismatch\n--- want ---\n%s--- got ---\n%s", want, got)
			}
		})
	}
}

func TestDependencyTreeProjectionIsBoundedAndUsesBackReferences(t *testing.T) {
	t.Parallel()

	result := repeatedDiamondTreeFixture("<PROJECT>")
	tree, err := projectDependencyTree(result)
	if err != nil {
		t.Fatalf("projectDependencyTree() error = %v", err)
	}
	if len(tree.Entries) != len(result.Inspect.Analysis.Graph.Edges) {
		t.Fatalf("entries/edges = %d / %d", len(tree.Entries), len(result.Inspect.Analysis.Graph.Edges))
	}
	wantTargets := []string{
		"res://scenes/branch-b.tscn",
		"res://scenes/shared.tscn",
		"res://scenes/leaf.tscn",
		"res://scenes/branch-c.tscn",
		"res://scenes/shared.tscn",
	}
	wantDepths := []int64{1, 2, 3, 1, 2}
	gotTargets := make([]string, 0, len(tree.Entries))
	gotDepths := make([]int64, 0, len(tree.Entries))
	backReferences := 0
	for _, entry := range tree.Entries {
		gotTargets = append(gotTargets, entry.Target)
		gotDepths = append(gotDepths, entry.Depth)
		if entry.BackReference {
			backReferences++
		}
	}
	if !reflect.DeepEqual(gotTargets, wantTargets) || !reflect.DeepEqual(gotDepths, wantDepths) {
		t.Fatalf("targets/depths = %v / %v", gotTargets, gotDepths)
	}
	if tree.Entries[0].Occurrences != 100 || backReferences != 1 || !tree.Entries[4].BackReference {
		t.Fatalf("repeated/back-reference projection = %#v", tree.Entries)
	}
}

func TestDependencyTreeReliabilityRetainsUnresolvedAndInheritedEvidence(t *testing.T) {
	t.Parallel()

	partial, err := projectDependencyTree(partialTreeFixture("<PROJECT>"))
	if err != nil {
		t.Fatalf("partial projection error = %v", err)
	}
	entry := partial.Entries[0]
	if entry.Reliability != analysis.ReliabilityLowerBound ||
		entry.Classification != analysis.TargetImportedScene ||
		entry.ResourceID != "1_tree" ||
		entry.RawTarget != "res://models/tree.glb" ||
		entry.ResolutionReason != project.ResolutionResolved {
		t.Fatalf("partial entry = %#v", entry)
	}

	inherited, err := projectDependencyTree(inheritedTreeFixture("<PROJECT>"))
	if err != nil {
		t.Fatalf("inherited projection error = %v", err)
	}
	if inherited.Entries[0].Reliability != analysis.ReliabilityApproximate ||
		inherited.Entries[0].Kind != analysis.EdgeInheritance {
		t.Fatalf("inherited entry = %#v", inherited.Entries[0])
	}
}

func TestDependencyTreeProjectionRetainsEveryUnresolvedClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		classification  analysis.TargetClassification
		reason          project.ResolutionReason
		kind            analysis.EdgeKind
		raw             string
		resourceID      string
		wantTarget      string
		wantReliability analysis.Reliability
	}{
		{analysis.TargetMissingExternalResource, project.ResolutionEmpty, analysis.EdgeInstance, "", "missing", "resource:missing", analysis.ReliabilityLowerBound},
		{analysis.TargetUnresolvedPath, project.ResolutionMissing, analysis.EdgeInstance, "res://missing.tscn", "1", "res://missing.tscn", analysis.ReliabilityLowerBound},
		{analysis.TargetImportedScene, project.ResolutionResolved, analysis.EdgeInstance, "res://model.glb", "2", "res://model.glb", analysis.ReliabilityLowerBound},
		{analysis.TargetUnsupportedScene, project.ResolutionResolved, analysis.EdgeInstance, "res://old.scn", "3", "res://old.scn", analysis.ReliabilityLowerBound},
		{analysis.TargetSubResource, project.ResolutionEmpty, analysis.EdgeInstance, "SubResource(\"4\")", "4", "SubResource(\"4\")", analysis.ReliabilityLowerBound},
		{analysis.TargetPlaceholder, project.ResolutionEmpty, analysis.EdgeInstance, "", "5", "resource:5", analysis.ReliabilityLowerBound},
		{analysis.TargetUnavailableScene, project.ResolutionFilesystem, analysis.EdgeInstance, "/private/checkout/scene.tscn", "6", "resource:6", analysis.ReliabilityLowerBound},
		{analysis.TargetInheritedScene, project.ResolutionMissing, analysis.EdgeInheritance, "res://base.tscn", "7", "res://base.tscn", analysis.ReliabilityApproximate},
	}
	for _, test := range tests {
		test := test
		t.Run(string(test.classification), func(t *testing.T) {
			t.Parallel()

			result := rootOnlyTreeFixture("<PROJECT>")
			root := result.Inspect.Analysis.Graph.RootCanonical
			result.Inspect.Analysis.Graph.Edges = []analysis.GraphEdge{{
				FromCanonical:    root,
				FromDisplay:      "res://scenes/root.tscn",
				RawTarget:        test.raw,
				ResourceID:       test.resourceID,
				Kind:             test.kind,
				Classification:   test.classification,
				ResolutionReason: test.reason,
				Occurrences:      1,
			}}
			projected, err := projectDependencyTree(result)
			if err != nil {
				t.Fatalf("projectDependencyTree() error = %v", err)
			}
			entry := projected.Entries[0]
			if entry.Target != test.wantTarget || entry.Classification != test.classification ||
				entry.ResolutionReason != test.reason || entry.Reliability != test.wantReliability {
				t.Fatalf("entry = %#v", entry)
			}
			if test.classification == analysis.TargetUnavailableScene &&
				(entry.RawTarget != "" || strings.Contains(entry.Target, "private")) {
				t.Fatalf("unsafe unavailable target leaked: %#v", entry)
			}
		})
	}
}

func TestTreePresentationIsPortableOrderedAndNonMutating(t *testing.T) {
	t.Parallel()

	left := repeatedDiamondTreeFixture(`/tmp/checkout-one`)
	right := repeatedDiamondTreeFixture(`D:\work\checkout-two`)
	slices.Reverse(right.Inspect.Analysis.Graph.Nodes)
	slices.Reverse(right.Inspect.Analysis.Graph.Edges)
	before := snapshotJSON(t, left)

	leftText, err := Tree(left, Options{Version: "test"})
	if err != nil {
		t.Fatalf("left Tree() error = %v", err)
	}
	rightText, err := Tree(right, Options{Version: "test"})
	if err != nil {
		t.Fatalf("right Tree() error = %v", err)
	}
	leftJSON, err := TreeJSON(left, Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("left TreeJSON() error = %v", err)
	}
	rightJSON, err := TreeJSON(right, Options{Version: "test"})
	if err != nil {
		t.Fatalf("right TreeJSON() error = %v", err)
	}
	if leftText != rightText || leftJSON != rightJSON {
		t.Fatalf("portable tree documents differ\n--- text left ---\n%s--- text right ---\n%s\n--- JSON left ---\n%s--- JSON right ---\n%s", leftText, rightText, leftJSON, rightJSON)
	}
	if repeated, err := TreeJSON(left, Options{Version: "test"}); err != nil || repeated != leftJSON {
		t.Fatalf("repeated TreeJSON = %v / %v", repeated == leftJSON, err)
	}
	if after := snapshotJSON(t, left); !bytes.Equal(after, before) {
		t.Fatal("tree presentation mutated caller-owned evidence")
	}
	for _, forbidden := range []string{"checkout-one", "checkout-two", `\\`, "\x1b["} {
		if strings.Contains(leftText, forbidden) || strings.Contains(leftJSON, forbidden) {
			t.Fatalf("portable output contains %q", forbidden)
		}
	}
	assertTextFraming(t, leftText)
	assertJSONFraming(t, leftJSON)
}

func TestDependencyTreeProjectionRejectsInconsistentGraphs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*application.TreeResult)
	}{
		{name: "missing root", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.RootCanonical = canonicalTreePath("<PROJECT>", "scenes/missing.tscn")
		}},
		{name: "missing source", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Edges[0].FromCanonical = canonicalTreePath("<PROJECT>", "scenes/missing.tscn")
		}},
		{name: "missing resolved target", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Edges[0].ToCanonical = canonicalTreePath("<PROJECT>", "scenes/missing.tscn")
		}},
		{name: "zero occurrences", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Edges[0].Occurrences = 0
		}},
		{name: "overflow boundary remains representable", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Edges[0].Occurrences = math.MaxInt64
		}},
		{name: "duplicate portable node", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Nodes = append(result.Inspect.Analysis.Graph.Nodes, analysis.GraphNode{
				Canonical: canonicalTreePath("<PROJECT>", "scenes/duplicate.tscn"),
				Display:   "res://scenes/branch-b.tscn",
			})
			result.Inspect.Analysis.Graph.SceneDependencies++
			result.Inspect.Analysis.Summary.Metrics.SceneDependencies++
			result.Inspect.Analysis.Coverage.ParsedSceneFiles++
		}},
		{name: "unreachable node", mutate: func(result *application.TreeResult) {
			result.Inspect.Analysis.Graph.Nodes = append(result.Inspect.Analysis.Graph.Nodes, analysis.GraphNode{
				Canonical: canonicalTreePath("<PROJECT>", "scenes/orphan.tscn"),
				Display:   "res://scenes/orphan.tscn",
			})
			result.Inspect.Analysis.Graph.SceneDependencies++
			result.Inspect.Analysis.Summary.Metrics.SceneDependencies++
			result.Inspect.Analysis.Coverage.ParsedSceneFiles++
		}},
		{name: "resolved cycle", mutate: func(result *application.TreeResult) {
			root := result.Inspect.Analysis.Graph.RootCanonical
			leaf := canonicalTreePath("<PROJECT>", "scenes/leaf.tscn")
			result.Inspect.Analysis.Graph.Edges = append(result.Inspect.Analysis.Graph.Edges, analysis.GraphEdge{
				FromCanonical: leaf, FromDisplay: "res://scenes/leaf.tscn",
				ToCanonical: root, ToDisplay: "res://scenes/root.tscn", RawTarget: "res://scenes/root.tscn",
				Kind: analysis.EdgeInstance, Resolved: true,
				ResolutionReason: project.ResolutionResolved, Occurrences: 1,
			})
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := repeatedDiamondTreeFixture("<PROJECT>")
			test.mutate(&result)
			_, err := projectDependencyTree(result)
			if test.name == "overflow boundary remains representable" {
				if err != nil {
					t.Fatalf("MaxInt64 occurrence rejected: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("projectDependencyTree() error = nil")
			}
		})
	}
}

func TestTreeJSONProjectionMatchesTextPreorder(t *testing.T) {
	t.Parallel()

	result := repeatedDiamondTreeFixture("<PROJECT>")
	projected, err := projectDependencyTree(result)
	if err != nil {
		t.Fatalf("projectDependencyTree() error = %v", err)
	}
	rendered, err := TreeJSON(result, Options{Version: "test"})
	if err != nil {
		t.Fatalf("TreeJSON() error = %v", err)
	}
	var document struct {
		Tree struct {
			Entries []struct {
				Depth  int64  `json:"depth"`
				Target string `json:"target"`
			} `json:"entries"`
		} `json:"dependency_tree"`
	}
	if err := json.Unmarshal([]byte(rendered), &document); err != nil {
		t.Fatalf("decode TreeJSON: %v", err)
	}
	if len(document.Tree.Entries) != len(projected.Entries) {
		t.Fatalf("JSON/projected entries = %d / %d", len(document.Tree.Entries), len(projected.Entries))
	}
	for index, entry := range document.Tree.Entries {
		if entry.Depth != projected.Entries[index].Depth || entry.Target != projected.Entries[index].Target {
			t.Fatalf("JSON entry %d = %#v, projected %#v", index, entry, projected.Entries[index])
		}
	}
}

func TestTreeJSONSchemaRejectsInvalidTreeSemantics(t *testing.T) {
	t.Parallel()

	rendered, err := TreeJSON(partialTreeFixture("<PROJECT>"), Options{Version: "test"})
	if err != nil {
		t.Fatalf("TreeJSON() error = %v", err)
	}
	var base map[string]any
	decoder := json.NewDecoder(strings.NewReader(rendered))
	decoder.UseNumber()
	if err := decoder.Decode(&base); err != nil {
		t.Fatalf("decode TreeJSON: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing tree", mutate: func(document map[string]any) { delete(document, "dependency_tree") }},
		{name: "mixed policy", mutate: func(document map[string]any) { document["policy"] = map[string]any{} }},
		{name: "zero depth", mutate: func(document map[string]any) { treeJSONEntry(document)["depth"] = 0 }},
		{name: "zero occurrences", mutate: func(document map[string]any) { treeJSONEntry(document)["occurrences"] = 0 }},
		{name: "invalid kind", mutate: func(document map[string]any) { treeJSONEntry(document)["kind"] = "future" }},
		{name: "invalid reliability", mutate: func(document map[string]any) { treeJSONEntry(document)["reliability"] = "future" }},
		{name: "missing unresolved classification", mutate: func(document map[string]any) { delete(treeJSONEntry(document), "classification") }},
		{name: "unresolved back-reference", mutate: func(document map[string]any) { treeJSONEntry(document)["back_reference"] = true }},
		{name: "canonical target", mutate: func(document map[string]any) {
			entry := treeJSONEntry(document)
			entry["resolved"] = true
			entry["target"] = "/private/scene.tscn"
			delete(entry, "classification")
			delete(entry, "resolution_reason")
			entry["reliability"] = "exact"
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			document := cloneDocumentMap(t, base)
			test.mutate(document)
			if err := reportSchema(t).Validate(document); err == nil {
				t.Fatalf("schema accepted invalid tree document: %#v", document)
			}
		})
	}
}

func treeJSONEntry(document map[string]any) map[string]any {
	tree := document["dependency_tree"].(map[string]any)
	return tree["entries"].([]any)[0].(map[string]any)
}

func assertTextFraming(t *testing.T, rendered string) {
	t.Helper()
	if strings.Contains(rendered, "\r") || !strings.HasSuffix(rendered, "\n") || strings.HasSuffix(rendered, "\n\n") {
		t.Fatalf("invalid text framing: %q", rendered)
	}
	if strings.Contains(rendered, "\x1b[") {
		t.Fatalf("plain text contains ANSI: %q", rendered)
	}
}

func rootOnlyTreeFixture(projectRoot string) application.TreeResult {
	inspect := completeInspect()
	root := canonicalTreePath(projectRoot, "scenes/root.tscn")
	inspect.Project = project.Root{Directory: projectRoot, ProjectFile: canonicalTreePath(projectRoot, "project.godot")}
	inspect.Scene = project.ResolvedPath{Canonical: root, Display: "res://scenes/root.tscn", Original: "res://scenes/root.tscn"}
	inspect.ConfigPresent = false
	inspect.ConfigSource = config.Source{}
	inspect.Analysis.Summary.Metrics.SceneDependencies = 0
	inspect.Analysis.Summary.Contributions[0].SceneCanonical = root
	inspect.Analysis.Summary.Contributions[0].SceneDisplay = "res://scenes/root.tscn"
	inspect.Analysis.Coverage.ParsedSceneFiles = 1
	inspect.Analysis.Graph = analysis.DependencyGraph{
		RootCanonical: root,
		RootDisplay:   "res://scenes/root.tscn",
		Nodes:         []analysis.GraphNode{{Canonical: root, Display: "res://scenes/root.tscn"}},
	}

	return application.TreeResult{Inspect: inspect}
}

func completeTreeFixture(projectRoot string) application.TreeResult {
	result := rootOnlyTreeFixture(projectRoot)
	root := result.Inspect.Analysis.Graph.RootCanonical
	branch := canonicalTreePath(projectRoot, "scenes/branch.tscn")
	leaf := canonicalTreePath(projectRoot, "scenes/leaf.tscn")
	setTreeGraph(&result, []analysis.GraphNode{
		{Canonical: root, Display: "res://scenes/root.tscn"},
		{Canonical: branch, Display: "res://scenes/branch.tscn"},
		{Canonical: leaf, Display: "res://scenes/leaf.tscn"},
	}, []analysis.GraphEdge{
		resolvedTreeEdge(root, "res://scenes/root.tscn", branch, "res://scenes/branch.tscn", analysis.EdgeInstance, 1),
		resolvedTreeEdge(branch, "res://scenes/branch.tscn", leaf, "res://scenes/leaf.tscn", analysis.EdgeInstance, 1),
	})
	return result
}

func repeatedDiamondTreeFixture(projectRoot string) application.TreeResult {
	result := rootOnlyTreeFixture(projectRoot)
	root := result.Inspect.Analysis.Graph.RootCanonical
	branchB := canonicalTreePath(projectRoot, "scenes/branch-b.tscn")
	branchC := canonicalTreePath(projectRoot, "scenes/branch-c.tscn")
	shared := canonicalTreePath(projectRoot, "scenes/shared.tscn")
	leaf := canonicalTreePath(projectRoot, "scenes/leaf.tscn")
	setTreeGraph(&result, []analysis.GraphNode{
		{Canonical: shared, Display: "res://scenes/shared.tscn"},
		{Canonical: root, Display: "res://scenes/root.tscn"},
		{Canonical: leaf, Display: "res://scenes/leaf.tscn"},
		{Canonical: branchC, Display: "res://scenes/branch-c.tscn"},
		{Canonical: branchB, Display: "res://scenes/branch-b.tscn"},
	}, []analysis.GraphEdge{
		resolvedTreeEdge(branchC, "res://scenes/branch-c.tscn", shared, "res://scenes/shared.tscn", analysis.EdgeInstance, 1),
		resolvedTreeEdge(root, "res://scenes/root.tscn", branchC, "res://scenes/branch-c.tscn", analysis.EdgeInstance, 1),
		resolvedTreeEdge(shared, "res://scenes/shared.tscn", leaf, "res://scenes/leaf.tscn", analysis.EdgeInstance, 1),
		resolvedTreeEdge(branchB, "res://scenes/branch-b.tscn", shared, "res://scenes/shared.tscn", analysis.EdgeInstance, 1),
		resolvedTreeEdge(root, "res://scenes/root.tscn", branchB, "res://scenes/branch-b.tscn", analysis.EdgeInstance, 100),
	})
	return result
}

func partialTreeFixture(projectRoot string) application.TreeResult {
	result := rootOnlyTreeFixture(projectRoot)
	root := result.Inspect.Analysis.Graph.RootCanonical
	result.Inspect.Analysis.Status = analysis.AnalysisPartial
	result.Inspect.Analysis.Reliability = analysis.ReliabilityLowerBound
	result.Inspect.Analysis.Coverage.UnresolvedSceneInstances = 3
	result.Inspect.Analysis.Summary.Contributions[0].Reliability = analysis.ReliabilityLowerBound
	result.Inspect.Analysis.Diagnostics = []diagnostic.Diagnostic{{
		Code: diagnostic.CodeImportedScene, Severity: diagnostic.SeverityWarning,
		Message: "imported PackedScene cannot be expanded statically",
		File:    "res://models/tree.glb", Occurrences: 3,
	}}
	result.Inspect.Analysis.Graph.Edges = []analysis.GraphEdge{{
		FromCanonical: root, FromDisplay: "res://scenes/root.tscn",
		RawTarget: "res://models/tree.glb", ResourceID: "1_tree",
		Kind: analysis.EdgeInstance, Resolved: false,
		Classification: analysis.TargetImportedScene, ResolutionReason: project.ResolutionResolved,
		Occurrences: 3,
	}}
	return result
}

func inheritedTreeFixture(projectRoot string) application.TreeResult {
	result := rootOnlyTreeFixture(projectRoot)
	root := result.Inspect.Analysis.Graph.RootCanonical
	base := canonicalTreePath(projectRoot, "scenes/base.tscn")
	setTreeGraph(&result, []analysis.GraphNode{
		{Canonical: root, Display: "res://scenes/root.tscn"},
		{Canonical: base, Display: "res://scenes/base.tscn"},
	}, []analysis.GraphEdge{
		resolvedTreeEdge(root, "res://scenes/root.tscn", base, "res://scenes/base.tscn", analysis.EdgeInheritance, 1),
	})
	result.Inspect.Analysis.Status = analysis.AnalysisPartial
	result.Inspect.Analysis.Reliability = analysis.ReliabilityApproximate
	result.Inspect.Analysis.Coverage.InheritedScenes = 1
	result.Inspect.Analysis.Summary.Contributions[0].Reliability = analysis.ReliabilityApproximate
	result.Inspect.Analysis.Diagnostics = []diagnostic.Diagnostic{{
		Code: diagnostic.CodeInheritedScene, Severity: diagnostic.SeverityWarning,
		Message: "inherited-scene overrides make expanded metrics approximate",
		File:    "res://scenes/root.tscn", Resource: "res://scenes/base.tscn", Occurrences: 1,
	}}
	return result
}

func setTreeGraph(result *application.TreeResult, nodes []analysis.GraphNode, edges []analysis.GraphEdge) {
	result.Inspect.Analysis.Graph.Nodes = nodes
	result.Inspect.Analysis.Graph.Edges = edges
	dependencies := int64(len(nodes) - 1)
	result.Inspect.Analysis.Graph.SceneDependencies = dependencies
	result.Inspect.Analysis.Summary.Metrics.SceneDependencies = dependencies
	result.Inspect.Analysis.Coverage.ParsedSceneFiles = int64(len(nodes))
}

func resolvedTreeEdge(
	fromCanonical, fromDisplay, toCanonical, toDisplay string,
	kind analysis.EdgeKind,
	occurrences int64,
) analysis.GraphEdge {
	return analysis.GraphEdge{
		FromCanonical:    fromCanonical,
		FromDisplay:      fromDisplay,
		ToCanonical:      toCanonical,
		ToDisplay:        toDisplay,
		RawTarget:        toDisplay,
		ResourceID:       strings.TrimSuffix(strings.TrimPrefix(toDisplay, "res://scenes/"), ".tscn"),
		Kind:             kind,
		Resolved:         true,
		ResolutionReason: project.ResolutionResolved,
		Occurrences:      occurrences,
	}
}

func canonicalTreePath(root, relative string) string {
	separator := "/"
	if strings.Contains(root, `\`) {
		separator = `\`
		relative = strings.ReplaceAll(relative, "/", `\`)
	}
	return strings.TrimRight(root, `/\`) + separator + relative
}
