package tscn_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func TestParseExtractsSupportedSceneSubset(t *testing.T) {
	t.Parallel()

	input := `[gd_scene load_steps=4 format=3 uid="uid://city"] ; header comment

[ext_resource type="PackedScene" uid="uid://lamp" path="../props/lamp;night.tscn" id="1_lamp"]
[ext_resource type="Script" path="res://scripts/city.gd" id=2]

[sub_resource type="ArrayMesh" id="Mesh_1"]
_surfaces = [{
"aabb": AABB(0, 0, 0, 1, 1, 1),
"data": PackedByteArray(1, 2, 3)
}]

[node name="City \"Center\"" type="Node3D" groups=[&"level"]]
metadata = {
  "literal": "[not a section]",
  "enabled": true
}

[node name="Sun" type="DirectionalLight3D" parent="." index="0" node_paths=PackedStringArray("follow_node")]
shadow_enabled = true ; property comment

[node name="LampA" parent="." instance=ExtResource("1_lamp")]
[node name="Embedded" parent="." instance=SubResource(7)]
[node name="NoShadow" type="OmniLight3D" parent="."]
shadow_enabled = false
[node name="Placeholder" parent="." owner="." instance_placeholder="res://deferred.tscn"]

[connection signal="ready" from="." to="." method="_on_ready" binds=[1, {"nested": true}]]
`

	document, err := tscn.Parse(strings.NewReader(input), "city.tscn")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if document.Header.Format != 3 || document.Header.UID != "uid://city" {
		t.Errorf("Header = %#v", document.Header)
	}
	if len(document.ExtResources) != 2 {
		t.Fatalf("len(ExtResources) = %d, want 2", len(document.ExtResources))
	}
	if resource := document.ExtResources["1_lamp"]; resource.Type != "PackedScene" || resource.Path != "../props/lamp;night.tscn" || resource.UID != "uid://lamp" {
		t.Errorf("PackedScene resource = %#v", resource)
	}
	if resource := document.ExtResources["2"]; resource.Type != "Script" || resource.Path != "res://scripts/city.gd" {
		t.Errorf("numeric-ID resource = %#v", resource)
	}

	if len(document.Nodes) != 6 {
		t.Fatalf("len(Nodes) = %d, want 6", len(document.Nodes))
	}
	if root := document.Nodes[0]; root.Name != `City "Center"` || root.Type != "Node3D" || root.Parent != "" {
		t.Errorf("root = %#v", root)
	}
	if sun := document.Nodes[1]; sun.ShadowEnabled == nil || !*sun.ShadowEnabled || sun.Index == nil || *sun.Index != 0 {
		t.Errorf("sun = %#v", sun)
	}
	if lamp := document.Nodes[2]; lamp.Instance == nil || lamp.Instance.Kind != tscn.ResourceRefExternal || lamp.Instance.ID != "1_lamp" {
		t.Errorf("lamp = %#v", lamp)
	}
	if embedded := document.Nodes[3]; embedded.Instance == nil || embedded.Instance.Kind != tscn.ResourceRefInternal || embedded.Instance.ID != "7" {
		t.Errorf("embedded = %#v", embedded)
	}
	if noShadow := document.Nodes[4]; noShadow.ShadowEnabled == nil || *noShadow.ShadowEnabled {
		t.Errorf("noShadow = %#v", noShadow)
	}
	if placeholder := document.Nodes[5]; placeholder.InstancePlaceholder != "res://deferred.tscn" || placeholder.Owner != "." {
		t.Errorf("placeholder = %#v", placeholder)
	}
	if document.Features.HasInheritedRoot || document.Features.HasOverrideNodes || document.Features.HasEditable {
		t.Errorf("Features = %#v, want all false", document.Features)
	}
}

func TestParseDetectsInheritanceOverridesAndEditableSections(t *testing.T) {
	t.Parallel()

	input := `[gd_scene format=3]

[ext_resource type="PackedScene" path="res://enemy.tscn" id="1_base"]

[node name="Zombie" instance=ExtResource("1_base")]
[node name="Body" parent="."]
visible = false

[editable path="Body"]
`

	document, err := tscn.Parse(strings.NewReader(input), "zombie.tscn")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := tscn.Features{
		HasInheritedRoot: true,
		HasOverrideNodes: true,
		HasEditable:      true,
	}
	if document.Features != want {
		t.Fatalf("Features = %#v, want %#v", document.Features, want)
	}
}

func TestParseLeavesAbsentShadowPropertyUnset(t *testing.T) {
	t.Parallel()

	document, err := tscn.Parse(strings.NewReader("[gd_scene format=3]\n[node name=\"Light\" type=\"OmniLight3D\"]\n"), "light.tscn")
	if err != nil {
		t.Fatal(err)
	}
	if document.Nodes[0].ShadowEnabled != nil {
		t.Fatalf("ShadowEnabled = %v, want nil", *document.Nodes[0].ShadowEnabled)
	}
}

func TestParseRejectsMalformedOrUnsupportedScenes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: "missing [gd_scene]"},
		{name: "wrong first section", input: `[ext_resource id="1"]`, want: "first section must be [gd_scene]"},
		{name: "missing format", input: "[gd_scene]\n[node name=\"Root\" type=\"Node\"]", want: "must define format=3"},
		{name: "format two", input: "[gd_scene format=2]\n[node name=\"Root\" type=\"Node\"]", want: "unsupported Godot scene format 2"},
		{name: "duplicate scene", input: "[gd_scene format=3]\n[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]", want: "duplicate [gd_scene]"},
		{name: "no root", input: "[gd_scene format=3]\n", want: "scene must contain exactly one root node"},
		{name: "duplicate resource", input: "[gd_scene format=3]\n[ext_resource id=\"1\"]\n[ext_resource id=1]\n[node name=\"Root\" type=\"Node\"]", want: "duplicate external resource id"},
		{name: "root parent", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node\" parent=\".\"]", want: "root node must not define parent"},
		{name: "child without parent", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\n[node name=\"Child\" type=\"Node\"]", want: "non-root node"},
		{name: "bad instance", input: "[gd_scene format=3]\n[node name=\"Root\" instance=Other(1)]", want: "must be ExtResource"},
		{name: "two instance forms", input: "[gd_scene format=3]\n[node name=\"Root\" instance=ExtResource(1) instance_placeholder=\"res://x.tscn\"]", want: "both instance and instance_placeholder"},
		{name: "unclosed string", input: "[gd_scene format=3]\n[node name=\"Root]", want: "unterminated string"},
		{name: "unclosed value", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\nvalue = [1, 2\n", want: "unclosed delimiter"},
		{name: "mismatched value", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\nvalue = [1, 2)\n", want: "mismatched closing delimiter"},
		{name: "bad shadow", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"OmniLight3D\"]\nshadow_enabled = 1\n", want: "must be true or false"},
		{name: "duplicate shadow", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"OmniLight3D\"]\nshadow_enabled = true\nshadow_enabled = false\n", want: "duplicate shadow_enabled"},
		{name: "missing property value", input: "[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\nvalue =\n", want: "has no value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := tscn.Parse(strings.NewReader(test.input), "bad.tscn")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want containing %q", err, test.want)
			}

			var parseError *tscn.ParseError
			if !errors.As(err, &parseError) {
				t.Fatalf("error type = %T, want *tscn.ParseError", err)
			}
			if parseError.Code != "SB2001" || parseError.Position.Line < 1 || parseError.Position.Column < 1 {
				t.Fatalf("ParseError = %#v", parseError)
			}
		})
	}
}

func TestParseSkipsLargePackedArrayWithoutBuildingVariantAST(t *testing.T) {
	t.Parallel()

	var input strings.Builder
	input.WriteString("[gd_scene format=3]\n[sub_resource type=\"ArrayMesh\" id=\"Mesh\"]\n_data = PackedByteArray(")
	for index := 0; index < 50_000; index++ {
		if index > 0 {
			input.WriteByte(',')
		}
		fmt.Fprintf(&input, "%d", index%256)
	}
	input.WriteString(")\n[node name=\"Root\" type=\"Node3D\"]\n")

	document, err := tscn.Parse(strings.NewReader(input.String()), "large.tscn")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(document.Nodes) != 1 || document.Nodes[0].Name != "Root" {
		t.Fatalf("Nodes = %#v", document.Nodes)
	}
}
