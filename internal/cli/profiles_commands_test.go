package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/cli"
	"github.com/stfulldev/deadweight.gdt/internal/config"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestProfilesCommandsTranslateRequestsAndSelectPresentation(t *testing.T) {
	t.Parallel()

	listResult, showResult := profileCommandResults(t)
	tests := []struct {
		name     string
		args     []string
		service  *fakeApplication
		wantKind string
		contains string
	}{
		{
			name: "list text",
			args: []string{"--project", "/game", "--config", "/game/.deadweight.gdt.json", "profiles"},
			service: &fakeApplication{listProfiles: func(request application.ProfileRequest) (application.ProfileListResult, error) {
				if request != (application.ProfileRequest{Project: "/game", Config: "/game/.deadweight.gdt.json"}) {
					t.Fatalf("list request = %#v", request)
				}
				return listResult, nil
			}},
			contains: "Custom profiles",
		},
		{
			name: "list json",
			args: []string{"profiles", "--format", "json"},
			service: &fakeApplication{listProfiles: func(request application.ProfileRequest) (application.ProfileListResult, error) {
				if request != (application.ProfileRequest{}) {
					t.Fatalf("list request = %#v", request)
				}
				return listResult, nil
			}},
			wantKind: "profiles",
		},
		{
			name: "show text",
			args: []string{"--project", "/game", "profiles", "show", "shipping"},
			service: &fakeApplication{showProfile: func(request application.ProfileShowRequest) (application.ProfileShowResult, error) {
				want := application.ProfileShowRequest{ProfileRequest: application.ProfileRequest{Project: "/game"}, ID: "shipping"}
				if request != want {
					t.Fatalf("show request = %#v, want %#v", request, want)
				}
				return showResult, nil
			}},
			contains: "Profile:       shipping",
		},
		{
			name: "show json",
			args: []string{"profiles", "show", "shipping", "--format", "json"},
			service: &fakeApplication{showProfile: func(request application.ProfileShowRequest) (application.ProfileShowResult, error) {
				return showResult, nil
			}},
			wantKind: "profile",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			exit := cli.ExecuteWithApplication(test.args, &stdout, &stderr, cli.BuildInfo{Version: "test"}, test.service)
			if exit != 0 || stderr.Len() != 0 || test.service.calls != 1 {
				t.Fatalf("exit/stdout/stderr/calls = %d / %q / %q / %d", exit, stdout.String(), stderr.String(), test.service.calls)
			}
			if test.contains != "" && !strings.Contains(stdout.String(), test.contains) {
				t.Fatalf("stdout = %q, want %q", stdout.String(), test.contains)
			}
			if test.wantKind != "" {
				var document struct {
					Kind string `json:"kind"`
				}
				if err := json.Unmarshal(stdout.Bytes(), &document); err != nil || document.Kind != test.wantKind {
					t.Fatalf("JSON kind/error = %q / %v: %s", document.Kind, err, stdout.String())
				}
			}
		})
	}
}

func TestProfilesCommandsRejectUsageBeforeApplication(t *testing.T) {
	t.Parallel()

	tests := [][]string{
		{"profiles", "extra"},
		{"profiles", "--format", "yaml"},
		{"profiles", "show"},
		{"profiles", "show", "a", "b"},
		{"profiles", "show", "shipping", "--format", "yaml"},
	}
	for _, args := range tests {
		args := append([]string(nil), args...)
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			t.Parallel()
			service := &fakeApplication{}
			var stdout, stderr bytes.Buffer
			exit := cli.ExecuteWithApplication(args, &stdout, &stderr, cli.BuildInfo{}, service)
			if exit != 2 || stdout.Len() != 0 || stderr.Len() == 0 || service.calls != 0 {
				t.Fatalf("exit/stdout/stderr/calls = %d / %q / %q / %d", exit, stdout.String(), stderr.String(), service.calls)
			}
		})
	}
}

func TestProfilesJSONFailureUsesVersionedErrorDocument(t *testing.T) {
	t.Parallel()

	service := &fakeApplication{showProfile: func(application.ProfileShowRequest) (application.ProfileShowResult, error) {
		return application.ProfileShowResult{}, errors.New("profile graph failed")
	}}
	var stdout, stderr bytes.Buffer
	exit := cli.ExecuteWithApplication(
		[]string{"profiles", "show", "shipping", "--format", "json"},
		&stdout,
		&stderr,
		cli.BuildInfo{Version: "test"},
		service,
	)
	if exit != 2 || stdout.Len() != 0 || service.calls != 1 {
		t.Fatalf("exit/stdout/calls = %d / %q / %d", exit, stdout.String(), service.calls)
	}
	var document struct {
		Kind  string `json:"kind"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &document); err != nil || document.Kind != "error" || document.Error.Message != "profile graph failed" {
		t.Fatalf("error document = %#v, err %v: %s", document, err, stderr.String())
	}
}

func profileCommandResults(t *testing.T) (application.ProfileListResult, application.ProfileShowResult) {
	t.Helper()
	configuration, err := config.Decode(strings.NewReader(`{
  "version": 1,
  "budgets": {"lights": 7},
  "profiles": {"shipping": {"extends": "steam-deck", "name": "Shipping", "budgets": {"nodes": 100}}}
}`), "/game/.deadweight.gdt.json")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	profiles, err := policy.ListProfiles("/game/.deadweight.gdt.json", configuration)
	if err != nil {
		t.Fatalf("policy.ListProfiles() error = %v", err)
	}
	explanation, err := policy.ExplainProfile("/game/.deadweight.gdt.json", configuration, "shipping")
	if err != nil {
		t.Fatalf("policy.ExplainProfile() error = %v", err)
	}
	root := project.Root{Directory: "/game", ProjectFile: "/game/project.godot"}
	source := config.Source{Path: "/game/.deadweight.gdt.json"}
	list := application.ProfileListResult{Project: root, ConfigSource: source, Profiles: profiles}
	show := application.ProfileShowResult{Project: root, ConfigSource: source, Explanation: explanation}
	if !reflect.DeepEqual(list.Project, show.Project) {
		t.Fatal("profile fixtures use different project contexts")
	}
	return list, show
}
