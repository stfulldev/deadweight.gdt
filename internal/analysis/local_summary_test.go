package analysis_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func TestBuildLocalSummaryComputesOrdinaryDepthsIndependentOfOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		nodes string
	}{
		{
			name: "parent first",
			nodes: `[node name="Root" type="Node3D"]
[node name="Arm" type="Node3D" parent="."]
[node name="Hand" type="Node3D" parent="Arm"]
[node name="Finger" type="Node3D" parent="Arm/Hand"]`,
		},
		{
			name: "children before parents",
			nodes: `[node name="Root" type="Node3D"]
[node name="Finger" type="Node3D" parent="Arm/Hand"]
[node name="Hand" type="Node3D" parent="Arm"]
[node name="Arm" type="Node3D" parent="."]`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			summary := parseSummary(t, "[gd_scene format=3]\n"+test.nodes+"\n")
			if summary.Metrics != (metrics.Values{Nodes: 4, TreeDepth: 4}) {
				t.Fatalf("Metrics = %#v, want four nodes at depth four", summary.Metrics)
			}
			if summary.DepthPartial || len(summary.Findings) != 0 {
				t.Fatalf("partial/findings = %v/%#v, want exact", summary.DepthPartial, summary.Findings)
			}

			wantDepths := map[string]int64{"Root": 1, "Arm": 2, "Hand": 3, "Finger": 4}
			for _, node := range summary.Nodes {
				want, exists := wantDepths[node.Name]
				if !exists || !node.Depth.Known || node.Depth.Value != want {
					t.Errorf("node %q depth = %#v, want %d", node.Name, node.Depth, want)
				}
				delete(wantDepths, node.Name)
			}
			if len(wantDepths) != 0 {
				t.Fatalf("missing node depths: %#v", wantDepths)
			}
		})
	}
}

func TestBuildLocalSummaryReportsUnsupportedParentsWithoutGuessing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parent   string
		wantKind analysis.ParentFindingKind
		wantPath string
	}{
		{name: "missing", parent: "Missing", wantKind: analysis.ParentMissing, wantPath: "Missing/Child"},
		{name: "parent traversal", parent: "..", wantKind: analysis.ParentInvalid},
		{name: "nested traversal", parent: "Arm/../Hand", wantKind: analysis.ParentInvalid},
		{name: "absolute", parent: "/root/Arm", wantKind: analysis.ParentInvalid},
		{name: "repeated separator", parent: "Arm//Hand", wantKind: analysis.ParentInvalid},
		{name: "non-root dot segment", parent: "./Arm", wantKind: analysis.ParentInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input := "[gd_scene format=3]\n[node name=\"Root\" type=\"Node3D\"]\n" +
				"[node name=\"Child\" type=\"MeshInstance3D\" parent=\"" + test.parent + "\"]\n"
			summary := parseSummary(t, input)

			if summary.Metrics != (metrics.Values{Nodes: 2, TreeDepth: 1, MeshInstances: 1}) {
				t.Errorf("Metrics = %#v, want known counts with root-only depth", summary.Metrics)
			}
			if !summary.DepthPartial || len(summary.Findings) != 1 {
				t.Fatalf("partial/findings = %v/%#v, want one finding", summary.DepthPartial, summary.Findings)
			}
			finding := summary.Findings[0]
			if finding.Kind != test.wantKind || finding.NodeName != "Child" || finding.NodePath != test.wantPath || finding.Parent != test.parent || finding.Position.Line != 3 {
				t.Errorf("Finding = %#v", finding)
			}
			child := findOrdinaryNode(t, summary, "Child")
			if child.Depth.Known {
				t.Errorf("Child depth = %#v, want unknown", child.Depth)
			}
		})
	}
}

func TestBuildLocalSummaryPropagatesUnknownDepthWithoutCascadedFinding(t *testing.T) {
	t.Parallel()

	summary := parseSummary(t, `[gd_scene format=3]
[node name="Root" type="Node3D"]
[node name="UnknownParent" type="Node3D" parent="Missing"]
[node name="Descendant" type="Node3D" parent="Missing/UnknownParent"]
`)

	if !summary.DepthPartial || len(summary.Findings) != 1 || summary.Findings[0].Kind != analysis.ParentMissing {
		t.Fatalf("partial/findings = %v/%#v", summary.DepthPartial, summary.Findings)
	}
	if findOrdinaryNode(t, summary, "UnknownParent").Depth.Known || findOrdinaryNode(t, summary, "Descendant").Depth.Known {
		t.Fatal("unknown depth was guessed for an affected node")
	}
	if summary.Metrics.Nodes != 3 || summary.Metrics.TreeDepth != 1 {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
}

func TestBuildLocalSummaryRejectsAmbiguousParentIdentity(t *testing.T) {
	t.Parallel()

	summary := parseSummary(t, `[gd_scene format=3]
[node name="Root" type="Node3D"]
[node name="Duplicate" type="Node3D" parent="."]
[node name="Duplicate" type="Node3D" parent="."]
[node name="Child" type="Node3D" parent="Duplicate"]
`)

	if !summary.DepthPartial || len(summary.Findings) != 1 || summary.Findings[0].Kind != analysis.ParentAmbiguous {
		t.Fatalf("partial/findings = %v/%#v", summary.DepthPartial, summary.Findings)
	}
	if findOrdinaryNode(t, summary, "Child").Depth.Known {
		t.Fatal("ambiguous child depth was guessed")
	}
	if summary.Metrics.Nodes != 4 || summary.Metrics.TreeDepth != 2 {
		t.Fatalf("Metrics = %#v", summary.Metrics)
	}
}

func TestBuildLocalSummaryClassifiesMountKindsWithoutOrdinaryRootContribution(t *testing.T) {
	t.Parallel()

	summary := parseSummary(t, `[gd_scene format=3]
[ext_resource type="PackedScene" path="res://child.tscn" id="1_scene"]
[ext_resource type="Script" path="res://logic.gd" id="2_script"]
[node name="Root" type="Node3D"]
[node name="Candidate" parent="." instance=ExtResource("1_scene")]
[node name="WrongType" parent="." instance=ExtResource("2_script")]
[node name="MissingID" parent="." instance=ExtResource("3_missing")]
[node name="Embedded" parent="." instance=SubResource("Scene_1")]
[node name="Placeholder" parent="." instance_placeholder="res://later.tscn"]
`)

	wantMetrics := metrics.Values{Nodes: 1, TreeDepth: 2, SceneInstances: 5}
	if summary.Metrics != wantMetrics {
		t.Fatalf("Metrics = %#v, want %#v", summary.Metrics, wantMetrics)
	}
	wantKinds := []analysis.MountKind{
		analysis.MountPackedSceneCandidate,
		analysis.MountExternalResource,
		analysis.MountExternalResource,
		analysis.MountSubResource,
		analysis.MountPlaceholder,
	}
	if len(summary.Mounts) != len(wantKinds) {
		t.Fatalf("Mounts = %#v", summary.Mounts)
	}
	for index, mount := range summary.Mounts {
		if mount.Kind != wantKinds[index] || !mount.Depth.Known || mount.Depth.Value != 2 {
			t.Errorf("Mounts[%d] = %#v", index, mount)
		}
	}
	if candidate := summary.Mounts[0].Candidate; candidate == nil || candidate.ID != "1_scene" || candidate.Path != "res://child.tscn" {
		t.Fatalf("Candidate = %#v", candidate)
	}
	if summary.Mounts[1].Candidate != nil || summary.Mounts[2].Candidate != nil || summary.Mounts[3].Candidate != nil || summary.Mounts[4].Candidate != nil {
		t.Fatal("non-PackedScene reference was classified as a candidate")
	}
}

func TestBuildLocalSummarySeparatesInheritedRootOverridesAndLocalAdditions(t *testing.T) {
	t.Parallel()

	summary := parseSummary(t, `[gd_scene format=3]
[ext_resource type="PackedScene" path="res://base.tscn" id="1_base"]
[ext_resource type="PackedScene" path="res://weapon.tscn" id="2_weapon"]
[node name="Inherited" instance=ExtResource("1_base")]
[node name="Body" parent="."]
[node name="Hat" type="MeshInstance3D" parent="Body"]
[node name="Weapon" parent="Body" instance=ExtResource("2_weapon")]
`)

	wantMetrics := metrics.Values{Nodes: 1, TreeDepth: 3, SceneInstances: 1, MeshInstances: 1}
	if summary.Metrics != wantMetrics {
		t.Fatalf("Metrics = %#v, want %#v", summary.Metrics, wantMetrics)
	}
	if summary.InheritedRoot == nil || summary.InheritedRoot.Reference.ID != "1_base" || summary.InheritedRoot.Candidate == nil || summary.InheritedRoot.Depth != (analysis.OptionalDepth{Value: 1, Known: true}) {
		t.Fatalf("InheritedRoot = %#v", summary.InheritedRoot)
	}
	if len(summary.OverrideStubs) != 1 || summary.OverrideStubs[0].Name != "Body" || summary.OverrideStubs[0].Depth != (analysis.OptionalDepth{Value: 2, Known: true}) {
		t.Fatalf("OverrideStubs = %#v", summary.OverrideStubs)
	}
	if len(summary.Nodes) != 1 || summary.Nodes[0].Name != "Hat" || len(summary.Mounts) != 1 || summary.Mounts[0].Name != "Weapon" {
		t.Fatalf("Nodes/Mounts = %#v/%#v", summary.Nodes, summary.Mounts)
	}
}

func TestBuildLocalSummaryCountsOnlyLiteralSupportedTypesAndShadows(t *testing.T) {
	t.Parallel()

	summary := parseSummary(t, `[gd_scene format=3]
[node name="Mesh" type="MeshInstance3D"]
[node name="Sun" type="DirectionalLight3D" parent="."]
shadow_enabled = true
[node name="Bulb" type="OmniLight3D" parent="."]
shadow_enabled = false
[node name="Spot" type="SpotLight3D" parent="."]
[node name="CanvasLight" type="PointLight2D" parent="."]
[node name="Custom" type="CustomLight" parent="."]
`)

	want := metrics.Values{Nodes: 6, TreeDepth: 2, MeshInstances: 1, Lights: 3, ShadowLights: 1}
	if summary.Metrics != want {
		t.Fatalf("Metrics = %#v, want %#v", summary.Metrics, want)
	}
}

func TestBuildLocalSummaryOwnsDeterministicResourceAndNodeEvidence(t *testing.T) {
	t.Parallel()

	shadow := true
	reference := &tscn.ResourceRef{Kind: tscn.ResourceRefExternal, ID: "a"}
	document := &tscn.Document{
		ExtResources: map[string]tscn.ExtResource{
			"z": {ID: "z", Type: "Texture2D", Path: "res://shared.res", Position: tscn.Position{Line: 3, Column: 1}},
			"a": {ID: "a", Type: "PackedScene", UID: "uid://child", Path: "res://shared.res", Position: tscn.Position{Line: 2, Column: 1}},
		},
		Nodes: []tscn.Node{
			{Name: "Root", Type: "OmniLight3D", ShadowEnabled: &shadow, Position: tscn.Position{Line: 4, Column: 1}},
			{Name: "Child", Parent: ".", Instance: reference, Position: tscn.Position{Line: 5, Column: 1}},
		},
	}

	first, err := analysis.BuildLocalSummary(document)
	if err != nil {
		t.Fatal(err)
	}
	second, err := analysis.BuildLocalSummary(document)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated summaries differ:\nfirst: %#v\nsecond: %#v", first, second)
	}
	if got := []string{first.ExternalResources[0].ID, first.ExternalResources[1].ID}; !reflect.DeepEqual(got, []string{"a", "z"}) {
		t.Fatalf("resource order = %#v", got)
	}
	if first.ExternalResources[0].Path != first.ExternalResources[1].Path {
		t.Fatal("duplicate raw targets were unexpectedly collapsed")
	}
	if first.Metrics.ExternalResources != 0 || first.Metrics.SceneDependencies != 0 {
		t.Fatalf("premature unique metrics = %#v", first.Metrics)
	}

	shadow = false
	reference.ID = "z"
	document.Nodes[0].Name = "Mutated"
	document.ExtResources["a"] = tscn.ExtResource{ID: "a", Type: "Script", Path: "res://mutated.gd"}

	if first.Nodes[0].Name != "Root" || first.Nodes[0].ShadowEnabled == nil || !*first.Nodes[0].ShadowEnabled {
		t.Fatalf("owned node evidence changed: %#v", first.Nodes[0])
	}
	if first.Mounts[0].Reference.ID != "a" || first.Mounts[0].Candidate == nil || first.Mounts[0].Candidate.Type != "PackedScene" {
		t.Fatalf("owned mount evidence changed: %#v", first.Mounts[0])
	}
	if first.ExternalResources[0].Type != "PackedScene" || first.ExternalResources[0].Path != "res://shared.res" {
		t.Fatalf("owned resource evidence changed: %#v", first.ExternalResources[0])
	}
}

func TestBuildLocalSummaryRequiresParsedRoot(t *testing.T) {
	t.Parallel()

	for _, document := range []*tscn.Document{nil, {}} {
		if _, err := analysis.BuildLocalSummary(document); err == nil {
			t.Fatal("BuildLocalSummary() error = nil, want missing-root error")
		}
	}
}

func parseSummary(t *testing.T, input string) analysis.LocalSummary {
	t.Helper()

	document, err := tscn.Parse(strings.NewReader(input), "local.tscn")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	summary, err := analysis.BuildLocalSummary(document)
	if err != nil {
		t.Fatalf("BuildLocalSummary() error = %v", err)
	}

	return summary
}

func findOrdinaryNode(t *testing.T, summary analysis.LocalSummary, name string) analysis.OrdinaryNode {
	t.Helper()

	for _, node := range summary.Nodes {
		if node.Name == name {
			return node
		}
	}
	t.Fatalf("ordinary node %q not found in %#v", name, summary.Nodes)

	return analysis.OrdinaryNode{}
}
