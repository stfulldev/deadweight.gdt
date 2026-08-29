package report

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestTopContributorsUsePortableStableOrderingWithoutMutation(t *testing.T) {
	t.Parallel()

	result := completeInspect()
	result.Analysis.Summary.Contributions = []analysis.SceneContribution{
		contributionFixture("/project/b.tscn", "res://b.tscn", 10, analysis.ReliabilityExact),
		contributionFixture("/project/a.tscn", "res://a.tscn", 10, analysis.ReliabilityLowerBound),
		contributionFixture("/project/c.tscn", "res://c.tscn", 3, analysis.ReliabilityApproximate),
	}
	before := snapshotJSON(t, result.Analysis.Summary.Contributions)
	selection := ContributionSelection{Metric: metrics.Nodes, Limit: 2}
	first, err := topContributions(result, selection)
	if err != nil {
		t.Fatalf("topContributions() error = %v", err)
	}
	second, err := topContributions(result, selection)
	if err != nil {
		t.Fatalf("second topContributions() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) || len(first) != 2 || first[0].Scene != "res://a.tscn" || first[1].Scene != "res://b.tscn" {
		t.Fatalf("top rows = %#v / %#v", first, second)
	}
	if after := snapshotJSON(t, result.Analysis.Summary.Contributions); !bytes.Equal(before, after) {
		t.Fatal("top projection mutated contribution rows")
	}

	rendered, err := Inspect(result, Options{Version: "test", Contributions: selection})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	firstPosition := strings.Index(rendered, "res://a.tscn")
	secondPosition := strings.Index(rendered, "res://b.tscn")
	if firstPosition < 0 || secondPosition <= firstPosition || !strings.Contains(rendered, "10+") || strings.Contains(rendered, "res://c.tscn") {
		t.Fatalf("top text ordering/limit = %s", rendered)
	}
}

func TestTopDepthKeepsUnavailableCandidatesAfterKnownRows(t *testing.T) {
	t.Parallel()

	result := completeInspect()
	known := contributionFixture("/project/known.tscn", "res://known.tscn", 1, analysis.ReliabilityExact)
	known.DepthCandidate = analysis.OptionalDepth{Value: 5, Known: true}
	unknown := contributionFixture("/project/unknown.tscn", "res://unknown.tscn", 100, analysis.ReliabilityLowerBound)
	unknown.DepthCandidate = analysis.OptionalDepth{}
	result.Analysis.Summary.Contributions = []analysis.SceneContribution{unknown, known}

	rows, err := topContributions(result, ContributionSelection{Metric: metrics.TreeDepth, Limit: 10})
	if err != nil {
		t.Fatalf("topContributions() error = %v", err)
	}
	if len(rows) != 2 || rows[0].Scene != "res://known.tscn" || !rows[0].ValueKnown || rows[1].Scene != "res://unknown.tscn" || rows[1].ValueKnown {
		t.Fatalf("depth rows = %#v", rows)
	}
}

func TestJSONContributionsRetainSharedEvidenceAndTopProjection(t *testing.T) {
	t.Parallel()

	result := sharedEvidenceInspect()

	rendered, err := InspectJSON(result, Options{
		Version:       "test",
		Contributions: ContributionSelection{Metric: metrics.Nodes, Limit: 2},
	})
	if err != nil {
		t.Fatalf("InspectJSON() error = %v", err)
	}
	validateReportDocument(t, []byte(rendered))
	for _, required := range []string{
		`"aggregation": "unique_union"`,
		`"identity": "res://assets/shared.png"`,
		`"top_contributors"`,
		`"limit": 2`,
	} {
		if !strings.Contains(rendered, required) {
			t.Errorf("JSON lacks %q: %s", required, rendered)
		}
	}
	if strings.Count(rendered, `"identity": "res://assets/shared.png"`) != 1 {
		t.Fatalf("shared identity is duplicated: %s", rendered)
	}
	if strings.Contains(rendered, "<PROJECT>") || strings.Contains(rendered, `\`) {
		t.Fatalf("JSON leaks non-portable identity: %s", rendered)
	}
}

func sharedEvidenceInspect() application.InspectResult {
	result := completeInspect()
	result.Analysis.Summary.Contributions = []analysis.SceneContribution{
		contributionFixture("<PROJECT>/levels/city.tscn", "res://levels/city.tscn", 20, analysis.ReliabilityExact),
		contributionFixture("<PROJECT>/actors/a.tscn", "res://actors/a.tscn", 10, analysis.ReliabilityExact),
		contributionFixture("<PROJECT>/actors/b.tscn", "res://actors/b.tscn", 5, analysis.ReliabilityExact),
	}
	result.Analysis.UniqueEvidence = []analysis.UniqueEvidence{{
		Metric: metrics.ExternalResources, Canonical: "<PROJECT>/assets/shared.png", Display: "res://assets/shared.png",
		Referrers: []analysis.UniqueReferrer{
			{SceneCanonical: "<PROJECT>/actors/b.tscn", SceneDisplay: "res://actors/b.tscn", ResourceID: "2_texture", RawTarget: "res://assets/shared.png", Occurrences: 1},
			{SceneCanonical: "<PROJECT>/actors/a.tscn", SceneDisplay: "res://actors/a.tscn", ResourceID: "1_texture", RawTarget: "res://assets/shared.png", Occurrences: 1},
		},
	}}

	return result
}

func contributionFixture(
	canonical string,
	display string,
	nodes int64,
	reliability analysis.Reliability,
) analysis.SceneContribution {
	return analysis.SceneContribution{
		Kind:             analysis.ContributionScene,
		SceneCanonical:   canonical,
		SceneDisplay:     display,
		DeclaringScene:   "/project/root.tscn",
		DeclaringDisplay: "res://root.tscn",
		MountPath:        "Child",
		Occurrences:      1,
		Values:           analysis.ContributionValues{Nodes: nodes},
		DepthCandidate:   analysis.OptionalDepth{Value: 2, Known: true},
		Reliability:      reliability,
	}
}
