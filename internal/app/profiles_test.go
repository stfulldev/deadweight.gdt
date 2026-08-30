package app

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestProfileFlowsUseOnlyProjectConfigurationAndPolicyDependencies(t *testing.T) {
	t.Parallel()

	limit := int64(12)
	configuration := config.Config{Version: config.CurrentVersion, Profiles: map[string]config.Profile{"shipping": {}}}
	source := config.Source{Path: "/game/custom.json", Explicit: true}
	root := project.Root{Directory: "/game", ProjectFile: "/game/project.godot"}
	events := make([]string, 0, 8)
	returnedProfiles := []policy.ProfileSummary{{ID: "shipping", Name: "Shipping"}}
	returnedExplanation := policy.Explanation{
		Effective: policy.Effective{
			Kind:    policy.KindProfile,
			ID:      "shipping",
			Budgets: budget.Limits{Nodes: &limit},
		},
		Chain: []policy.Layer{{Kind: policy.LayerProfile, ID: "shipping"}},
		BudgetSources: policy.LimitSources{
			Nodes: &policy.Layer{Kind: policy.LayerProfile, ID: "shipping"},
		},
	}
	application := New(Dependencies{
		WorkingDirectory: func() (string, error) {
			events = append(events, "cwd")
			return "/work", nil
		},
		FindProjectContext: func(request project.ContextRequest) (project.Root, error) {
			events = append(events, "project")
			if request != (project.ContextRequest{WorkingDirectory: "/work", ExplicitProject: "/game"}) {
				t.Fatalf("context request = %#v", request)
			}
			return root, nil
		},
		LoadConfig: func(projectRoot, explicitPath string) (config.Config, config.Source, bool, error) {
			events = append(events, "config")
			if projectRoot != "/game" || explicitPath != "/game/custom.json" {
				t.Fatalf("LoadConfig(%q, %q)", projectRoot, explicitPath)
			}
			return configuration, source, true, nil
		},
		ListProfiles: func(gotSource string, got config.Config) ([]policy.ProfileSummary, error) {
			events = append(events, "list")
			if gotSource != source.Path || !reflect.DeepEqual(got, configuration) {
				t.Fatalf("ListProfiles(%q, %#v)", gotSource, got)
			}
			return returnedProfiles, nil
		},
		ExplainProfile: func(gotSource string, got config.Config, id string) (policy.Explanation, error) {
			events = append(events, "show")
			if gotSource != source.Path || id != "shipping" || !reflect.DeepEqual(got, configuration) {
				t.Fatalf("ExplainProfile(%q, %#v, %q)", gotSource, got, id)
			}
			return returnedExplanation, nil
		},
	})

	request := ProfileRequest{Project: "/game", Config: "/game/custom.json"}
	listed, err := application.ListProfiles(request)
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	shown, err := application.ShowProfile(ProfileShowRequest{ProfileRequest: request, ID: "shipping"})
	if err != nil {
		t.Fatalf("ShowProfile() error = %v", err)
	}
	if !reflect.DeepEqual(events, []string{"cwd", "project", "config", "list", "cwd", "project", "config", "show"}) {
		t.Fatalf("events = %v", events)
	}
	if listed.Project != root || listed.ConfigSource != source || listed.Profiles[0].ID != "shipping" {
		t.Fatalf("listed result = %#v", listed)
	}
	if shown.Project != root || shown.ConfigSource != source || shown.Explanation.Effective.ID != "shipping" {
		t.Fatalf("shown result = %#v", shown)
	}
	listed.Profiles[0].ID = "changed"
	*shown.Explanation.Effective.Budgets.Nodes = 99
	if returnedProfiles[0].ID != "shipping" || *returnedExplanation.Effective.Budgets.Nodes != 12 {
		t.Fatal("profile application results alias dependency-owned data")
	}
}

func TestProfileFlowsRejectMissingConfigurationBeforePolicyResolution(t *testing.T) {
	t.Parallel()

	policyCalls := 0
	application := New(Dependencies{
		WorkingDirectory: func() (string, error) { return "/game", nil },
		FindProjectContext: func(project.ContextRequest) (project.Root, error) {
			return project.Root{Directory: "/game", ProjectFile: "/game/project.godot"}, nil
		},
		LoadConfig: func(string, string) (config.Config, config.Source, bool, error) {
			return config.Config{}, config.Source{}, false, nil
		},
		ListProfiles: func(string, config.Config) ([]policy.ProfileSummary, error) {
			policyCalls++
			return nil, errors.New("must not resolve")
		},
	})

	result, err := application.ListProfiles(ProfileRequest{})
	if err == nil || !strings.Contains(err.Error(), config.DefaultFilename) || !strings.Contains(err.Error(), "--config") {
		t.Fatalf("ListProfiles() result/error = %#v / %v", result, err)
	}
	if policyCalls != 0 || !reflect.DeepEqual(result, ProfileListResult{}) {
		t.Fatalf("policy calls/result = %d / %#v", policyCalls, result)
	}
}
