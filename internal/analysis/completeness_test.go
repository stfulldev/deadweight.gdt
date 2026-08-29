package analysis

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func TestAnalysisStatusReliabilityAndCoverageValidation(t *testing.T) {
	for _, status := range []AnalysisStatus{AnalysisComplete, AnalysisPartial} {
		if !status.Valid() {
			t.Errorf("%q.Valid() = false", status)
		}
	}
	if AnalysisStatus("unknown").Valid() {
		t.Fatal("unknown analysis status is valid")
	}
	for _, reliability := range []Reliability{
		ReliabilityExact,
		ReliabilityLowerBound,
		ReliabilityApproximate,
	} {
		if !reliability.Valid() {
			t.Errorf("%q.Valid() = false", reliability)
		}
	}
	if Reliability("unknown").Valid() {
		t.Fatal("unknown reliability is valid")
	}

	valid := Coverage{
		ResolvedSceneInstances: 3, UnresolvedSceneInstances: 2,
		ParsedSceneFiles: 4, InheritedScenes: 1,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid coverage error = %v", err)
	}
	tests := []Coverage{
		{ResolvedSceneInstances: -1},
		{UnresolvedSceneInstances: -1},
		{ParsedSceneFiles: -1},
		{InheritedScenes: -1},
		{UnresolvedSceneInstances: 1, InheritedScenes: 2},
	}
	for _, coverage := range tests {
		if err := coverage.Validate(); err == nil {
			t.Errorf("Coverage(%#v).Validate() error = nil", coverage)
		}
	}

	validPairs := []completionResult{
		{Status: AnalysisComplete, Reliability: ReliabilityExact},
		{Status: AnalysisPartial, Reliability: ReliabilityLowerBound},
		{Status: AnalysisPartial, Reliability: ReliabilityApproximate},
	}
	for _, result := range validPairs {
		if err := validateCompletion(result); err != nil {
			t.Errorf("validateCompletion(%#v) error = %v", result, err)
		}
	}
	invalidPairs := []completionResult{
		{Status: "unknown", Reliability: ReliabilityExact},
		{Status: AnalysisComplete, Reliability: "unknown"},
		{Status: AnalysisComplete, Reliability: ReliabilityLowerBound},
		{Status: AnalysisPartial, Reliability: ReliabilityExact},
	}
	for _, result := range invalidPairs {
		if err := validateCompletion(result); err == nil {
			t.Errorf("validateCompletion(%#v) error = nil", result)
		}
	}
}

func TestFinalizeCompletenessReturnsCompleteExactCoverage(t *testing.T) {
	summary := ExpandedSummary{
		Metrics:  metrics.Values{SceneInstances: 2},
		Coverage: SceneInstanceCoverage{Resolved: 2},
		ExternalResources: []ResourceIdentity{
			{Resolved: true, Canonical: "/project/texture.png"},
			{Resolved: true, Canonical: "/project/material.tres"},
		},
	}
	result, err := finalizeCompleteness(summary, completionTestGraph(), 3)
	if err != nil {
		t.Fatalf("finalizeCompleteness() error = %v", err)
	}
	want := completionResult{
		Status:      AnalysisComplete,
		Reliability: ReliabilityExact,
		Coverage: Coverage{
			ResolvedSceneInstances: 2,
			ParsedSceneFiles:       3,
		},
		Diagnostics: []diagnostic.Diagnostic{},
	}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("result = %#v, want %#v", result, want)
	}
}

func TestFinalizeCompletenessClassifiesEveryPartialBranch(t *testing.T) {
	root := "/project/root.tscn"
	display := "res://root.tscn"
	unresolved := []UnresolvedInstance{
		testUnresolved(root, display, TargetMissingExternalResource, "", "missing_id", "", 1, 10),
		testUnresolved(root, display, TargetImportedScene, project.ResolutionResolved, "imported", "model.glb", 3, 20),
		testUnresolved(root, display, TargetUnsupportedScene, project.ResolutionResolved, "script", "logic.gd", 1, 30),
		testUnresolved(root, display, TargetSubResource, "", "embedded", "Scene_1", 1, 40),
		testUnresolved(root, display, TargetPlaceholder, "", "", "res://later.tscn", 2, 50),
		testUnresolved(root, display, TargetUnavailableScene, project.ResolutionFilesystem, "unreadable", "locked.tscn", 1, 60),
		testUnresolved(root, display, TargetUnresolvedPath, project.ResolutionUIDOnly, "uid", "uid://abc", 1, 70),
		testUnresolved(root, display, TargetUnresolvedPath, project.ResolutionMissing, "missing", "missing.tscn", 1, 80),
	}
	summary := ExpandedSummary{
		Metrics:    metrics.Values{SceneInstances: 13},
		Coverage:   SceneInstanceCoverage{Unresolved: 13},
		Unresolved: unresolved,
		InheritedTargets: []InheritedTarget{{
			Classification: TargetInheritedScene, DeclaringScene: root, DeclaringDisplay: display,
			TargetCanonical: "/project/inherited.tscn", TargetDisplay: "res://inherited.tscn",
			Occurrences: 2, MountPosition: tscn.Position{Line: 90, Column: 1},
		}},
		ParentFindings: []SceneParentFinding{{
			DeclaringScene: root, DeclaringDisplay: display, Occurrences: 4,
			Finding: ParentFinding{
				Kind: ParentMissing, NodeName: "Lost", NodePath: "Lost", Parent: "Unknown",
				Position: tscn.Position{Line: 100, Column: 1},
			},
		}},
		DepthPartial: true,
	}

	result, err := finalizeCompleteness(summary, completionTestGraph(), 4)
	if err != nil {
		t.Fatalf("finalizeCompleteness() error = %v", err)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityApproximate {
		t.Fatalf("status/reliability = %q/%q", result.Status, result.Reliability)
	}
	if result.Coverage != (Coverage{
		UnresolvedSceneInstances: 13,
		ParsedSceneFiles:         4,
		InheritedScenes:          2,
	}) {
		t.Fatalf("coverage = %#v", result.Coverage)
	}
	wantCodes := map[diagnostic.Code]bool{
		diagnostic.CodeUnresolvedSceneInstance: false,
		diagnostic.CodeImportedScene:           false,
		diagnostic.CodeInheritedScene:          false,
		diagnostic.CodeUnavailableResource:     false,
		diagnostic.CodeInstancePlaceholder:     false,
		diagnostic.CodeUnsupportedResourcePath: false,
		diagnostic.CodeUnsupportedParent:       false,
	}
	for index, item := range result.Diagnostics {
		if err := item.Validate(); err != nil {
			t.Errorf("diagnostic %d invalid: %v", index, err)
		}
		if index > 0 && diagnosticLess(item, result.Diagnostics[index-1]) {
			t.Errorf("diagnostics are not sorted: %#v", result.Diagnostics)
		}
		if _, expected := wantCodes[item.Code]; expected {
			wantCodes[item.Code] = true
		}
	}
	for code, seen := range wantCodes {
		if !seen {
			t.Errorf("diagnostic code %q not emitted: %#v", code, result.Diagnostics)
		}
	}
}

func TestFinalizeCompletenessGroupsRepeatedEvidenceAndOwnsResults(t *testing.T) {
	root := "/project/root.tscn"
	display := "res://root.tscn"
	summary := ExpandedSummary{
		Metrics:  metrics.Values{SceneInstances: 100},
		Coverage: SceneInstanceCoverage{Unresolved: 100},
		Unresolved: []UnresolvedInstance{
			testUnresolved(root, display, TargetImportedScene, project.ResolutionResolved, "model", "model.glb", 40, 20),
			testUnresolved(root, display, TargetImportedScene, project.ResolutionResolved, "model", "model.glb", 60, 10),
		},
	}
	first, err := finalizeCompleteness(summary, completionTestGraph(), 1)
	if err != nil {
		t.Fatalf("first finalizeCompleteness() error = %v", err)
	}
	if len(first.Diagnostics) != 1 || first.Diagnostics[0].Code != diagnostic.CodeImportedScene ||
		first.Diagnostics[0].Occurrences != 100 || first.Diagnostics[0].Line != 10 {
		t.Fatalf("grouped diagnostics = %#v", first.Diagnostics)
	}
	first.Diagnostics[0].Message = "mutated"
	second, err := finalizeCompleteness(summary, completionTestGraph(), 1)
	if err != nil {
		t.Fatalf("second finalizeCompleteness() error = %v", err)
	}
	if second.Diagnostics[0].Message == "mutated" {
		t.Fatalf("caller mutation leaked into later result: %#v", second.Diagnostics)
	}
}

func TestFinalizeCompletenessCountsRepeatedInheritance(t *testing.T) {
	summary := ExpandedSummary{
		Metrics:  metrics.Values{SceneInstances: 100},
		Coverage: SceneInstanceCoverage{Unresolved: 100},
		InheritedTargets: []InheritedTarget{{
			Classification:  TargetInheritedScene,
			DeclaringScene:  "/project/root.tscn",
			TargetCanonical: "/project/inherited.tscn",
			Occurrences:     100,
		}},
	}
	result, err := finalizeCompleteness(summary, completionTestGraph(), 3)
	if err != nil {
		t.Fatalf("finalizeCompleteness() error = %v", err)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityApproximate ||
		result.Coverage != (Coverage{
			UnresolvedSceneInstances: 100,
			ParsedSceneFiles:         3,
			InheritedScenes:          100,
		}) || len(result.Diagnostics) != 1 || result.Diagnostics[0].Occurrences != 100 {
		t.Fatalf("result = %#v", result)
	}
}

func TestResourceIdentityUnionKeepsFrozenUniquenessKey(t *testing.T) {
	resolved := ResourceIdentity{Resolved: true, Canonical: "/project/shared.res"}
	unresolvedMissing := ResourceIdentity{
		DeclaringScene: "/project/root.tscn", ResourceID: "asset", RawPath: "shared.res",
		ResolutionReason: project.ResolutionMissing,
	}
	unresolvedOutside := unresolvedMissing
	unresolvedOutside.ResolutionReason = project.ResolutionOutsideProject
	otherDeclaration := unresolvedMissing
	otherDeclaration.DeclaringScene = "/project/child.tscn"

	got := mergeResourceIdentities(
		[]ResourceIdentity{resolved, unresolvedOutside, otherDeclaration},
		[]ResourceIdentity{resolved, unresolvedMissing},
	)
	if len(got) != 3 {
		t.Fatalf("resource union = %#v, want three identities", got)
	}
	matchingTuple := 0
	for _, identity := range got {
		if identity.DeclaringScene == unresolvedMissing.DeclaringScene &&
			identity.ResourceID == unresolvedMissing.ResourceID &&
			identity.RawPath == unresolvedMissing.RawPath {
			matchingTuple++
		}
	}
	if matchingTuple != 1 {
		t.Fatalf("same unresolved tuple occurrences = %d, want 1 in %#v", matchingTuple, got)
	}
}

func TestFinalizeCompletenessClassifiesUnresolvedResourceReasons(t *testing.T) {
	reasons := []project.ResolutionReason{
		project.ResolutionEmpty,
		project.ResolutionMissing,
		project.ResolutionOutsideProject,
		project.ResolutionFilesystem,
		project.ResolutionInvalidDeclaringScene,
		project.ResolutionUnsupportedTarget,
		project.ResolutionUIDOnly,
		project.ResolutionUserData,
	}
	resources := make([]ResourceIdentity, 0, len(reasons))
	for index, reason := range reasons {
		resources = append(resources, ResourceIdentity{
			DeclaringScene:   "/project/root.tscn",
			ResourceID:       string(rune('a' + index)),
			RawPath:          string(reason),
			ResolutionReason: reason,
		})
	}
	result, err := finalizeCompleteness(
		ExpandedSummary{ExternalResources: resources},
		completionTestGraph(),
		1,
	)
	if err != nil {
		t.Fatalf("finalizeCompleteness() error = %v", err)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityLowerBound ||
		len(result.Diagnostics) != len(reasons) {
		t.Fatalf("result = %#v", result)
	}
	counts := map[diagnostic.Code]int{}
	for _, item := range result.Diagnostics {
		counts[item.Code]++
	}
	if counts[diagnostic.CodeUnavailableResource] != 6 ||
		counts[diagnostic.CodeUnsupportedResourcePath] != 2 {
		t.Fatalf("diagnostic code counts = %#v", counts)
	}
}

func TestFinalizeCompletenessRejectsInvalidEvidenceTransactionally(t *testing.T) {
	root := "/project/root.tscn"
	display := "res://root.tscn"
	tests := []struct {
		name    string
		summary ExpandedSummary
		parsed  int64
	}{
		{name: "negative coverage", summary: ExpandedSummary{Coverage: SceneInstanceCoverage{Resolved: -1}}, parsed: 1},
		{name: "coverage mismatch", summary: ExpandedSummary{Metrics: metrics.Values{SceneInstances: 1}}, parsed: 1},
		{name: "unknown classification", summary: ExpandedSummary{
			Metrics: metrics.Values{SceneInstances: 1}, Coverage: SceneInstanceCoverage{Unresolved: 1},
			Unresolved: []UnresolvedInstance{{Classification: "unknown", DeclaringScene: root, Occurrences: 1}},
		}, parsed: 1},
		{name: "invalid resource reason", summary: ExpandedSummary{ExternalResources: []ResourceIdentity{{
			DeclaringScene: root, ResourceID: "asset", RawPath: "asset.res", ResolutionReason: "unknown",
		}}}, parsed: 1},
		{name: "diagnostic grouping overflow", summary: ExpandedSummary{
			Metrics:  metrics.Values{SceneInstances: math.MaxInt64},
			Coverage: SceneInstanceCoverage{Unresolved: math.MaxInt64},
			Unresolved: []UnresolvedInstance{
				testUnresolved(root, display, TargetImportedScene, project.ResolutionResolved, "model", "model.glb", math.MaxInt64, 1),
				testUnresolved(root, display, TargetImportedScene, project.ResolutionResolved, "model", "model.glb", 1, 2),
			},
		}, parsed: 1},
		{name: "inherited coverage overflow", summary: ExpandedSummary{
			Metrics:  metrics.Values{SceneInstances: math.MaxInt64},
			Coverage: SceneInstanceCoverage{Unresolved: math.MaxInt64},
			InheritedTargets: []InheritedTarget{
				{Classification: TargetInheritedScene, DeclaringScene: root, Occurrences: math.MaxInt64},
				{Classification: TargetInheritedScene, DeclaringScene: root, Occurrences: 1},
			},
		}, parsed: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := finalizeCompleteness(test.summary, completionTestGraph(), test.parsed)
			if err == nil || !reflect.DeepEqual(result, completionResult{}) {
				t.Fatalf("finalizeCompleteness() = %#v, %v; want zero result and error", result, err)
			}
			if test.name == "diagnostic grouping overflow" || test.name == "inherited coverage overflow" {
				var overflow *OverflowError
				if !errors.As(err, &overflow) {
					t.Fatalf("error = %T %v, want *OverflowError", err, err)
				}
			}
		})
	}
}

func TestRecursiveAnalyzerKeepsResolvedOrdinaryResourcesComplete(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	paths := []string{"data.tres", "texture.png", "material.tres", "logic.gd", "sound.ogg"}
	resolver := &memoryResolver{results: make(map[string]project.Resolution)}
	for _, raw := range paths {
		resolver.results[resolverKey(root.Canonical, raw)] = resolvedResource(raw, testResourcePath(rootDir, raw))
	}
	loader := &memorySceneEffects{sources: map[string]string{root.Canonical: `[gd_scene format=3]
[ext_resource type="Resource" path="data.tres" id="1"]
[ext_resource type="Texture2D" path="texture.png" id="2"]
[ext_resource type="Material" path="material.tres" id="3"]
[ext_resource type="Script" path="logic.gd" id="4"]
[ext_resource type="AudioStream" path="sound.ogg" id="5"]
[node name="Root" type="Node3D"]
`}, errors: map[string]error{}}
	analyzer := newTestRecursiveAnalyzer(t, resolver, loader)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Status != AnalysisComplete || result.Reliability != ReliabilityExact ||
		result.Coverage != (Coverage{ParsedSceneFiles: 1}) || len(result.Diagnostics) != 0 ||
		result.Summary.Metrics.ExternalResources != 5 {
		t.Fatalf("result = %#v", result)
	}
}

func TestRecursiveAnalyzerMarksMissingOrdinaryResourcePartial(t *testing.T) {
	rootDir := t.TempDir()
	root := testScenePath(rootDir, "root.tscn")
	loader := &memorySceneEffects{sources: map[string]string{root.Canonical: `[gd_scene format=3]
[ext_resource type="Texture2D" path="missing.png" id="1_texture"]
[node name="Root" type="Node3D"]
`}, errors: map[string]error{}}
	analyzer := newTestRecursiveAnalyzer(
		t,
		&memoryResolver{results: map[string]project.Resolution{}},
		loader,
	)

	result, err := analyzer.Analyze(root)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if result.Status != AnalysisPartial || result.Reliability != ReliabilityLowerBound ||
		result.Coverage != (Coverage{ParsedSceneFiles: 1}) || len(result.Diagnostics) != 1 ||
		result.Diagnostics[0].Code != diagnostic.CodeUnavailableResource ||
		result.Diagnostics[0].Occurrences != 1 {
		t.Fatalf("result = %#v", result)
	}
	if len(result.Summary.ExternalResources) != 1 ||
		result.Summary.ExternalResources[0].ResolutionReason != project.ResolutionMissing {
		t.Fatalf("unresolved resource evidence = %#v", result.Summary.ExternalResources)
	}
}

func completionTestGraph() DependencyGraph {
	return DependencyGraph{
		RootCanonical: "/project/root.tscn",
		RootDisplay:   "res://root.tscn",
		Nodes: []GraphNode{{
			Canonical: "/project/root.tscn",
			Display:   "res://root.tscn",
		}},
	}
}

func testUnresolved(
	declaring string,
	display string,
	classification TargetClassification,
	reason project.ResolutionReason,
	resourceID string,
	rawTarget string,
	occurrences int64,
	line int,
) UnresolvedInstance {
	return UnresolvedInstance{
		Classification: classification, ResolutionReason: reason,
		DeclaringScene: declaring, DeclaringDisplay: display,
		ResourceID: resourceID, RawTarget: rawTarget,
		MountPath: "Mount", Position: tscn.Position{Line: line, Column: 1},
		Occurrences: occurrences,
	}
}

func diagnosticLess(left, right diagnostic.Diagnostic) bool {
	if left.Severity != right.Severity {
		return left.Severity < right.Severity
	}
	if left.Code != right.Code {
		return left.Code < right.Code
	}
	if left.File != right.File {
		return left.File < right.File
	}
	if left.Line != right.Line {
		return left.Line < right.Line
	}
	if left.Column != right.Column {
		return left.Column < right.Column
	}
	if left.Resource != right.Resource {
		return left.Resource < right.Resource
	}
	return left.Message < right.Message
}
