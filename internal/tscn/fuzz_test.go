package tscn_test

import (
	"bytes"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

func FuzzParse(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\n"))
	f.Add([]byte("[gd_scene format=3]\n[node name=\"Root\" type=\"Node\"]\nvalue=[{\"x\": PackedInt32Array(1,2,3)}]\n"))
	f.Add([]byte{0xff, 0x00, '[', 'g', 'd', '_', 's', 'c', 'e', 'n', 'e'})

	f.Fuzz(func(t *testing.T, input []byte) {
		_, _ = tscn.Parse(bytes.NewReader(input), "fuzz.tscn")
	})
}
