package report

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestProfileReportGoldens(t *testing.T) {
	t.Parallel()

	list, show := profileReportResults(t, "/game", "/game/.deadweight.gdt.json")
	tests := []struct {
		name   string
		json   bool
		render func() (string, error)
	}{
		{name: "profiles_list", render: func() (string, error) { return ProfileList(list, Options{}) }},
		{name: "profiles_show", render: func() (string, error) { return ProfileShow(show, Options{}) }},
		{name: "profiles_list", json: true, render: func() (string, error) {
			return ProfileListJSON(list, Options{Version: "0.2.0-test", Color: true})
		}},
		{name: "profiles_show", json: true, render: func() (string, error) {
			return ProfileShowJSON(show, Options{Version: "0.2.0-test", Color: true})
		}},
	}
	for _, test := range tests {
		test := test
		label := test.name
		if test.json {
			label = "json_" + label
		}
		t.Run(label, func(t *testing.T) {
			t.Parallel()
			got, err := test.render()
			if err != nil {
				t.Fatalf("render error = %v", err)
			}
			if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") || strings.Contains(got, "\x1b[") {
				t.Fatalf("invalid framing: %q", got)
			}
			directory := filepath.Join("testdata", "golden")
			if test.json {
				directory = filepath.Join(directory, "json")
				validateReportDocument(t, []byte(got))
			}
			goldenPath := filepath.Join(directory, test.name+".golden")
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

func TestProfileReportsArePortableDeterministicAndImmutable(t *testing.T) {
	t.Parallel()

	leftList, leftShow := profileReportResults(t, "/tmp/checkout-one", "/tmp/checkout-one/.deadweight.gdt.json")
	rightList, rightShow := profileReportResults(t, `D:\work\checkout-two`, `D:\work\checkout-two\.deadweight.gdt.json`)
	before := leftShow.Explanation.Clone()

	leftListJSON, err := ProfileListJSON(leftList, Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("ProfileListJSON(left) error = %v", err)
	}
	rightListJSON, err := ProfileListJSON(rightList, Options{Version: "test"})
	if err != nil {
		t.Fatalf("ProfileListJSON(right) error = %v", err)
	}
	leftShowJSON, err := ProfileShowJSON(leftShow, Options{Version: "test", Color: true})
	if err != nil {
		t.Fatalf("ProfileShowJSON(left) error = %v", err)
	}
	rightShowJSON, err := ProfileShowJSON(rightShow, Options{Version: "test"})
	if err != nil {
		t.Fatalf("ProfileShowJSON(right) error = %v", err)
	}
	if leftListJSON != rightListJSON || leftShowJSON != rightShowJSON {
		t.Fatalf("portable profile JSON differs\nLIST LEFT\n%sLIST RIGHT\n%sSHOW LEFT\n%sSHOW RIGHT\n%s", leftListJSON, rightListJSON, leftShowJSON, rightShowJSON)
	}
	for _, forbidden := range []string{"checkout-one", "checkout-two", `\`, "\x1b["} {
		if strings.Contains(leftListJSON+leftShowJSON, forbidden) {
			t.Fatalf("profile JSON contains %q", forbidden)
		}
	}
	repeated, err := ProfileShowJSON(leftShow, Options{Version: "test"})
	if err != nil || repeated != leftShowJSON || !reflect.DeepEqual(leftShow.Explanation, before) {
		t.Fatalf("repeated render/error/mutation = %v / %v / %v", repeated == leftShowJSON, err, reflect.DeepEqual(leftShow.Explanation, before))
	}
}

func TestProfileListTextHasExplicitEmptyState(t *testing.T) {
	t.Parallel()

	result := application.ProfileListResult{
		Project:      project.Root{Directory: "/game", ProjectFile: "/game/project.godot"},
		ConfigSource: config.Source{Path: "/game/.deadweight.gdt.json"},
		Profiles:     []policy.ProfileSummary{},
	}
	got, err := ProfileList(result, Options{})
	if err != nil || !strings.Contains(got, "No custom profiles are declared.") {
		t.Fatalf("ProfileList(empty) = %q / %v", got, err)
	}
}

func profileReportResults(t *testing.T, projectRoot, configPath string) (application.ProfileListResult, application.ProfileShowResult) {
	t.Helper()
	configuration, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "budgets": {"lights": 7},
  "profiles": {
    "shipping": {
      "extends": "steam-deck",
      "name": "Shipping",
      "quality": "low",
      "budgets": {"nodes": 100}
    },
    "ci": {"description": "Minimal CI policy", "budgets": {"nodes": 1}}
  }
}`), configPath)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	profiles, err := policy.ListProfiles(configPath, configuration)
	if err != nil {
		t.Fatalf("policy.ListProfiles() error = %v", err)
	}
	explanation, err := policy.ExplainProfile(configPath, configuration, "shipping")
	if err != nil {
		t.Fatalf("policy.ExplainProfile() error = %v", err)
	}
	root := project.Root{Directory: projectRoot, ProjectFile: filepath.Join(projectRoot, "project.godot")}
	source := config.Source{Path: configPath}
	return application.ProfileListResult{Project: root, ConfigSource: source, Profiles: profiles}, application.ProfileShowResult{
		Project: root, ConfigSource: source, Explanation: explanation,
	}
}
