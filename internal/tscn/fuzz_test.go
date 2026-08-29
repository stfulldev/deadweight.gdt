package tscn_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\n",
		"[gd_scene uid=\"uid://fixture\" load_steps=2 format=3]\n[node type=\"Node\" name=\"Root\"]\n",
		"[gd_scene format=3]\n[ext_resource type=\"PackedScene\" path=\"child.tscn\" id=1]\n[node name=\"Root\" type=\"Node\"]\n[node name=\"Child\" parent=\".\" instance=ExtResource(1)]\n",
		"[gd_scene format=3]\n[node name=\"quote\\\";still-string\" type=\"Node\"] ; trailing comment\n",
		"[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\nvalue = [{\n  \"nested\": Transform3D(1, 0, 0, 0, 1, 0, 0, 0, 1, 2, 3, 4),\n  \"items\": [1, 2, 3]\n}]\n",
		"[gd_scene format=3]\n[unknown_section arbitrary={\"a\": [1, 2, 3]}]\nignored = PackedInt32Array(1, 2, 3)\n[node name=\"Root\" type=\"Node\"]\nunknown = Color(1, 0.5, 0.25, 1)\n",
		"[gd_scene format=3]\n[ext_resource type=\"PackedScene\" path=\"child.tscn\" id=\"child\"]\n[node name=\"Root\" type=\"Node\"]\n[node name=\"Child\" parent=\".\" instance=ExtResource(\"child\")]\n",
		"[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\n[node name=\"On\" type=\"OmniLight3D\" parent=\".\"]\nshadow_enabled = true\n[node name=\"Off\" type=\"SpotLight3D\" parent=\".\"]\nshadow_enabled = false\n[node name=\"Absent\" type=\"DirectionalLight3D\" parent=\".\"]\n",
		"[gd_scene format=3]\n[ext_resource id=\"duplicate\"]\n[ext_resource id=\"duplicate\"]\n[node name=\"Root\" type=\"Node\"]\n",
		"[ext_resource id=\"missing-scene\"]\n",
		"[gd_scene format=2]\n[node name=\"Root\" type=\"Node\"]\n",
		"[gd_scene format=3]\n[node name=\"Root]\nvalue = [1, 2\n",
		"[gd_scene format=3]\n[ext_resource type=\"PackedScene\" path=\"base.tscn\" id=\"base\"]\n[ext_resource type=\"PackedScene\" path=\"child.tscn\" id=\"child\"]\n[node name=\"Inherited\" instance=ExtResource(\"base\")]\n[node name=\"Override\" parent=\".\"]\n[node name=\"Ordinary\" type=\"MeshInstance3D\" parent=\"Override\"]\n[node name=\"Instance\" parent=\"Override\" instance=ExtResource(\"child\")]\n[editable path=\"Override\"]\n",
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}
	f.Add([]byte("[gd_scene format=3]\n[sub_resource type=\"ArrayMesh\" id=\"Mesh\"]\n_data=PackedByteArray(" + strings.Repeat("1,", 4_096) + "0)\n[node name=\"Root\" type=\"Node3D\"]\n"))
	f.Add([]byte{0xff, 0x00, '[', 'g', 'd', '_', 's', 'c', 'e', 'n', 'e'})

	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = tscn.Parse(bytes.NewReader(input), "fuzz.tscn")
	})
}
