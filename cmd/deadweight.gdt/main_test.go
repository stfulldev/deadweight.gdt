package main

import "testing"

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		injected      string
		moduleVersion string
		want          string
	}{
		{name: "explicit linker value wins", injected: "1.2.3-release", moduleVersion: "v9.9.9", want: "1.2.3-release"},
		{name: "explicit leading v is retained", injected: "v1.2.3", moduleVersion: "v9.9.9", want: "v1.2.3"},
		{name: "tagged module", injected: "dev", moduleVersion: "v0.1.1", want: "0.1.1"},
		{name: "pseudo version", injected: "dev", moduleVersion: "v0.1.1-0.20260829153000-abcdef123456", want: "0.1.1-0.20260829153000-abcdef123456"},
		{name: "module without leading v", injected: "dev", moduleVersion: "0.2.0", want: "0.2.0"},
		{name: "only one leading v is removed", injected: "dev", moduleVersion: "vv0.2.0", want: "v0.2.0"},
		{name: "empty linker uses module", injected: "", moduleVersion: "v0.3.0", want: "0.3.0"},
		{name: "empty metadata falls back", injected: "dev", moduleVersion: "", want: "dev"},
		{name: "devel metadata falls back", injected: "dev", moduleVersion: "(devel)", want: "dev"},
		{name: "devel linker uses module", injected: "(devel)", moduleVersion: "v0.4.0", want: "0.4.0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := resolveVersion(test.injected, test.moduleVersion); actual != test.want {
				t.Fatalf("resolveVersion(%q, %q) = %q, want %q", test.injected, test.moduleVersion, actual, test.want)
			}
		})
	}
}
